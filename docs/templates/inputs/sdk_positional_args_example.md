```go
type DeployConfig struct {
    Environment string `arg:"0" help:"Target environment" rules:"required"`
    Version     string `arg:"1" help:"Version to deploy"`
}
```

```
myapp deploy production v1.2.3
```
