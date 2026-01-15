# CLIProxyAPI Plus Systemd Quick Reference

## Service Management

```bash
# Start/Stop/Restart
sudo systemctl start cli-proxy-api-plus
sudo systemctl stop cli-proxy-api-plus
sudo systemctl restart cli-proxy-api-plus

# Enable/Disable auto-start on boot
sudo systemctl enable cli-proxy-api-plus
sudo systemctl disable cli-proxy-api-plus

# Check status
sudo systemctl status cli-proxy-api-plus
sudo systemctl is-active cli-proxy-api-plus
sudo systemctl is-enabled cli-proxy-api-plus
```

## Log Management

```bash
# View logs (live tail)
sudo journalctl -u cli-proxy-api-plus -f

# View last 100 lines
sudo journalctl -u cli-proxy-api-plus -n 100

# View logs since specific time
sudo journalctl -u cli-proxy-api-plus --since "1 hour ago"
sudo journalctl -u cli-proxy-api-plus --since "2024-01-01 00:00:00"

# View logs with full output
sudo journalctl -u cli-proxy-api-plus -o cat

# Export logs to file
sudo journalctl -u cli-proxy-api-plus > /tmp/cliproxy-logs.txt
```

## Service Information

```bash
# Show service details
systemctl show cli-proxy-api-plus

# Show restart count
systemctl show cli-proxy-api-plus -p NRestarts

# Show memory usage
systemctl show cli-proxy-api-plus -p MemoryCurrent

# Show CPU usage
systemctl show cli-proxy-api-plus -p CPUUsageNSec

# View service dependencies
systemctl list-dependencies cli-proxy-api-plus
```

## Configuration Files

```bash
# Main config
/opt/cli-proxy-api-plus/config.yaml

# Environment variables
/etc/cli-proxy-api-plus/environment
/opt/cli-proxy-api-plus/.env

# Service unit file
/etc/systemd/system/cli-proxy-api-plus.service

# Data directory
/var/lib/cli-proxy-api-plus/

# Log directory
/var/log/cli-proxy-api-plus/
```

## Editing Service

```bash
# Edit service file
sudo systemctl edit --full cli-proxy-api-plus

# Or manually edit and reload
sudo nano /etc/systemd/system/cli-proxy-api-plus.service
sudo systemctl daemon-reload
sudo systemctl restart cli-proxy-api-plus
```

## Resource Monitoring

```bash
# Real-time resource usage
systemd-cgtop

# Service-specific cgroup stats
cat /sys/fs/cgroup/system.slice/cli-proxy-api-plus.service/memory.current
cat /sys/fs/cgroup/system.slice/cli-proxy-api-plus.service/cpu.stat
```

## Troubleshooting

```bash
# Check for failed services
systemctl --failed

# View detailed error logs
sudo journalctl -u cli-proxy-api-plus -p err -n 50

# Verify configuration syntax
sudo systemd-analyze verify /etc/systemd/system/cli-proxy-api-plus.service

# Check service file for issues
sudo systemctl cat cli-proxy-api-plus

# Clear journal logs (if too large)
sudo journalctl --rotate
sudo journalctl --vacuum-time=7d
sudo journalctl --vacuum-size=100M
```

## Maintenance Timer (Optional)

```bash
# Enable maintenance timer
sudo systemctl enable cli-proxy-api-plus-maintenance.timer
sudo systemctl start cli-proxy-api-plus-maintenance.timer

# Check timer status
systemctl status cli-proxy-api-plus-maintenance.timer

# List all timers
systemctl list-timers

# View next scheduled run
systemctl list-timers cli-proxy-api-plus-maintenance.timer

# Manually trigger maintenance task
sudo systemctl start cli-proxy-api-plus-maintenance.service

# View maintenance logs
sudo journalctl -u cli-proxy-api-plus-maintenance.service
```

## Performance Tuning

```bash
# Adjust memory limits (edit service file)
MemoryMax=4G
MemoryHigh=3G

# Adjust CPU limits (400% = 4 cores)
CPUQuota=400%

# Adjust file descriptor limits
LimitNOFILE=131072

# Apply changes
sudo systemctl daemon-reload
sudo systemctl restart cli-proxy-api-plus
```

## Backup and Restore

```bash
# Backup configuration
sudo tar -czf cliproxy-backup-$(date +%Y%m%d).tar.gz \
  /opt/cli-proxy-api-plus/config.yaml \
  /etc/cli-proxy-api-plus/ \
  /var/lib/cli-proxy-api-plus/

# Restore configuration
sudo tar -xzf cliproxy-backup-YYYYMMDD.tar.gz -C /

# Restart after restore
sudo systemctl restart cli-proxy-api-plus
```

## Health Checks

```bash
# Check if service is running
systemctl is-active cli-proxy-api-plus && echo "Running" || echo "Not running"

# Check service uptime
systemctl show cli-proxy-api-plus -p ActiveEnterTimestamp

# Check if service is enabled
systemctl is-enabled cli-proxy-api-plus

# Test API endpoint (adjust port if different)
curl http://localhost:8317/v1/models
```

## Security

```bash
# View service permissions
sudo systemctl show cli-proxy-api-plus | grep -E 'User|Group|ReadWrite|Protect'

# Check file permissions
ls -la /opt/cli-proxy-api-plus/
ls -la /var/lib/cli-proxy-api-plus/
ls -la /etc/cli-proxy-api-plus/

# Audit security settings
sudo systemd-analyze security cli-proxy-api-plus
```

## Updates

```bash
# Update application
sudo systemctl stop cli-proxy-api-plus
sudo cp cli-proxy-api-plus /opt/cli-proxy-api-plus/
sudo chown cliproxy:cliproxy /opt/cli-proxy-api-plus/cli-proxy-api-plus
sudo systemctl start cli-proxy-api-plus

# Verify update
systemctl status cli-proxy-api-plus
sudo journalctl -u cli-proxy-api-plus -n 20
```

## Emergency Operations

```bash
# Force stop (SIGKILL)
sudo systemctl kill -s SIGKILL cli-proxy-api-plus

# Disable restart on failure temporarily
sudo systemctl set-property cli-proxy-api-plus Restart=no

# Reset restart counter
sudo systemctl reset-failed cli-proxy-api-plus

# Mask service (prevent any start)
sudo systemctl mask cli-proxy-api-plus

# Unmask service
sudo systemctl unmask cli-proxy-api-plus
```
