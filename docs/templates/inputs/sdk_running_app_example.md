```go
err := app.Run(os.Args[1:])
if err != nil {
    if errors.Is(err, cli.ErrShowingHelp) || errors.Is(err, cli.ErrShowingVersion) {
        return
    }
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}
```
