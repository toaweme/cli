```go
// Codec serializes and deserializes config values.
type Codec interface {
    Marshal(v any) ([]byte, error)
    Unmarshal(data []byte, v any) error
    Extension() string
}
```
