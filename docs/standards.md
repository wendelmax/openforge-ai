# Standards

## Code Standards

### Go

- **Go 1.23+** with modern Go idioms (range-over-func, iterators)
- **Packages**: lowercase, single word, singular (`model.go` not `models.go`)
- **Files**: one primary type per file, `_test.go` alongside
- **Imports**: stdlib → third-party → internal (three groups, blank line between)
- **Naming**: avoid abbreviations (except `ctx`, `wg`, `mu`, `err`)
- **Comments**: GoDoc on every exported symbol
- **Errors**: wrapped with context: `fmt.Errorf("context: %w", err)`
- **Zero values**: prefer zero-value usability over constructors

### Architecture

- **Clean Architecture**: handlers → usecases → providers, dependencies inward
- **SOLID**: every principle, every time
- **Keep it simple**: YAGNI — build what's needed, not what's imagined
- **Small interfaces**: 1-3 methods, defined in the consumer package
- **DI everywhere**: no `init()`, no globals, no singletons

### Testing

- **Unit tests**: required for all exported functions
- **Integration tests**: `//go:build integration` for provider tests
- **Table-driven tests**: prefer subtests for multiple scenarios
- **Mocks**: `mockery` or manual interfaces for test doubles
- **Coverage**: 90% minimum, tracked in CI
- **Naming**: `TestFuncName_Scenario_ExpectedBehavior`

## Documentation Standards

- **English** for all technical documentation
- **Portuguese** also accepted for community docs
- **GoDoc**: every export, every package
- **ADRs**: every architectural decision in `docs/adr/`
- **Examples**: runnable examples in `docs/examples/`

## Release Standards

### Versioning

Semantic Versioning: `MAJOR.MINOR.PATCH`

- **MAJOR**: breaking API changes
- **MINOR**: backward-compatible features
- **PATCH**: backward-compatible fixes

### Release Process

1. All tests pass on all Tier 1 platforms
2. Benchmarks show no regression (or regression is documented)
3. Changelog updated
4. Release tagged and signed
5. GitHub Release with binaries for all platforms
6. Docker image published

## Performance Standards

| Metric | Target |
|--------|:------:|
| Startup time | < 2 seconds |
| Time to first token | < 500ms (warm) |
| Memory growth | Bounded, no leaks |
| Concurrent requests | 10+ without degradation |
| Cache hit latency | < 1ms |
| Embedding latency (BGE Small) | < 20ms |

## Security Standards

- No secrets in code (use env vars)
- No network requests (offline-first)
- No eval/exec of untrusted input
- File operations restricted to configured paths
- Telemetry opt-in only
