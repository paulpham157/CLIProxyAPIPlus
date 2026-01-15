package providers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/constant"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/runtime/executor"
)

// ProviderType represents the type of AI service provider.
type ProviderType string

const (
	ProviderTypeGemini       ProviderType = constant.Gemini
	ProviderTypeGeminiCLI    ProviderType = constant.GeminiCLI
	ProviderTypeCodex        ProviderType = constant.Codex
	ProviderTypeClaude       ProviderType = constant.Claude
	ProviderTypeOpenAI       ProviderType = constant.OpenAI
	ProviderTypeAntigravity  ProviderType = constant.Antigravity
	ProviderTypeKiro         ProviderType = constant.Kiro
	ProviderTypeCursor       ProviderType = constant.Cursor
)

// Provider represents a configured AI service provider with its executor and metadata.
type Provider struct {
	Type        ProviderType
	Name        string
	Description string
	Executor    interface{}
	Enabled     bool
}

// ProviderFactory manages the creation and retrieval of AI service providers.
type ProviderFactory struct {
	cfg       *config.Config
	providers map[ProviderType]Provider
	mu        sync.RWMutex
}

// NewProviderFactory creates a new provider factory with the given configuration.
func NewProviderFactory(cfg *config.Config) *ProviderFactory {
	factory := &ProviderFactory{
		cfg:       cfg,
		providers: make(map[ProviderType]Provider),
	}
	factory.initializeProviders()
	return factory
}

// initializeProviders initializes all available providers based on configuration.
func (f *ProviderFactory) initializeProviders() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.cfg == nil {
		return
	}

	if len(f.cfg.GeminiKey) > 0 {
		f.providers[ProviderTypeGemini] = Provider{
			Type:        ProviderTypeGemini,
			Name:        "Gemini",
			Description: "Google Gemini AI service",
			Executor:    executor.NewGeminiExecutor(f.cfg),
			Enabled:     true,
		}
	}

	if len(f.cfg.CodexKey) > 0 {
		f.providers[ProviderTypeCodex] = Provider{
			Type:        ProviderTypeCodex,
			Name:        "Codex",
			Description: "OpenAI Codex service",
			Executor:    executor.NewCodexExecutor(f.cfg),
			Enabled:     true,
		}
	}

	if len(f.cfg.ClaudeKey) > 0 {
		f.providers[ProviderTypeClaude] = Provider{
			Type:        ProviderTypeClaude,
			Name:        "Claude",
			Description: "Anthropic Claude AI service",
			Executor:    executor.NewClaudeExecutor(f.cfg),
			Enabled:     true,
		}
	}

	if len(f.cfg.KiroKey) > 0 {
		f.providers[ProviderTypeKiro] = Provider{
			Type:        ProviderTypeKiro,
			Name:        "Kiro",
			Description: "AWS CodeWhisperer (Kiro) service",
			Executor:    executor.NewKiroExecutor(f.cfg),
			Enabled:     true,
		}
	}

	if len(f.cfg.OpenAICompatibility) > 0 {
		providerName := "openai-compatibility"
		if len(f.cfg.OpenAICompatibility) > 0 && f.cfg.OpenAICompatibility[0].Name != "" {
			providerName = f.cfg.OpenAICompatibility[0].Name
		}
		f.providers[ProviderTypeOpenAI] = Provider{
			Type:        ProviderTypeOpenAI,
			Name:        "OpenAI Compatible",
			Description: "OpenAI compatible API service",
			Executor:    executor.NewOpenAICompatExecutor(providerName, f.cfg),
			Enabled:     true,
		}
	}

	f.providers[ProviderTypeGeminiCLI] = Provider{
		Type:        ProviderTypeGeminiCLI,
		Name:        "Gemini CLI",
		Description: "Google Gemini CLI service",
		Executor:    executor.NewGeminiCLIExecutor(f.cfg),
		Enabled:     true,
	}

	f.providers[ProviderTypeAntigravity] = Provider{
		Type:        ProviderTypeAntigravity,
		Name:        "Antigravity",
		Description: "Antigravity AI service",
		Executor:    executor.NewAntigravityExecutor(f.cfg),
		Enabled:     true,
	}

	f.providers[ProviderTypeCursor] = Provider{
		Type:        ProviderTypeCursor,
		Name:        "Cursor",
		Description: "Cursor AI service",
		Executor:    executor.NewCursorExecutor(f.cfg),
		Enabled:     true,
	}
}

// GetProvider returns the provider instance for the specified type.
// Returns an error if the provider type is not found or not enabled.
func (f *ProviderFactory) GetProvider(providerType ProviderType) (*Provider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	provider, exists := f.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("provider type %q not found", providerType)
	}

	if !provider.Enabled {
		return nil, fmt.Errorf("provider type %q is not enabled", providerType)
	}

	return &provider, nil
}

// GetProviderByName returns the provider instance for the specified name (case-insensitive).
// Returns an error if the provider name is not found or not enabled.
func (f *ProviderFactory) GetProviderByName(name string) (*Provider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	normalizedName := strings.ToLower(strings.TrimSpace(name))
	for _, provider := range f.providers {
		if strings.ToLower(string(provider.Type)) == normalizedName {
			if !provider.Enabled {
				return nil, fmt.Errorf("provider %q is not enabled", name)
			}
			return &provider, nil
		}
	}

	return nil, fmt.Errorf("provider %q not found", name)
}

// ListProviders returns metadata for all available providers.
// This includes both enabled and disabled providers with their current status.
func (f *ProviderFactory) ListProviders() []ProviderMetadata {
	f.mu.RLock()
	defer f.mu.RUnlock()

	metadata := make([]ProviderMetadata, 0, len(f.providers))
	for _, provider := range f.providers {
		metadata = append(metadata, ProviderMetadata{
			Type:        string(provider.Type),
			Name:        provider.Name,
			Description: provider.Description,
			Enabled:     provider.Enabled,
		})
	}

	return metadata
}

// ProviderMetadata contains metadata information about a provider.
type ProviderMetadata struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// UpdateConfiguration updates the factory configuration and reinitializes providers.
// This should be called when the configuration changes at runtime.
func (f *ProviderFactory) UpdateConfiguration(cfg *config.Config) {
	f.mu.Lock()
	f.cfg = cfg
	f.providers = make(map[ProviderType]Provider)
	f.mu.Unlock()

	f.initializeProviders()
}

// IsProviderAvailable checks if a provider type is available and enabled.
func (f *ProviderFactory) IsProviderAvailable(providerType ProviderType) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	provider, exists := f.providers[providerType]
	return exists && provider.Enabled
}

// GetEnabledProviders returns a list of all enabled provider types.
func (f *ProviderFactory) GetEnabledProviders() []ProviderType {
	f.mu.RLock()
	defer f.mu.RUnlock()

	enabled := make([]ProviderType, 0, len(f.providers))
	for _, provider := range f.providers {
		if provider.Enabled {
			enabled = append(enabled, provider.Type)
		}
	}

	return enabled
}
