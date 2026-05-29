```go
cfg := cli.NewConfig(cli.NewFileConfig(cli.FileConfig{
    Name:       "myapp",
    PerProject: true, // also reads a ".myapp" dir found by walking up
}))

type Settings struct {
    Host string `arg:"host" env:"MYAPP_HOST" json:"host" default:"localhost"`
    Port int    `arg:"port" env:"MYAPP_PORT" json:"port" default:"8080"`
}

var s Settings
// merges low -> high: defaults -> home store -> project store -> env -> flags
err := cfg.LoadLayered(&s, cli.LoadOptions{
    Env:   true,
    Flags: map[string]any{"port": 9090}, // e.g. parsed CLI options; highest priority
})
```
