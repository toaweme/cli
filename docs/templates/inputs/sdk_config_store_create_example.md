```go
import "github.com/toaweme/cli/config"

store := config.NewFileStore(config.HomePath("myapp"))         // ~/.myapp/, 0644
secrets := config.NewSecretFileStore(config.HomePath("myapp")) // ~/.myapp/, 0600
```
