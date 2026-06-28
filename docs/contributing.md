# Contributing

We welcome contributions from the community. Here's how to get started.

## Code of Conduct

Be respectful, inclusive, and professional. We follow the [Contributor Covenant](https://www.contributor-covenant.org/).

## Quick Start

```bash
# Fork the repository
git clone https://github.com/YOUR_USERNAME/openforge.git
cd openforge

# Create a branch
git checkout -b feat/my-feature

# Make changes
# ...

# Run tests
go test ./... -count=1
go vet ./...
golangci-lint run ./...

# Commit
git commit -m "feat: add my feature"
git push origin feat/my-feature
```

Then open a Pull Request.

## Development Setup

```bash
# Install Go 1.23+
# Install OpenVINO 2025.x (for provider development)
# Install development tools
task deps
```

### Without OpenVINO

CGO-dependent code is behind build tags. Most packages compile with `CGO_ENABLED=0`:

```bash
CGO_ENABLED=0 go test ./... -count=1
```

## What We Accept

| Type | Examples | Label |
|------|----------|:-----:|
| 🐛 Bug fixes | Crash fixes, incorrect behavior | `bug` |
| ✨ Features | New endpoints, providers, skills | `feature` |
| 📚 Docs | README, tutorials, API docs | `docs` |
| 🧪 Tests | Unit tests, integration tests, benchmarks | `test` |
| 🔧 Refactors | Code cleanup, performance | `refactor` |
| 🎨 Style | Formatting, naming | `style` |

## Pull Request Process

1. **Small PRs preferred** — one feature/fix per PR
2. **Tests required** — new code must include tests
3. **Documentation** — API changes need doc updates
4. **Benchmarks** — performance changes need benchmark results
5. **No breaking changes** — unless discussed in an issue first

### Checklist

- [ ] Code follows project conventions
- [ ] Tests pass (`go test ./...`)
- [ ] No lint errors (`golangci-lint run ./...`)
- [ ] Go vet passes (`go vet ./...`)
- [ ] New code has tests
- [ ] API changes are documented
- [ ] ADR created (for architectural decisions)
- [ ] Commit messages follow conventional commits

## Commit Convention

```
<type>: <description>

[optional body]
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `style`, `chore`, `perf`

Examples:
```
feat: add NPU device detection
fix: race condition in session manager
docs: update installation guide
test: add embedding cache tests
```

## Issue Tracker

- **Bug reports**: include hardware, OS, OpenVINO version, steps to reproduce
- **Feature requests**: describe the use case and expected behavior
- **Questions**: use GitHub Discussions

## Review Process

1. PR is reviewed within 48 hours
2. CI must pass (tests, lint, build)
3. At least one maintainer approval required
4. Changes requested will be clearly explained

## Becoming a Maintainer

Regular contributors may be invited to become maintainers. Maintainers have:
- Write access to the repository
- Vote on architectural decisions
- Review and merge PRs
- Access to maintainer discussions
