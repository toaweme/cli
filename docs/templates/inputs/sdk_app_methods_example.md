```go
// NewApp builds the app from its identity and built-in global flags.
app := cli.NewApp(
    cli.Settings{Name: "myapp", Version: "1.0.0"},
    cli.GlobalOptions{},
)

// Add registers a command and returns it, so subcommands can be chained.
build := app.Add("build", &BuildCommand{BaseCommand: cli.NewBaseCommand[BuildConfig]()})
build.Add("clean", &CleanCommand{BaseCommand: cli.NewBaseCommand[CleanConfig]()})

// Default runs when the binary is invoked with no arguments.
app.Default(build)

// Help registers the help command under the reserved name (no magic string).
app.Help(help.NewHelpCommand(app.Settings, app.Commands))

// Settings and Commands expose what the app was built with.
name := app.Settings().Name
cmds := app.Commands()
_, _ = name, cmds

// Run dispatches os.Args and returns sentinel errors for help/version.
if err := app.Run(os.Args[1:]); err != nil {
    if errors.Is(err, cli.ErrShowingHelp) || errors.Is(err, cli.ErrShowingVersion) {
        return
    }
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}
```
