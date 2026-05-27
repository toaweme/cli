```go
var _ cli.ExampleProvider = (*BuildCommand)(nil)

func (c *BuildCommand) Examples() []string {
    return []string{
        "myapp build",
        "myapp build -o ./dist --race",
        "myapp build --tags=integration,e2e",
    }
}
```
