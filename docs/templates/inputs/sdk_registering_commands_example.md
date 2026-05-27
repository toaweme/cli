```go
app.Add("serve", &ServeCommand{BaseCommand: cli.NewBaseCommand[ServeConfig]()})
app.Add("build", &BuildCommand{BaseCommand: cli.NewBaseCommand[BuildConfig]()})
```
