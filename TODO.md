# TODO

Working notes for the CLI framework. Pick up from here in a new session.

## Uncommitted work (this session)

Everything below is in the working tree, **not committed**. Build + vet + full test
suite pass (root, `commands/*`, `config`, and the `yaml`/`toml` submodules).

Suggested commit: `feat: multiline command descriptions and config SDK`

Changed/new files:

- `config.go` (new) - root-package config SDK
- `config_test.go` (new)
- `command.go` - added `DescriptionProvider` interface
- `app.go` - added `Settings.Config` field
- `help/help.go`, `help/docs.go`, `help/json.go` - render `Description()`
- `commands/completion/*` - completion `Description()` with install instructions
- `examples/full/*` - demonstrates the new config SDK via constructor injection

### Done: multiline command descriptions (API-compatible)

Optional `cli.DescriptionProvider` interface (`Description() string`), mirroring
`ExampleProvider`. `Help()` stays the one-line summary used in listings;
`Description()` carries the rich body. Rendered in single-command help, agent docs
(md/plain), and JSON (`CommandInfo.Description`). Listings defend against a
multiline `Help()` via `firstLine`. `completion` implements it with per-shell
install instructions.

### Done: config SDK (root package)

```go
type Config interface {
    config.Store                 // Load / Save / Delete / Exists
    Secrets() config.Store       // 0600 files under <dir>/secrets
    Dir() string
}
type FileConfig struct {
    Dir        string            // leading "~" expands; defaults to "~/.<Name>"
    Name       string
    PerProject bool              // walk up for a ".<Name>" dir, else fall back
    Codecs     []config.Codec    // json built-in; pass yaml.Codec / toml.Codec
}
type ConfigSource interface { Stores() (cfg, secrets config.Store, dir string) }
func NewFileConfig(opts FileConfig) ConfigSource
func NewConfig(src ConfigSource) Config
// Settings gained: Config Config  (optional, nil-safe)
```

Decisions:
- Struct-options form of `NewFileConfig` (not positional) - scales to PerProject/Codecs.
- Two layers (`NewConfig(NewFileConfig(...))`) so backends are swappable.
- **No auto-wiring.** Config reaches commands via explicit constructor injection
  (`NewFooCommand(cfg cli.Config)`). An earlier `ConfigAware`/`SetConfig` auto-inject
  was rejected as "magic". `Settings.Config` holds it for the app/built-ins.
- `cli.NewApp()` name kept (not renamed to `cli.New()`).

## Open

### Arg parser: comma-separated `[]string` (do this)

`--tags=a,b,c` -> `["a","b","c"]`, same for env vars; optional `sep:"|"` override.
Unambiguous and useful. Changes in `args.go` + `toaweme/structs` field handling.
Needs table tests for edge cases.

### Arg parser: bare string-flag defaults (DROPPED - do not do)

`--format` (no value) -> `default:` tag. Rejected: conflicts with the space-separated
value form (`-opt val`), making `myapp --format build` ambiguous (value or command?).
Bools dodge this; strings can't. If a "default format" shortcut is ever wanted, use an
explicit bool flag instead, not bare-string magic.

### Config: merge-layer loader (not started)

Only direct store access exists. Planned layered load (low -> high):
defaults -> `~/.<app>/config` -> project config (walk-up) -> env -> CLI flags.
`PerProject` is currently single-store ("project `.<Name>` dir if found, else home"),
no merge.

### Help: multiline flag/arg descriptions (not started)

Struct `help:` tags are single-line. Would need a newline convention in the tag or a
separate mechanism. Also: multiline usage examples with output samples.

## Conventions reminder

- New optional capabilities follow the `ExampleProvider`/`DescriptionProvider` pattern:
  optional interface, type-checked at the render site only, never injected.
- Prefer explicit constructor injection over interface-assertion auto-wiring.

## Env note

GoLand injected a stale `GOROOT=/usr/local/go` (go1.26.0) while homebrew `go` is
1.26.3, which broke `go run` in tests (`compile: version mismatch`). Fixed by pointing
GoLand's Go SDK at `/opt/homebrew/opt/go/libexec`. If it recurs, `env -u GOROOT` works
around it.
