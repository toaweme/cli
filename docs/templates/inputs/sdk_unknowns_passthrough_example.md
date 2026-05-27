```go
func (c *MyCommand) Run(_ cli.GlobalOptions, unknowns cli.Unknowns) error {
    // unknowns.Args    - unmatched positional args
    // unknowns.Options - unmatched key-value flags
    return nil
}
```
