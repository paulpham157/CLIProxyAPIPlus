package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   struct {
		Version   string `json:"version"`
		Commit    string `json:"commit"`
		BuildDate string `json:"build_date"`
	} `json:"version"`
	Uptime struct {
		Seconds       int64  `json:"seconds"`
		HumanReadable string `json:"human_readable"`
	} `json:"uptime"`
	Providers struct {
		Total            int               `json:"total"`
		Active           int               `json:"active"`
		Error            int               `json:"error"`
		Disabled         int               `json:"disabled"`
		Unavailable      int               `json:"unavailable"`
		ByProvider       map[string]int    `json:"by_provider"`
		TokensValid      map[string]bool   `json:"tokens_valid"`
		ConnectionStatus map[string]string `json:"connection_status"`
	} `json:"providers"`
	Metrics struct {
		Requests struct {
			Total       int64   `json:"total"`
			Success     int64   `json:"success"`
			Failed      int64   `json:"failed"`
			SuccessRate float64 `json:"success_rate"`
		} `json:"requests"`
		Tokens struct {
			Total int64 `json:"total"`
		} `json:"tokens"`
	} `json:"metrics"`
	System struct {
		GoVersion     string `json:"go_version"`
		NumGoroutines int    `json:"num_goroutines"`
		MemoryUsageMB uint64 `json:"memory_usage_mb"`
	} `json:"system"`
}

func main() {
	baseURL := os.Getenv("CLIPROXY_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	healthURL := baseURL + "/v0/health"

	fmt.Printf("Checking health at %s\n\n", healthURL)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(healthURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to connect to health endpoint: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: Health endpoint returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read response body: %v\n", err)
		os.Exit(1)
	}

	var health HealthResponse
	if err := json.Unmarshal(body, &health); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to parse JSON response: %v\n", err)
		os.Exit(1)
	}

	printHealthSummary(health)

	if health.Status != "healthy" {
		os.Exit(1)
	}
}

func printHealthSummary(health HealthResponse) {
	fmt.Println("=== Service Health ===")
	fmt.Printf("Status: %s\n", formatStatus(health.Status))
	fmt.Printf("Version: %s (commit: %s)\n", health.Version.Version, health.Version.Commit[:7])
	fmt.Printf("Uptime: %s\n\n", health.Uptime.HumanReadable)

	fmt.Println("=== Providers ===")
	fmt.Printf("Total: %d | Active: %d | Error: %d | Disabled: %d | Unavailable: %d\n\n",
		health.Providers.Total,
		health.Providers.Active,
		health.Providers.Error,
		health.Providers.Disabled,
		health.Providers.Unavailable,
	)

	if len(health.Providers.ByProvider) > 0 {
		fmt.Println("Provider Details:")
		for provider, count := range health.Providers.ByProvider {
			status := health.Providers.ConnectionStatus[provider]
			valid := health.Providers.TokensValid[provider]
			validStr := "✗"
			if valid {
				validStr = "✓"
			}
			fmt.Printf("  %s: %d credential(s) | Status: %s | Token Valid: %s\n",
				provider, count, status, validStr)
		}
		fmt.Println()
	}

	fmt.Println("=== Metrics ===")
	fmt.Printf("Requests:\n")
	fmt.Printf("  Total: %d\n", health.Metrics.Requests.Total)
	fmt.Printf("  Success: %d\n", health.Metrics.Requests.Success)
	fmt.Printf("  Failed: %d\n", health.Metrics.Requests.Failed)
	fmt.Printf("  Success Rate: %.2f%%\n", health.Metrics.Requests.SuccessRate)
	fmt.Printf("\nTokens:\n")
	fmt.Printf("  Total Consumed: %s\n", formatNumber(health.Metrics.Tokens.Total))
	fmt.Println()

	fmt.Println("=== System ===")
	fmt.Printf("Go Version: %s\n", health.System.GoVersion)
	fmt.Printf("Goroutines: %d\n", health.System.NumGoroutines)
	fmt.Printf("Memory Usage: %d MB\n", health.System.MemoryUsageMB)
}

func formatStatus(status string) string {
	switch status {
	case "healthy":
		return "✓ Healthy"
	case "degraded":
		return "⚠ Degraded"
	default:
		return status
	}
}

func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	} else if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	} else if n < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	return fmt.Sprintf("%.1fB", float64(n)/1000000000)
}
