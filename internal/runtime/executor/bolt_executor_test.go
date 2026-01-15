package executor

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
	"github.com/tidwall/gjson"
)

func TestBoltExecutor_Identifier(t *testing.T) {
	exec := NewBoltExecutor(&config.Config{})
	if exec.Identifier() != "bolt" {
		t.Errorf("expected identifier 'bolt', got '%s'", exec.Identifier())
	}
}

func TestBoltExecutor_injectBoltSystemPrompt(t *testing.T) {
	exec := NewBoltExecutor(&config.Config{})
	
	tests := []struct {
		name           string
		input          string
		expectedPrefix string
	}{
		{
			name:           "empty payload",
			input:          `{}`,
			expectedPrefix: "You are Bolt, an expert AI assistant",
		},
		{
			name:           "with existing system",
			input:          `{"system":"Existing prompt"}`,
			expectedPrefix: "You are Bolt, an expert AI assistant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.injectBoltSystemPrompt([]byte(tt.input))
			systemPrompt := gjson.GetBytes(result, "system").String()
			
			if systemPrompt == "" {
				t.Error("system prompt should not be empty")
			}
			
			if len(systemPrompt) < 100 {
				t.Error("system prompt seems too short")
			}
			
			if !gjson.ValidBytes(result) {
				t.Error("result should be valid JSON")
			}
		})
	}
}

func TestBoltExecutor_resolveAPIKey(t *testing.T) {
	exec := NewBoltExecutor(&config.Config{})
	
	tests := []struct {
		name     string
		auth     *cliproxyauth.Auth
		expected string
	}{
		{
			name:     "nil auth",
			auth:     nil,
			expected: "",
		},
		{
			name: "api key in attributes",
			auth: &cliproxyauth.Auth{
				Attributes: map[string]string{
					"api_key": "sk-ant-test-key",
				},
			},
			expected: "sk-ant-test-key",
		},
		{
			name: "api key in metadata",
			auth: &cliproxyauth.Auth{
				Metadata: map[string]any{
					"api_key": "sk-ant-meta-key",
				},
			},
			expected: "sk-ant-meta-key",
		},
		{
			name: "prefer attributes over metadata",
			auth: &cliproxyauth.Auth{
				Attributes: map[string]string{
					"api_key": "sk-ant-attr-key",
				},
				Metadata: map[string]any{
					"api_key": "sk-ant-meta-key",
				},
			},
			expected: "sk-ant-attr-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.resolveAPIKey(tt.auth)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBoltExecutor_resolveBaseURL(t *testing.T) {
	exec := NewBoltExecutor(&config.Config{})
	
	tests := []struct {
		name     string
		auth     *cliproxyauth.Auth
		expected string
	}{
		{
			name:     "nil auth returns default",
			auth:     nil,
			expected: boltBaseURL,
		},
		{
			name: "custom base url in attributes",
			auth: &cliproxyauth.Auth{
				Attributes: map[string]string{
					"base_url": "https://custom.api.com",
				},
			},
			expected: "https://custom.api.com",
		},
		{
			name: "strips trailing slash",
			auth: &cliproxyauth.Auth{
				Attributes: map[string]string{
					"base_url": "https://custom.api.com/",
				},
			},
			expected: "https://custom.api.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.resolveBaseURL(tt.auth)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBoltExecutor_resolveModel(t *testing.T) {
	exec := NewBoltExecutor(&config.Config{})
	
	tests := []struct {
		name      string
		modelName string
		auth      *cliproxyauth.Auth
		expected  string
	}{
		{
			name:      "empty model returns default",
			modelName: "",
			auth:      nil,
			expected:  boltDefaultModel,
		},
		{
			name:      "auto returns default",
			modelName: "auto",
			auth:      nil,
			expected:  boltDefaultModel,
		},
		{
			name:      "specific model passed through",
			modelName: "claude-opus-4-5",
			auth:      nil,
			expected:  "claude-opus-4-5",
		},
		{
			name:      "model override in auth attributes",
			modelName: "",
			auth: &cliproxyauth.Auth{
				Attributes: map[string]string{
					"model": "claude-haiku-4-5",
				},
			},
			expected: "claude-haiku-4-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exec.resolveModel(tt.modelName, tt.auth)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBoltWebContainerPrompt(t *testing.T) {
	// Verify the Bolt prompt contains key elements
	if len(boltWebContainerPrompt) < 1000 {
		t.Error("Bolt WebContainer prompt seems too short")
	}
	
	requiredElements := []string{
		"You are Bolt",
		"WebContainer",
		"boltArtifact",
		"boltAction",
		"Vite",
		"npm",
		"<boltAction type=\"file\"",
		"<boltAction type=\"shell\"",
		"<boltAction type=\"start\"",
	}
	
	for _, element := range requiredElements {
		if !containsSubstring(boltWebContainerPrompt, element) {
			t.Errorf("Bolt prompt missing required element: %s", element)
		}
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
