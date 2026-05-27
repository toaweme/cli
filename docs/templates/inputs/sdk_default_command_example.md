```go
buildCmd := &BuildCommand{BaseCommand: cli.NewBaseCommand[BuildConfig]()}
app.Add("build", buildCmd)
app.Default(buildCmd)
```
