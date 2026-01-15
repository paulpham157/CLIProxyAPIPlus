// Package config provides provider configuration management.
package config

import (
	"os"
	"strconv"
	"strings"
)

// ProviderConfig represents configuration for a specific provider.
type ProviderConfig struct {
	Name    string
	APIKey  string
	Enabled bool
}

// ProvidersConfig holds configuration for all providers.
type ProvidersConfig struct {
	Bolt ProviderConfig
	V0   ProviderConfig
}

// LoadProvidersConfig loads provider configuration from environment variables.
// It reads BOLT_ANTHROPIC_API_KEY, V0_API_KEY, ENABLE_BOLT_PROVIDER, and ENABLE_V0_PROVIDER.
func LoadProvidersConfig() *ProvidersConfig {
	return &ProvidersConfig{
		Bolt: ProviderConfig{
			Name:    "bolt",
			APIKey:  strings.TrimSpace(os.Getenv("BOLT_ANTHROPIC_API_KEY")),
			Enabled: parseBoolEnv("ENABLE_BOLT_PROVIDER", true),
		},
		V0: ProviderConfig{
			Name:    "v0",
			APIKey:  strings.TrimSpace(os.Getenv("V0_API_KEY")),
			Enabled: parseBoolEnv("ENABLE_V0_PROVIDER", true),
		},
	}
}

// GetEnabledProviders returns a list of all enabled provider configurations.
func (pc *ProvidersConfig) GetEnabledProviders() []ProviderConfig {
	var enabled []ProviderConfig
	
	if pc.Bolt.Enabled && pc.Bolt.APIKey != "" {
		enabled = append(enabled, pc.Bolt)
	}
	
	if pc.V0.Enabled && pc.V0.APIKey != "" {
		enabled = append(enabled, pc.V0)
	}
	
	return enabled
}

// IsProviderEnabled checks if a specific provider is enabled and has an API key configured.
func (pc *ProvidersConfig) IsProviderEnabled(providerName string) bool {
	switch strings.ToLower(providerName) {
	case "bolt":
		return pc.Bolt.Enabled && pc.Bolt.APIKey != ""
	case "v0":
		return pc.V0.Enabled && pc.V0.APIKey != ""
	default:
		return false
	}
}

// GetProviderAPIKey retrieves the API key for a specific provider.
// Returns empty string if provider is not found or not configured.
func (pc *ProvidersConfig) GetProviderAPIKey(providerName string) string {
	switch strings.ToLower(providerName) {
	case "bolt":
		return pc.Bolt.APIKey
	case "v0":
		return pc.V0.APIKey
	default:
		return ""
	}
}

// parseBoolEnv parses a boolean environment variable with a default value.
// Returns defaultVal if the environment variable is not set or cannot be parsed.
func parseBoolEnv(key string, defaultVal bool) bool {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}
	
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return defaultVal
	}
	
	return parsed
}
