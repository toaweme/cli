```go
type Config struct {
    Port int `arg:"port" short:"p" env:"PORT" default:"8080" help:"Port to listen on" rules:"required"`
}
```
