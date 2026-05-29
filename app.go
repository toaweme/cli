package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/toaweme/structs"
)

var ErrCommandNotFound = errors.New("command not found")
var ErrNoCommands = errors.New("no commands registered")
var ErrNoArguments = errors.New("no arguments provided")
var ErrDisplaySubCommands = errors.New("print sub commands")
var ErrShowingHelp = errors.New("showing help")
var ErrShowingVersion = errors.New("showing version")

const helpCommand = "help"

type app struct {
	config         Config
	globalOptions  *GlobalOptions
	commands       []Command[any]
	defaultCommand Command[any]
}

func NewApp(config Config, opts GlobalOptions) *app {
	return &app{
		config:        config,
		globalOptions: &opts,
		commands:      make([]Command[any], 0),
	}
}

var _ App = (*app)(nil)

func (c *app) Commands() []Command[any] {
	return c.commands
}

func (c *app) Config() Config {
	return c.config
}

// isRegisteredFormat reports whether value is the name of an output codec registered
// in Config.Formats, matched against each codec's extension ("yaml" vs ".yaml").
func (c *app) isRegisteredFormat(value any) bool {
	name, ok := value.(string)
	if !ok || name == "" {
		return false
	}
	for _, codec := range c.config.Formats {
		if strings.TrimPrefix(codec.Extension(), ".") == name {
			return true
		}
	}
	return false
}

func (c *app) Default(cmd Command[any]) Command[any] {
	c.defaultCommand = cmd

	return cmd
}

func (c *app) Add(name string, cmd Command[any]) Command[any] {
	cmd.Name(name)
	c.commands = append(c.commands, cmd)

	return cmd
}

func (c *app) Help(cmd Command[any]) Command[any] {
	return c.Add(helpCommand, cmd)
}

func (c *app) Run(osArgs []string) error {
	if len(c.commands) < 1 {
		return ErrNoCommands
	}

	// hand every command (and subcommand) the app Config so it can read global
	// configuration via BaseCommand.Config()/Store(); ordering-independent, unlike
	// binding at registration time.
	c.bindConfigTree()

	if len(osArgs) > 0 && osArgs[0] == "__complete" {
		c.handleComplete(osArgs[1:])
		return nil
	}

	if len(osArgs) < 1 {
		if c.defaultCommand == nil {
			err := c.runHelp(nil)
			if err != nil {
				return fmt.Errorf("failed to run help: %w", err)
			}
			return ErrShowingHelp
		}

		// the default command takes no CLI flags, only config + env
		if err := c.loadCommandConfig(c.defaultCommand, nil); err != nil {
			return err
		}

		err := c.defaultCommand.Run(*c.globalOptions, Unknowns{
			Args:    []string{},
			Options: map[string]any{},
		})
		if err != nil {
			return fmt.Errorf("failed to run default command: %w", err)
		}

		return nil
	}

	globalOptions, globalUnknownOpts := c.getGlobalOptions(osArgs)

	// a --format value naming a registered output codec is valid even though the
	// static oneof rule on GlobalOptions.Format only lists the built-in formats.
	var skipValidate []string
	if c.isRegisteredFormat(globalOptions["format"]) {
		skipValidate = append(skipValidate, "format")
	}

	err := mapStructToOptions(c.globalOptions, globalOptions, skipValidate...)
	if err != nil {
		return fmt.Errorf("failed to update global options struct: %w", err)
	}

	if c.globalOptions.Version {
		c.printVersion()
		return ErrShowingVersion
	}

	// commandArgs holds the osArgs that are commands
	// allArgs holds the osArgs that are not commands
	command, commandArgs, allArgs, err := c.matchCommandByArgs(osArgs)
	if err != nil {
		if errors.Is(err, ErrCommandNotFound) && c.globalOptions.Help {
			helpErr := c.runHelp(commandArgs, globalUnknownOpts)
			if helpErr != nil {
				return fmt.Errorf("failed to run help: %w", helpErr)
			}

			return fmt.Errorf("%w: %w", err, ErrShowingHelp)
		}

		helpErr := c.runHelp(commandArgs, globalUnknownOpts)
		if helpErr != nil {
			return fmt.Errorf("failed to run help: %w", helpErr)
		}
		return fmt.Errorf("%w: %w", err, ErrShowingHelp)
	}

	commandInputs := command.Options()
	commandFields, err := structs.GetStructFields(commandInputs, nil, structs.DefaultEncodingTags)
	if err != nil {
		return fmt.Errorf("failed to get struct fields: %w", err)
	}

	// cmdArgs are the args defined as numeric tags in the struct e.g. `arg:"0"`
	// cmdUnknownArgs are the args that are not defined in the struct
	// commandOptions are the options defined in the struct e.g. `arg:"cwd"`
	// cmdUnknownOptions are the options that are not defined in the struct
	cmdArgs, cmdUnknownArgs, commandOptions, cmdUnknownOptions := getCommandArgs(allArgs, commandFields)
	unknowns := Unknowns{
		Args:    cmdUnknownArgs,
		Options: cmdUnknownOptions,
	}

	// commandOptions holds the parsed flags; fold in positionals keyed by index
	// so the two together form the highest-precedence flags layer.
	flags := commandOptions
	for i, arg := range cmdArgs {
		flags[fmt.Sprintf("%d", i)] = arg
	}

	// if --help is passed, show help
	if c.globalOptions.Help {
		err := c.runHelp(commandArgs, globalUnknownOpts)
		if err != nil {
			return fmt.Errorf("failed to run help: %w", err)
		}

		return ErrShowingHelp
	}

	if err := c.loadCommandConfig(command, flags); err != nil {
		return err
	}

	err = command.Run(*c.globalOptions, unknowns)
	if err != nil {
		if errors.Is(err, ErrDisplaySubCommands) {
			return c.runHelp(commandArgs, globalUnknownOpts)
		}
		return fmt.Errorf("failed to run command %q: %w", command.Name(""), err)
	}

	return nil
}

// loadCommandConfig populates command.Options() according to the resolved merge
// strategy, then validates the merged result. flags are the explicit CLI inputs
// (parsed flags plus positionals keyed by index); pass nil when the command takes
// none.
//
// MergeLayered (with a Store) layers defaults -> config store(s) -> env -> flags,
// so a shared section in the config file (e.g. a `database:` block) fills any
// field tagged to match it, while env and flags override per field. MergeEnvFlags
// (the default, and the fallback when MergeLayered is requested without a Store)
// applies defaults -> env -> flags only. Validation runs after the merge so
// `required` is satisfied by config- or default-provided values, not just flags.
func (c *app) loadCommandConfig(command Command[any], flags map[string]any) error {
	inputs := command.Options()
	cmdStrategy, mapping := command.ConfigStrategy()

	if c.resolveStrategy(cmdStrategy) == MergeLayered && c.config.Store != nil {
		// default layout: shared top-level config (the plain tag match inside
		// Load) plus the command's own "<name>:" section overriding it. A command
		// that declares its own mapping opts out of the name-namespace default.
		if mapping == nil {
			if name := command.Name(""); name != "" {
				mapping = Namespaced(name)
			}
		}
		if err := c.config.Store.Load(inputs, LoadOptions{Env: true, Flags: flags, Mapping: mapping}); err != nil {
			return fmt.Errorf("failed to load config for command %q: %w", command.Name(""), err)
		}
	} else if err := mergeConfig(inputs, nil, "", true, flags, nil); err != nil {
		return fmt.Errorf("failed to merge config for command %q: %w", command.Name(""), err)
	}

	// validate against the explicit inputs the user supplied; rules like
	// `required` fall back to the now-populated field values, so values sourced
	// from the config file or defaults still satisfy them.
	validateInputs := map[string]any{}
	env(validateInputs)
	for k, v := range flags {
		validateInputs[k] = v
	}
	if err := command.Validate(validateInputs); err != nil {
		return fmt.Errorf("failed to validate command %q: %w", command.Name(""), err)
	}

	return nil
}

// bindConfigTree hands the app Config to every registered command and subcommand
// that can receive it (anything embedding BaseCommand), plus the default command.
// Walking the tree at Run time avoids the ordering pitfalls of binding at
// registration, when a parent may be added before its config is known.
func (c *app) bindConfigTree() {
	var walk func(cmds []Command[any])
	walk = func(cmds []Command[any]) {
		for _, cmd := range cmds {
			if binder, ok := cmd.(configBinder); ok {
				binder.bindConfig(c.config)
			}
			walk(cmd.Commands())
		}
	}
	walk(c.commands)

	if c.defaultCommand != nil {
		if binder, ok := c.defaultCommand.(configBinder); ok {
			binder.bindConfig(c.config)
		}
	}
}

// resolveStrategy resolves the effective merge strategy from a command's declared
// strategy: it wins, falling back to the app-wide Config.Merge, and finally to
// MergeEnvFlags when neither is set (both MergeInherit).
func (c *app) resolveStrategy(cmdStrategy MergeStrategy) MergeStrategy {
	strategy := cmdStrategy
	if strategy == MergeInherit {
		strategy = c.config.Merge
	}
	if strategy == MergeInherit {
		strategy = MergeEnvFlags
	}
	return strategy
}

func env(commandOptions map[string]any) {
	environ := os.Environ()
	for _, env := range environ {
		pair := strings.SplitN(env, "=", 2)
		commandOptions[pair[0]] = pair[1]
	}
}

func (c *app) printVersion() {
	fmt.Printf("%s %s\n", c.config.Name, c.config.Version)
}

func (c *app) runHelp(args []string, opts ...map[string]any) error {
	options := map[string]any{}
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}

	for _, cmd := range c.commands {
		if cmd.Name("") == helpCommand {
			err := cmd.Run(*c.globalOptions, Unknowns{
				Args:    args,
				Options: options,
			})
			if err != nil {
				return fmt.Errorf("failed to run help command: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("help command not found")
}

const (
	shellCompDirectiveNoFileComp = 4
)

func (c *app) handleComplete(args []string) {
	toComplete := ""
	if len(args) > 0 {
		toComplete = args[len(args)-1]
		args = args[:len(args)-1]
	}

	// walk args to find the deepest matching command
	commands := c.commands
	var matched Command[any]
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		found := c.matchCommandByName(arg, commands)
		if found == nil {
			break
		}
		matched = found
		commands = found.Commands()
	}

	if strings.HasPrefix(toComplete, "-") {
		prefix := strings.TrimLeft(toComplete, "-")
		c.completeFlagNames(matched, prefix)
	} else {
		for _, cmd := range commands {
			name := cmd.Name("")
			if strings.HasPrefix(name, toComplete) {
				fmt.Fprintf(os.Stdout, "%s\t%s\n", name, cmd.Help())
			}
		}
	}

	fmt.Fprintf(os.Stdout, ":%d\n", shellCompDirectiveNoFileComp)
}

func (c *app) completeFlagNames(cmd Command[any], prefix string) {
	seen := make(map[string]bool)

	if cmd != nil {
		c.completeFlagsFromOptions(cmd.Options(), prefix, seen)
	}
	c.completeFlagsFromOptions(c.globalOptions, prefix, seen)
}

func (c *app) completeFlagsFromOptions(options any, prefix string, seen map[string]bool) {
	if options == nil {
		return
	}

	fields, err := structs.GetStructFields(options, nil, structs.DefaultEncodingTags)
	if err != nil {
		return
	}

	for _, field := range fields {
		name := field.Tags["arg"]
		if name == "" {
			continue
		}
		if seen[name] {
			continue
		}
		if strings.HasPrefix(name, prefix) {
			seen[name] = true
			fmt.Fprintf(os.Stdout, "--%s\t%s\n", name, field.Tags["help"])
		}
	}
}

func (c *app) getGlobalOptions(osArgs []string) (map[string]any, map[string]any) {
	// c.globalOptions is always a non-nil *GlobalOptions (set once in NewApp),
	// so GetStructFields cannot return an error here.
	globalFields, _ := structs.GetStructFields(c.globalOptions, nil, structs.DefaultEncodingTags)

	_, _, globalOptions, unknownOptions := getCommandArgs(osArgs, globalFields)

	return globalOptions, unknownOptions
}

func (c *app) matchCommandByArgs(args []string) (Command[any], []string, []string, error) {
	var command Command[any]
	var commandNameIndexes []int

	for a := 0; a < len(args); a++ {
		// previous arg is a command
		// assert if this arg is a sub command
		if command != nil {
			subCommand := c.matchCommandByName(args[a], command.Commands())
			if subCommand != nil {
				command = subCommand
				commandNameIndexes = append(commandNameIndexes, a)
				continue
			}

			// no sub command found, but
			// we have a command already so let's keep it and try further args
			break
		}

		cmd := c.matchCommandByName(args[a], c.commands)
		if cmd != nil {
			command = cmd
			commandNameIndexes = append(commandNameIndexes, a)
			continue
		}
	}

	if command == nil {
		return nil, nil, nil, ErrCommandNotFound
	}

	// create a new slice that excludes the command args
	// we don't need the command args anymore
	allOtherArgs := make([]string, 0)
	commandNameArgs := make([]string, 0)
	for i := 0; i < len(args); i++ {
		if exists(commandNameIndexes, i) {
			commandNameArgs = append(commandNameArgs, args[i])
			continue
		}
		allOtherArgs = append(allOtherArgs, args[i])
	}

	return command, commandNameArgs, allOtherArgs, nil
}

func exists(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}

	return false
}

func (c *app) matchCommandByName(arg string, commands []Command[any]) Command[any] {
	var command Command[any]
	for i := 0; i < len(commands); i++ {
		cmd := commands[i]
		if cmd.Name("") == arg {
			command = cmd
			break
		}
	}

	return command
}
