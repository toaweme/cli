```go
type GlobalOptions struct {
    Cwd       string `arg:"cwd" short:"c" env:"CWD" help:"Current working directory"`
    Help      bool   `arg:"help" short:"h" env:"HELP" help:"Show help"`
    Version   bool   `arg:"version" short:"v" env:"VERSION" help:"Show version"`
    Verbosity int    `arg:"verbosity" env:"VERBOSITY" help:"Verbosity level (0, 1, 2)"`
    Format    string `arg:"format" help:"Help output format (plain, plain-flags, pretty, md, json, jsonschema)"`
}
```
