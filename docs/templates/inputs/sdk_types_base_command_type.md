```go
// BaseCommand provides default implementations for the Command interface.
// Embed it in your command struct to get name management, subcommand registration,
// config struct handling, and validation for free.
type BaseCommand[T any] struct {
    command  string
    commands []Command[any]
    Inputs   *T
}
```
