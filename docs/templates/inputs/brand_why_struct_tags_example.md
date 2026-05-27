```go
type Config struct {
    Port   int    `arg:"port" short:"p" env:"PORT" default:"8080" help:"Port to listen on"`
    Host   string `arg:"host" env:"HOST" default:"localhost" help:"Host to bind to"`
    Reload bool   `arg:"reload" short:"r" help:"Enable live reload"`
}
```
