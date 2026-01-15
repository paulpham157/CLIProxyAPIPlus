# Health Check Example

This example demonstrates how to use the CLI Proxy API health check endpoint (`/v0/health`) to monitor the service status.

## Usage

Run the example:

```bash
go run main.go
```

Or set a custom URL:

```bash
CLIPROXY_URL=http://your-server:8080 go run main.go
```

## Sample Output

```
Checking health at http://localhost:8080/v0/health

=== Service Health ===
Status: ✓ Healthy
Version: v1.2.3 (commit: abc123d)
Uptime: 2 hours 15 minutes

=== Providers ===
Total: 4 | Active: 3 | Error: 0 | Disabled: 1 | Unavailable: 0

Provider Details:
  gemini: 2 credential(s) | Status: connected | Token Valid: ✓
  claude: 1 credential(s) | Status: connected | Token Valid: ✓
  copilot: 1 credential(s) | Status: disabled | Token Valid: ✗

=== Metrics ===
Requests:
  Total: 1543
  Success: 1521
  Failed: 22
  Success Rate: 98.57%

Tokens:
  Total Consumed: 543.2K

=== System ===
Go Version: go1.24.0
Goroutines: 42
Memory Usage: 85 MB
```

## Exit Codes

- `0`: Service is healthy
- `1`: Service is degraded or unreachable

## Integration Example

Use this in scripts for automated health checks:

```bash
#!/bin/bash
# health-check.sh

if go run main.go > /dev/null 2>&1; then
    echo "Service is healthy"
    exit 0
else
    echo "Service is degraded or down"
    exit 1
fi
```

Or build once and use the binary:

```bash
go build -o health-check main.go
./health-check
```
