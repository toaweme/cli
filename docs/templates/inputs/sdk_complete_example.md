```go
package main

import (
    "errors"
    "fmt"
    "os"

    "github.com/toaweme/cli"
    "github.com/toaweme/cli/commands/completion"
    "github.com/toaweme/cli/commands/help"
    "github.com/toaweme/cli/commands/version"
    "github.com/toaweme/cli/config"
)

func main() {
    cli.DotEnv()

    cwd, _ := os.Getwd()
    app := cli.NewApp(
        cli.Settings{Name: "myapp", Version: "1.0.0"},
        cli.GlobalOptions{Cwd: cwd},
    )

    app.Add("help", help.NewHelpCommand(app.Settings, app.Commands))
    app.Add("version", version.NewVersionCommand(app.Settings))
    app.Add("completion", completion.NewCompletionCommand("myapp"))

    buildCmd := &BuildCommand{BaseCommand: cli.NewBaseCommand[BuildConfig]()}
    app.Add("build", buildCmd)
    app.Default(buildCmd)

    app.Add("serve", &ServeCommand{BaseCommand: cli.NewBaseCommand[ServeConfig]()})

    store := config.NewFileStore(config.HomePath("myapp"))
    cfgParent := help.NewParentPlaceholder()
    cfgParent.Add("show", &ConfigShowCommand{
        BaseCommand: cli.NewBaseCommand[ConfigShowConfig](),
        store:       store,
    })
    cfgParent.Add("set", &ConfigSetCommand{
        BaseCommand: cli.NewBaseCommand[ConfigSetConfig](),
        store:       store,
    })
    app.Add("config", cfgParent)

    err := app.Run(os.Args[1:])
    if err != nil {
        if errors.Is(err, cli.ErrShowingHelp) || errors.Is(err, cli.ErrShowingVersion) {
            return
        }
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}
```
