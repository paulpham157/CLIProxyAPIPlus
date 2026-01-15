package handlers

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/buildinfo"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/usage"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/api/handlers"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

var serverStartTime = time.Now()

type HealthHandler struct {
	baseHandler *handlers.BaseAPIHandler
}

func NewHealthHandler(baseHandler *handlers.BaseAPIHandler) *HealthHandler {
	return &HealthHandler{
		baseHandler: baseHandler,
	}
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   VersionInfo            `json:"version"`
	Uptime    UptimeInfo             `json:"uptime"`
	Providers ProviderHealthSummary  `json:"providers"`
	Metrics   MetricsInfo            `json:"metrics"`
	System    SystemInfo             `json:"system"`
}

type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
}

type UptimeInfo struct {
	Seconds      int64  `json:"seconds"`
	HumanReadable string `json:"human_readable"`
}

type ProviderHealthSummary struct {
	Total           int               `json:"total"`
	Active          int               `json:"active"`
	Error           int               `json:"error"`
	Disabled        int               `json:"disabled"`
	Unavailable     int               `json:"unavailable"`
	ByProvider      map[string]int    `json:"by_provider"`
	TokensValid     map[string]bool   `json:"tokens_valid"`
	ConnectionStatus map[string]string `json:"connection_status"`
}

type MetricsInfo struct {
	Requests RequestMetrics `json:"requests"`
	Tokens   TokenMetrics   `json:"tokens"`
}

type RequestMetrics struct {
	Total       int64   `json:"total"`
	Success     int64   `json:"success"`
	Failed      int64   `json:"failed"`
	SuccessRate float64 `json:"success_rate"`
}

type TokenMetrics struct {
	Total int64 `json:"total"`
}

type SystemInfo struct {
	GoVersion     string `json:"go_version"`
	NumGoroutines int    `json:"num_goroutines"`
	MemoryUsageMB uint64 `json:"memory_usage_mb"`
}

func (h *HealthHandler) GetHealth(c *gin.Context) {
	now := time.Now()
	uptime := now.Sub(serverStartTime)

	authManager := h.getAuthManager()
	providerHealth := h.computeProviderHealth(authManager)

	stats := usage.GetRequestStatistics().Snapshot()
	successRate := 0.0
	if stats.TotalRequests > 0 {
		successRate = float64(stats.SuccessCount) / float64(stats.TotalRequests) * 100
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: now.UTC().Format(time.RFC3339),
		Version: VersionInfo{
			Version:   buildinfo.Version,
			Commit:    buildinfo.Commit,
			BuildDate: buildinfo.BuildDate,
		},
		Uptime: UptimeInfo{
			Seconds:      int64(uptime.Seconds()),
			HumanReadable: formatUptime(uptime),
		},
		Providers: providerHealth,
		Metrics: MetricsInfo{
			Requests: RequestMetrics{
				Total:       stats.TotalRequests,
				Success:     stats.SuccessCount,
				Failed:      stats.FailureCount,
				SuccessRate: successRate,
			},
			Tokens: TokenMetrics{
				Total: stats.TotalTokens,
			},
		},
		System: SystemInfo{
			GoVersion:     runtime.Version(),
			NumGoroutines: runtime.NumGoroutine(),
			MemoryUsageMB: memStats.Alloc / 1024 / 1024,
		},
	}

	if providerHealth.Error > 0 || providerHealth.Total == 0 {
		response.Status = "degraded"
	}

	c.JSON(http.StatusOK, response)
}

func (h *HealthHandler) getAuthManager() *auth.Manager {
	if h.baseHandler == nil || h.baseHandler.AuthManager == nil {
		return nil
	}
	return h.baseHandler.AuthManager
}

func (h *HealthHandler) computeProviderHealth(authManager *auth.Manager) ProviderHealthSummary {
	summary := ProviderHealthSummary{
		ByProvider:       make(map[string]int),
		TokensValid:      make(map[string]bool),
		ConnectionStatus: make(map[string]string),
	}

	if authManager == nil {
		return summary
	}

	auths := authManager.List()
	summary.Total = len(auths)

	providerStatus := make(map[string]string)
	providerValid := make(map[string]bool)

	for _, a := range auths {
		if a == nil {
			continue
		}

		summary.ByProvider[a.Provider]++

		if a.Disabled {
			summary.Disabled++
			providerStatus[a.Provider] = "disabled"
			providerValid[a.Provider] = false
			continue
		}

		if a.Status == auth.StatusError {
			summary.Error++
			providerStatus[a.Provider] = "error"
			providerValid[a.Provider] = false
		} else if a.Unavailable {
			summary.Unavailable++
			providerStatus[a.Provider] = "unavailable"
			providerValid[a.Provider] = false
		} else if a.Status == auth.StatusActive {
			summary.Active++
			if _, exists := providerStatus[a.Provider]; !exists {
				providerStatus[a.Provider] = "connected"
			}
			providerValid[a.Provider] = true
		}

		if expiry, hasExpiry := a.ExpirationTime(); hasExpiry && !expiry.IsZero() {
			if time.Now().After(expiry) {
				providerValid[a.Provider] = false
				providerStatus[a.Provider] = "expired"
			}
		}
	}

	for provider, status := range providerStatus {
		summary.ConnectionStatus[provider] = status
	}

	for provider, valid := range providerValid {
		summary.TokensValid[provider] = valid
	}

	return summary
}

func formatUptime(d time.Duration) string {
	days := int64(d.Hours()) / 24
	hours := int64(d.Hours()) % 24
	minutes := int64(d.Minutes()) % 60
	seconds := int64(d.Seconds()) % 60

	parts := []string{}
	
	if days > 0 {
		parts = append(parts, formatDuration(days, "day"))
		parts = append(parts, formatDuration(hours, "hour"))
	} else if hours > 0 {
		parts = append(parts, formatDuration(hours, "hour"))
		parts = append(parts, formatDuration(minutes, "minute"))
	} else if minutes > 0 {
		parts = append(parts, formatDuration(minutes, "minute"))
		parts = append(parts, formatDuration(seconds, "second"))
	} else {
		parts = append(parts, formatDuration(seconds, "second"))
	}
	
	result := []string{}
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	
	return strings.Join(result, " ")
}

func formatDuration(value int64, unit string) string {
	if value == 0 {
		return ""
	}
	if value == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", value, unit)
}
