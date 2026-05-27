```go
// save
store.Save("settings", map[string]any{"theme": "dark"})

// load
var settings map[string]any
store.Load("settings", &settings)

// check and delete
if store.Exists("settings") {
    store.Delete("settings")
}
```
