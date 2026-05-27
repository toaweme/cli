```go
// walk up the directory tree to find config files
path := config.Discover(".", []string{".myapp.yaml", ".myapp.json"})

// get the user home config directory
dir := config.HomePath("myapp") // ~/.myapp
```
