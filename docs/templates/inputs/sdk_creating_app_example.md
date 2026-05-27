```go
app := cli.NewApp(
    cli.Settings{Name: "myapp", Version: "1.0.0"},
    cli.GlobalOptions{Cwd: cwd},
)
```
