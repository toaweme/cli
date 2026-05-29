# Features

## Public SDK

### App & commands
- [x] Struct-tag-driven command config: embed `BaseCommand[T]`, parsed values land on `c.Inputs`
- [x] App lifecycle: `NewApp`, `Add` (chainable for subcommands), `Default` command, `Help` (registers help under a reserved name), `Commands`, `Config`, `Run` dispatch
- [x] Control-flow sentinel errors: `ErrShowingHelp`, `ErrShowingVersion`, `ErrNoCommands`, `ErrCommandNotFound`, `ErrNoArguments`, `ErrDisplaySubCommands`
- [x] Pass-through of unmatched tokens via `Unknowns` (positional args + unknown flags)

### Flags & validation
- [x] Env var binding via `env` tag and `default` tag fallback values
- [x] Validation rules: `required` and `oneof` (enum), combinable with `|`

### Help & output
- [x] Help system: one-line `Help`, multiline `Description`, `Examples`, per-arg and per-flag detail docs
- [x] Help output formats: `plain`, `plain-flags`, `pretty`, `md`, `json`, `jsonschema`
- [x] Pluggable output codecs via `cli.Config.Formats`: registered yaml/toml codecs add `--format` values, render the command tree, and appear in the `--format` hint, with the core free of those libraries (structural `OutputCodec`)
- [x] Built-in commands: help, version, completion (bash/zsh/fish), dev

### Config & storage
- [x] Storage SDK: `NewStorage` / `NewFileStorage`, primary `Store` and 0600 `Secrets`, per-project dirs
- [x] Layered config `Load`: defaults -> config stores (home, then per-project) -> env -> flags
- [x] Config codecs: JSON built in, plus yaml/toml/json addons
- [x] Merge strategy for command configs: app-wide `Config.Merge` (`MergeInherit`/`MergeEnvFlags`/`MergeLayered`) with per-command `ConfigStrategy` override
- [x] Config field remapping via `ConfigMapping` (`TopLevelMapping`, `Namespaced`)

## Internal

### Argument parsing engine
- [x] Arg parser: positional args by numeric `arg` index, long/short flags, `--key=value` and `--key value`, bare bool flags
- [x] Nested struct flags addressed by dotted FQN (`--database.host`) with env joined by `_`
- [x] Comma-separated scalar slice flags with `sep` tag override

### Rendering & runtime
- [x] Oneof allowed values surfaced in help, flag tables, and JSON Schema, including nested struct sub-fields
- [x] Shell completion runtime: dynamic command and flag-name completion behind the completion scripts
- [x] Dotenv loading

# Todo
- [ ] `--format` rejection error lists only built-in formats; include registered `Config.Formats` names in the "must be one of" message (the oneof rule in mapStructToOptions/utils.go does not know about app.go `isRegisteredFormat`)
