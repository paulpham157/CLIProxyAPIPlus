// Package main demonstrates how to integrate the Bolt provider executor
// for streaming code generation with WebContainer execution context.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/runtime/executor"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/api"
	sdkAuth "github.com/router-for-me/CLIProxyAPI/v6/sdk/auth"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy"
	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize token store
	tokenStore := sdkAuth.GetTokenStore()
	if dirSetter, ok := tokenStore.(interface{ SetBaseDir(string) }); ok {
		dirSetter.SetBaseDir(cfg.AuthDir)
	}

	// Create core auth manager
	core := coreauth.NewManager(tokenStore, nil, nil)

	// Register Bolt executor
	boltExec := executor.NewBoltExecutor(cfg)
	core.RegisterExecutor(boltExec)

	fmt.Println("✓ Registered Bolt executor for streaming code generation")

	// Setup hooks to register Bolt models
	hooks := cliproxy.Hooks{
		OnAfterStart: func(s *cliproxy.Service) {
			// Register Bolt models for discovery via /v1/models endpoint
			models := []*cliproxy.ModelInfo{
				{
					ID:          "bolt-sonnet",
					Object:      "model",
					Type:        "bolt",
					DisplayName: "Bolt Sonnet (Code Generation)",
				},
				{
					ID:          "bolt-sonnet-thinking",
					Object:      "model",
					Type:        "bolt",
					DisplayName: "Bolt Sonnet with Extended Thinking",
				},
				{
					ID:          "bolt-opus",
					Object:      "model",
					Type:        "bolt",
					DisplayName: "Bolt Opus (Advanced Code Generation)",
				},
				{
					ID:          "bolt-haiku",
					Object:      "model",
					Type:        "bolt",
					DisplayName: "Bolt Haiku (Fast Code Tasks)",
				},
			}

			// Register models for each Bolt auth
			for _, a := range core.List() {
				if strings.EqualFold(a.Provider, "bolt") {
					cliproxy.GlobalModelRegistry().RegisterClient(a.ID, "bolt", models)
					fmt.Printf("✓ Registered %d Bolt models for auth: %s\n", len(models), a.Label)
				}
			}
		},
	}

	// Build the service
	svc, err := cliproxy.NewBuilder().
		WithConfig(cfg).
		WithConfigPath("config.yaml").
		WithCoreAuthManager(core).
		WithServerOptions(
			api.WithRequestLoggerFactory(func(cfg *config.Config, cfgPath string) api.RequestLogger {
				// Use default request logger
				return nil
			}),
		).
		WithHooks(hooks).
		Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build service: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🚀 Starting Bolt provider service on %s\n", cfg.Server.Host)
	fmt.Println("📝 Endpoints:")
	fmt.Println("   - POST /v1/chat/completions - Streaming code generation")
	fmt.Println("   - GET  /v1/models - List available models")
	fmt.Println("   - POST /v1/chat/completions/count_tokens - Token counting")
	fmt.Println()
	fmt.Println("💡 Example curl command:")
	fmt.Println(`   curl -X POST http://localhost:8080/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{
       "model": "bolt-sonnet",
       "messages": [{"role": "user", "content": "Create a React todo app"}],
       "metadata": {"webcontainer": true},
       "stream": true
     }'`)
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the service
	if errRun := svc.Run(ctx); errRun != nil && !errors.Is(errRun, context.Canceled) {
		fmt.Fprintf(os.Stderr, "Service error: %v\n", errRun)
		os.Exit(1)
	}
}
