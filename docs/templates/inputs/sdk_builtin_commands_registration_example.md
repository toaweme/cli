```go
import (
    "github.com/toaweme/cli/commands/help"
    "github.com/toaweme/cli/commands/version"
    "github.com/toaweme/cli/commands/completion"
)

app.Add("help", help.NewHelpCommand(app.Settings, app.Commands))
app.Add("version", version.NewVersionCommand(app.Settings))
app.Add("completion", completion.NewCompletionCommand("myapp"))
```
