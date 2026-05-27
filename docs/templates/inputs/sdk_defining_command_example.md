```go
type ServeConfig struct {
    Port   int    `arg:"port" short:"p" env:"SERVER_PORT" default:"8080" help:"Port to listen on"`
    Host   string `arg:"host" env:"SERVER_HOST" default:"localhost" help:"Host to bind to"`
    TLS    bool   `arg:"tls" env:"SERVER_TLS" help:"Enable TLS"`
    Reload bool   `arg:"reload" short:"r" help:"Enable live reload"`
}

type ServeCommand struct {
    cli.BaseCommand[ServeConfig]
}

var _ cli.Command[ServeConfig] = (*ServeCommand)(nil)

func (c *ServeCommand) Run(_ cli.GlobalOptions, _ cli.Unknowns) error {
    fmt.Printf("serving %s:%d\n", c.Inputs.Host, c.Inputs.Port)
    return nil
}

func (c *ServeCommand) Help() string {
    return "Start the development server"
}
```
