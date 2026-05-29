# TODO

## Now: auto-merge app config into command configs (no wiring)

Goal: a command's `Options()` struct is populated automatically, layered
**defaults → app store section → env → flags** (later wins), keyed by struct tags.
End users never call `Load` or wire anything; they just declare fields.

- [ ] In `app.Run`, when `app.config.Store != nil`, populate the matched command's
      `Options()` via the existing layered engine, e.g.
      `app.config.Store.Load(command.Options(), cli.LoadOptions{Env: true, Flags: flags})`
      (reuse `Storage.Load`; `flags` = parsed flags + positionals keyed by arg name).
- [ ] When `Store == nil`, fall back to env + flags only (current behavior).
- [ ] Fix precedence: flags must beat env (today's single `structs.Set` checks the
      `env:` tag first, so env wrongly wins). Sequential layering fixes this.
- [ ] Reconcile `Validate` with the layered load (validate the merged result).
- [ ] Decide the per-command store key: start with always `"config"`; add an optional
      `ConfigKey()` override only if needed.
- [ ] Verify nested case end-to-end: `Database database.Connection` filled from
      `app.yml` `database:` block, overridden by `DATABASE_*`, then `--database.*`.

## types.go

- [ ] Finish moving core declarations (`App`, `Command[T]`, `Config`, `Storage`,
      `LoadOptions`, `GlobalOptions`, `Unknowns`) into `types.go`; keep logic in
      `app.go`/`config_*.go`/`command.go`.
- [ ] Drop `CommandSDK { Config() Config }` — injecting a config accessor is the
      auto-wire pattern we rejected; the auto-merge above is the declarative answer.

## Smaller cleanups

- [ ] Render slice element type (`[]string`) in help instead of `slice`.
- [ ] Move global `--verbosity` / `--format` allowed values to a `oneof` rule (drop
      the hardcoded `globalOptionValues` map in help/docs.go).
- [ ] Delete orphaned `docs/templates/inputs/sdk_*` partials (README is self-contained now).

## Decided against

- [ ] ~~Bare string-flag defaults (`--format` with no value → `default:` tag)~~ —
      ambiguous with `-opt val`; do not implement.
