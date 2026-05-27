```go
cmd := app.Add("db", help.NewParentPlaceholder())
cmd.Add("migrate", &MigrateCommand{BaseCommand: cli.NewBaseCommand[MigrateConfig]()})
cmd.Add("seed", &SeedCommand{BaseCommand: cli.NewBaseCommand[SeedConfig]()})
```
