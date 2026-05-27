```go
db := help.NewParentPlaceholder()
app.Add("db", db)
db.Add("migrate", &MigrateCommand{BaseCommand: cli.NewBaseCommand[MigrateConfig]()})
db.Add("seed", &SeedCommand{BaseCommand: cli.NewBaseCommand[SeedConfig]()})
```
