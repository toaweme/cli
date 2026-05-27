```go
type App interface {
    Commands() []Command[any]
    Settings() Settings
    Default(cmd Command[any]) Command[any]
    Add(name string, cmd Command[any]) Command[any]
    Run(osArgs []string) error
}
```
