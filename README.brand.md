# toaweme/cli

A struct-driven CLI framework for Go. One dependency. No codegen. No DSL.

Define a Go struct, tag its fields, and you have a fully wired command with flags, short flags, environment variable binding, defaults, validation, and help output. Nothing else needed.

## Why

Most CLI frameworks make you choose: either a sprawling API with method chains and builder patterns, or a codegen tool that owns your project layout. Both pull in transitive dependencies you did not ask for and abstractions you have to learn before you can write a single command.

This module takes a different approach. Your config is a struct. Your tags are the schema. The framework reads them and does the rest.

{{ file "docs/templates/inputs/brand_why_struct_tags_example.md" }}

That is the entire flag definition. No `StringVar`, no `AddFlag`, no `Bind`. One struct, multiple input sources, zero boilerplate.

## What you get

**Struct tag driven** - flags, positional args, short forms, env vars, defaults, validation rules, and help text are all declared on the struct. One source of truth.

**Environment variable binding** - every field with an `env` tag reads from the environment automatically. Combined with the built-in dotenv loader, your commands work the same locally and in production without any wiring code.

**Built-in dotenv** - `cli.DotEnv()` loads `.env` files before command dispatch. Missing files are silently skipped. Variables already set in the environment are never overwritten.

**Validation** - tag a field `rules:"required"` and the framework rejects the command before `Run()` is called. No manual checks in your handler.

**Subcommands** - commands nest arbitrarily deep. `app db migrate`, `app config set`. Parent placeholders group related commands without requiring handler logic.

**Default command** - one command can be marked as the default. Running the binary with no arguments executes it directly, populated from env vars.

**Shell completions** - built-in `__complete` protocol with bash, zsh, and fish script generation. Commands and flags complete out of the box.

**Multiple help formats** - plain text, pretty (ANSI), markdown, JSON, and JSON Schema. The same command tree powers human-readable help and machine-readable API docs.

**Config store** - file-based key-value storage with JSON built in. YAML and TOML available as codec addons. Separate constructors for regular config (0644) and secrets (0600). Atomic writes.

**Config discovery** - walks up the directory tree looking for config files by name. Finds the nearest `.myapp.yaml` or equivalent without hardcoding paths.

**One dependency** - the entire framework depends on a single module (`toaweme/structs`) for struct reflection. No dependency tree to audit.

**Generics, not reflection hacks** - commands are `Command[T]` where T is your config struct. The framework knows your types at compile time.

## Minimal example

{{ file "docs/templates/inputs/brand_minimal_example.md" }}

That is a working CLI. Add more commands with `app.Add`, nest them with `cmd.Add`, set a default with `app.Default`. The framework handles flag parsing, env binding, validation, help, version, and completions.
