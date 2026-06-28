# Plugin Architecture

Plugins extend OpenForge's capabilities without modifying core code.

## Overview

```
Plugin
  │
  ├── Manifest (plugin.yaml)
  ├── Hooks (lifecycle callbacks)
  ├── Provider (inference backend)
  └── Skills (pipeline steps)
```

## Types of Plugins

| Type | Description | Loading |
|:-----|-------------|:--------|
| **Provider** | Add a new inference backend | Shared object (.so) |
| **Skill Step** | Add custom step type | YAML + optional binary |
| **Hook** | Extend lifecycle events | Shared object (.so) |
| **Middleware** | Intercept API requests | Go interface |

## Provider Plugin

### Interface

```go
package plugin

import "context"

type ProviderPlugin interface {
    Name() string
    Version() string
    Initialize(ctx context.Context) error
    Shutdown(ctx context.Context) error
    // Must also implement runtime.Provider
}
```

### Lifecycle

```
1. Load: plugin.Open("provider.so")
2. Init: Lookup("Plugin") → call Initialize()
3. Register: Add to provider registry
4. Use: Engine routes inference to plugin
5. Stop: Shutdown() → Close()
```

### Example

```yaml
# my-provider/plugin.yaml
name: my-provider
version: 1.0.0
type: provider
description: Custom inference backend
entrypoint: my-provider.so
```

## Plugin Manifest

Every plugin must include a manifest:

```yaml
# plugin.yaml
name: my-plugin
version: 1.0.0
type: provider          # provider, skill-step, hook, middleware
description: |
  Multi-line description of what this plugin does.

author:
  name: Your Name
  email: your@email.com

license: MIT

requires:
  openforge: ">= 0.1.0"
  openvino: ">= 2025.0"

capabilities:
  - chat
  - completion
  - embedding
```

## Installation

```bash
# From directory
openforge plugin install ./my-plugin/

# From archive
openforge plugin install my-plugin.tar.gz

# From URL
openforge plugin install https://plugins.example.com/my-plugin.tar.gz

# List installed
openforge plugin list

# Remove
openforge plugin remove my-plugin

# Enable/disable
openforge plugin enable my-plugin
openforge plugin disable my-plugin
```

## Configuration

```yaml
# config.yaml
plugins:
  enabled:
    - my-provider
    - custom-skills

  settings:
    my-provider:
      api_key: "${MY_PROVIDER_KEY}"
      endpoint: "http://localhost:8080"
```

## Development

### Scaffold a Plugin

```bash
openforge plugin scaffold my-provider
```

Creates:
```
my-provider/
├── main.go          # Plugin implementation
├── plugin.yaml      # Manifest
├── Makefile         # Build
├── README.md        # Documentation
└── test/            # Tests
```

### Build

```bash
cd my-provider
go build -buildmode=plugin -o my-provider.so
```

### Test

```bash
openforge plugin test ./my-provider.so
```

## Security

See [Security](security.md) for plugin sandboxing details.

## Publishing

Community plugins can be listed in the community repository:

```
https://github.com/openforge-ai/plugins
```

Submit a PR with your plugin manifest and source.
