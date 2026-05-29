```go
// BaseCommand implements these as no-ops; override any of them to enrich help.
func (c *BaseCommand[T]) Description() string        { return "" }
func (c *BaseCommand[T]) Examples() [][]string       { return nil }
func (c *BaseCommand[T]) Args() map[int][]string     { return nil }
func (c *BaseCommand[T]) Flags() map[string][]string { return nil }
```
