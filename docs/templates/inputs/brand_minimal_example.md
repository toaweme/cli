```go
func main() {
    app := cli.NewApp(
        cli.Settings{Name: "myapp", Version: "0.1.0"},
        cli.GlobalOptions{},
    )

    app.Add("greet", &GreetCommand{BaseCommand: cli.NewBaseCommand[GreetConfig]()})

    app.Run(os.Args[1:])
}
```
