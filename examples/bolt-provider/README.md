# Bolt Provider Example

This example demonstrates how to configure and use the Bolt provider for streaming code generation with WebContainer execution context.

## Overview

The Bolt provider integrates with Anthropic's Claude API following the Bolt.new open-source implementation pattern. It provides:

- Streaming code generation
- WebContainer execution context
- Extended thinking mode for complex code tasks
- Full Claude API feature support

## Configuration

### 1. Auth File Setup

Create an auth file in your `auths/` directory (e.g., `auths/bolt.yaml`):

```yaml
provider: bolt
attributes:
  api_key: sk-ant-api03-your-anthropic-api-key-here
  base_url: https://api.anthropic.com  # Optional, defaults to Anthropic API
label: Bolt Dev Assistant
```

### 2. Config File Setup

Add Bolt-specific configuration in your `config.yaml`:

```yaml
claude_key:
  - api_key: sk-ant-api03-your-anthropic-api-key-here
    base_url: https://api.anthropic.com
    models:
      - name: claude-sonnet-4-5
        alias: bolt-sonnet
      - name: claude-3-7-sonnet-20250219
        alias: bolt-sonnet-thinking
      - name: claude-opus-4-5
        alias: bolt-opus
```

## Usage

### Basic Request

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "bolt-sonnet",
    "messages": [
      {
        "role": "user",
        "content": "Create a simple React component for a todo list"
      }
    ],
    "stream": true
  }'
```

### WebContainer Context

Enable WebContainer-specific system prompts by passing metadata:

```json
{
  "model": "bolt-sonnet",
  "messages": [
    {
      "role": "user",
      "content": "Set up a Vite + React project with TypeScript"
    }
  ],
  "metadata": {
    "webcontainer": true
  },
  "stream": true
}
```

### Extended Thinking Mode

For complex code generation tasks, enable thinking mode:

```json
{
  "model": "bolt-sonnet-thinking",
  "messages": [
    {
      "role": "user",
      "content": "Design and implement a state machine for a multi-step form with validation"
    }
  ],
  "metadata": {
    "webcontainer": true,
    "thinking": {
      "type": "enabled",
      "budget_tokens": 5000
    }
  },
  "max_tokens": 8000,
  "stream": true
}
```

## Model Mapping

The Bolt provider uses Claude models under the hood. Recommended mappings:

| Bolt Alias | Claude Model | Use Case |
|------------|--------------|----------|
| bolt-sonnet | claude-sonnet-4-5 | General code generation |
| bolt-sonnet-thinking | claude-3-7-sonnet-20250219 | Complex reasoning tasks |
| bolt-opus | claude-opus-4-5 | Advanced code generation |
| bolt-haiku | claude-3-5-haiku-20241022 | Fast, simple tasks |

## Integration Example

```go
package main

import (
    "context"
    "github.com/router-for-me/CLIProxyAPI/v6/internal/config"
    "github.com/router-for-me/CLIProxyAPI/v6/internal/runtime/executor"
    "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy"
    coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

func main() {
    cfg, _ := config.LoadConfig("config.yaml")
    
    // Create Bolt executor
    boltExec := executor.NewBoltExecutor(cfg)
    
    // Register with auth manager
    core := coreauth.NewManager(nil, nil, nil)
    core.RegisterExecutor(boltExec)
    
    // Build and run service
    svc, _ := cliproxy.NewBuilder().
        WithConfig(cfg).
        WithCoreAuthManager(core).
        Build()
    
    svc.Run(context.Background())
}
```
