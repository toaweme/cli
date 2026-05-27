```go
// Unknowns holds arguments and options that were not matched to any defined field.
// Commands receive these to support pass-through or dynamic flag handling.
type Unknowns struct {
    // Args are positional arguments not matched to numbered struct tags.
    Args []string
    // Options are key-value flags not defined in the command's config struct.
    Options map[string]any
}
```
