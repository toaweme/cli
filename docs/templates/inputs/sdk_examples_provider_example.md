```go
// Each example is a slice of lines: the invocation first, sample output after.
func (c *BuildCommand) Examples() [][]string {
    return [][]string{
        {"myapp build"},
        {"myapp build -o ./dist --race"},
        {
            "myapp build --tags=integration,e2e",
            "building output=./bin tags=[integration e2e] race=false",
        },
    }
}

// Args and Flags attach multi-line detail to positional args and flags.
func (c *BuildCommand) Args() map[int][]string {
    return map[int][]string{0: {"target to build", "defaults to the whole module"}}
}

func (c *BuildCommand) Flags() map[string][]string {
    return map[string][]string{
        "--tags, -t": {"comma-separated build tags", "splits into a []string"},
    }
}
```
