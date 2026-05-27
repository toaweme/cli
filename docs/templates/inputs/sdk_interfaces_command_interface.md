```go
type Command[T any] interface {
    // Name gets or sets the command name. Pass "" to get, non-empty to set.
    Name(name string) string
    // Add registers a subcommand under this command.
    Add(name string, cmd Command[any])
    // Options returns a pointer to the config struct for flag parsing.
    Options() any
    // Commands returns the list of registered subcommands.
    Commands() []Command[any]
    // Run executes the command logic with parsed global options and unknown args.
    Run(options GlobalOptions, unknowns Unknowns) error
    // Validate checks the parsed options map against struct validation rules.
    Validate(options map[string]any) error
    // Help returns a short one-line description shown in command listings.
    Help() string
}
```
