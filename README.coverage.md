## Test Coverage

| Package | Coverage | Status |
|---------|----------|--------|
| github.com/toaweme/cli | 74.6% | PASS |
| github.com/toaweme/cli/commands/completion | 94.4% | PASS |
| github.com/toaweme/cli/commands/dev | 29.2% | PASS |
| github.com/toaweme/cli/commands/help | 100.0% | PASS |
| github.com/toaweme/cli/commands/version | 100.0% | PASS |
| github.com/toaweme/cli/config | 85.5% | PASS |
| github.com/toaweme/cli/config/addons/json | 100.0% | PASS |
| github.com/toaweme/cli/help | 78.1% | PASS |

**Total: 73.8%** (excluding examples)

The `commands/dev` package generates documentation by shelling out to `go run`
against the example binaries; only its pure helpers are unit tested, so its
coverage is intentionally low.
