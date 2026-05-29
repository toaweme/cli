| Tag | Purpose | Example |
|-----|---------|---------|
| `arg` | Flag name, or numeric index for positional args | `arg:"port"`, `arg:"0"` |
| `short` | Single-char shorthand | `short:"p"` |
| `env` | Environment variable name | `env:"SERVER_PORT"` |
| `help` | Description shown in help output | `help:"Port to listen on"` |
| `default` | Value used when the flag is omitted | `default:"8080"` |
| `rules` | Validation rules | `rules:"required"` |
| `sep` | Separator for `[]T` flags (default `,`) | `sep:"|"` |
