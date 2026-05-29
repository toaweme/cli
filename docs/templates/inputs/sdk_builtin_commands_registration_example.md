```go
import (
    "github.com/toaweme/cli/commands/help"
    "github.com/toaweme/cli/commands/version"
    "github.com/toaweme/cli/commands/completion"
)

// Help registers the help command under the reserved name for you.
app.Help(help.NewHelpCommand(app.Settings, app.Commands))
app.Add("version", version.NewVersionCommand(app.Settings))
app.Add("completion", completion.NewCompletionCommand("myapp"))
```
