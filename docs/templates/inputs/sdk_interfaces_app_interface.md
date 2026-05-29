```go
type App interface {
    Commands() []Command[any]
    Settings() Settings
    Default(cmd Command[any]) Command[any]
    Add(name string, cmd Command[any]) Command[any]
    Run(osArgs []string) error
    // Help registers cmd as the help command under the reserved name, so callers
    // never type it themselves. Use instead of Add for the help command.
    Help(cmd Command[any]) Command[any]
}
```
