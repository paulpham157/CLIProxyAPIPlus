# Systemd Deployment Guide for CLIProxyAPI Plus

This document provides an overview of the systemd deployment option for CLIProxyAPI Plus, which serves as a production-ready alternative to PM2 for Linux servers.

## Overview

The systemd deployment provides a robust, secure, and efficient way to run CLIProxyAPI Plus as a system service with the following benefits:

- ✅ **Automatic restart on failure** with configurable retry limits
- ✅ **Resource limits** for CPU, memory, and file descriptors
- ✅ **Security hardening** with sandboxing and privilege restrictions
- ✅ **Boot-time startup** with proper dependency management
- ✅ **Centralized logging** via systemd journal
- ✅ **No additional dependencies** (no Node.js or PM2 required)

## Quick Start

### 1. Build the Application

```bash
go build -o cli-proxy-api-plus ./cmd/server
```

### 2. Run the Installation Script

```bash
sudo ./systemd/install.sh
```

The script will:
- Create a dedicated `cliproxy` system user
- Set up directory structure (`/opt`, `/var/lib`, `/var/log`)
- Install the binary and configuration files
- Configure and start the systemd service
- Enable automatic startup on boot

### 3. Verify Installation

```bash
# Check service status
sudo systemctl status cli-proxy-api-plus

# View logs
sudo journalctl -u cli-proxy-api-plus -f
```

## Files and Directories

### Service Files

- **`cli-proxy-api-plus.service`** (repo root): Main systemd service unit file
- **`systemd/`**: Directory containing all systemd-related configuration

### Systemd Directory Contents

- **`INSTALL.md`**: Detailed installation and configuration guide
- **`QUICKREF.md`**: Quick reference for common operations
- **`README.md`**: Overview of systemd deployment features
- **`environment.example`**: Example environment variable configuration
- **`install.sh`**: Automated installation script
- **`cli-proxy-api-plus-maintenance.service`**: Optional maintenance task service
- **`cli-proxy-api-plus-maintenance.timer`**: Optional systemd timer for scheduled tasks

### Production Directories

After installation, the following directories are created:

- **`/opt/cli-proxy-api-plus/`**: Application binary and config.yaml
- **`/var/lib/cli-proxy-api-plus/`**: Data directory (tokens, auth, storage)
- **`/var/log/cli-proxy-api-plus/`**: Log directory (when file logging enabled)
- **`/etc/cli-proxy-api-plus/`**: System-wide environment configuration

## Key Features

### Automatic Restart

The service automatically restarts on failure with the following configuration:

```ini
Restart=on-failure
RestartSec=5s
StartLimitInterval=300
StartLimitBurst=5
```

This means:
- Restarts automatically if the process fails
- Waits 5 seconds between restart attempts
- Allows up to 5 restarts within 300 seconds (5 minutes)
- Stops trying if limit is exceeded (prevents restart loops)

### Resource Limits

Default resource limits (customizable in service file):

```ini
MemoryMax=2G          # Maximum memory usage
MemoryHigh=1.5G       # Soft limit (triggers throttling)
CPUQuota=200%         # CPU limit (200% = 2 cores)
LimitNOFILE=65536     # File descriptor limit
LimitNPROC=512        # Process/thread limit
TasksMax=512          # Task limit
```

### Security Hardening

The service runs with multiple security features enabled:

- **Non-root user**: Runs as dedicated `cliproxy` system user
- **Private tmp**: Isolated `/tmp` namespace
- **Read-only system**: System directories are read-only
- **Protected home**: Home directory is inaccessible
- **Restricted syscalls**: Limited system call access
- **No new privileges**: Cannot gain additional privileges
- **No SUID/SGID**: Cannot execute privileged binaries

### Boot-time Startup

The service is automatically started on system boot through:

```ini
[Install]
WantedBy=multi-user.target
```

Enable/disable with:
```bash
sudo systemctl enable cli-proxy-api-plus   # Enable auto-start
sudo systemctl disable cli-proxy-api-plus  # Disable auto-start
```

## Storage Backend Support

The systemd deployment fully supports all CLIProxyAPI storage backends:

### File-Based Storage (Default)

No additional configuration needed. Data is stored in `/var/lib/cli-proxy-api-plus`.

### PostgreSQL Storage

For production deployments with multiple instances, add to `/etc/cli-proxy-api-plus/environment`:

```bash
PGSTORE_DSN=postgresql://user:pass@localhost:5432/cliproxy
PGSTORE_SCHEMA=public
PGSTORE_LOCAL_PATH=/var/lib/cli-proxy-api-plus/pgstore
```

### Git-Backed Storage

For GitOps workflows, add to `/etc/cli-proxy-api-plus/environment`:

```bash
GITSTORE_GIT_URL=https://github.com/your-org/cli-proxy-config.git
GITSTORE_GIT_USERNAME=git-user
GITSTORE_GIT_TOKEN=ghp_your_personal_access_token
GITSTORE_LOCAL_PATH=/var/lib/cli-proxy-api-plus/gitstore
```

### Object Store (S3-Compatible)

For S3, MinIO, or other compatible storage, add to `/etc/cli-proxy-api-plus/environment`:

```bash
OBJECTSTORE_ENDPOINT=https://s3.your-cloud.example.com
OBJECTSTORE_BUCKET=cli-proxy-config
OBJECTSTORE_ACCESS_KEY=your_access_key
OBJECTSTORE_SECRET_KEY=your_secret_key
OBJECTSTORE_LOCAL_PATH=/var/lib/cli-proxy-api-plus/objectstore
```

## Common Operations

```bash
# Service control
sudo systemctl start cli-proxy-api-plus
sudo systemctl stop cli-proxy-api-plus
sudo systemctl restart cli-proxy-api-plus

# Status and logs
sudo systemctl status cli-proxy-api-plus
sudo journalctl -u cli-proxy-api-plus -f

# Enable/disable auto-start
sudo systemctl enable cli-proxy-api-plus
sudo systemctl disable cli-proxy-api-plus

# Resource monitoring
systemctl show cli-proxy-api-plus -p MemoryCurrent
systemctl show cli-proxy-api-plus -p NRestarts
```

## Configuration

### Application Configuration

Edit `/opt/cli-proxy-api-plus/config.yaml` to configure:

- Server port and host binding
- API keys
- Storage backend preferences
- Logging options
- Security settings

See `config.example.yaml` for all available options.

### Environment Variables

Add environment variables to `/etc/cli-proxy-api-plus/environment`:

```bash
# Example: Configure PostgreSQL store
PGSTORE_DSN=postgresql://localhost/cliproxy
```

After editing environment files, restart the service:

```bash
sudo systemctl restart cli-proxy-api-plus
```

### Resource Limits

Edit `/etc/systemd/system/cli-proxy-api-plus.service` and adjust limits:

```ini
[Service]
MemoryMax=4G
CPUQuota=400%
LimitNOFILE=131072
```

Apply changes:

```bash
sudo systemctl daemon-reload
sudo systemctl restart cli-proxy-api-plus
```

## Advantages Over PM2

| Feature | Systemd | PM2 |
|---------|---------|-----|
| **Native Integration** | Built into Linux | Requires Node.js |
| **Resource Limits** | cgroups (kernel-level) | Process-level only |
| **Security Sandboxing** | Extensive options | Limited |
| **Centralized Logging** | systemd journal | Custom log files |
| **Boot Integration** | Native systemd targets | Init scripts |
| **Dependency Management** | Full systemd support | Limited |
| **Process Supervision** | Kernel-integrated | Userspace daemon |
| **Memory Overhead** | Minimal | Node.js + PM2 |

## Monitoring and Alerting

### Built-in Monitoring

```bash
# Check service health
systemctl is-active cli-proxy-api-plus

# View restart count
systemctl show cli-proxy-api-plus -p NRestarts

# Monitor resources
systemd-cgtop
```

### Integration with Monitoring Tools

Systemd integrates with various monitoring solutions:

- **Prometheus**: Use `node_exporter` with systemd collector
- **Datadog**: Systemd integration available
- **Grafana**: Via Prometheus or Loki
- **ELK Stack**: Via journalbeat or filebeat

### Log Aggregation

Export logs to external systems:

```bash
# Export to file
sudo journalctl -u cli-proxy-api-plus > /tmp/logs.txt

# Stream to syslog
# Configure in systemd service:
StandardOutput=syslog
StandardError=syslog
```

## Updating the Application

```bash
# Build new version
go build -o cli-proxy-api-plus ./cmd/server

# Stop service
sudo systemctl stop cli-proxy-api-plus

# Backup current binary
sudo cp /opt/cli-proxy-api-plus/cli-proxy-api-plus \
        /opt/cli-proxy-api-plus/cli-proxy-api-plus.backup

# Install new binary
sudo cp cli-proxy-api-plus /opt/cli-proxy-api-plus/
sudo chown cliproxy:cliproxy /opt/cli-proxy-api-plus/cli-proxy-api-plus

# Start service
sudo systemctl start cli-proxy-api-plus

# Verify
sudo systemctl status cli-proxy-api-plus
sudo journalctl -u cli-proxy-api-plus -n 20
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs for errors
sudo journalctl -u cli-proxy-api-plus -n 50

# Verify configuration
sudo systemctl cat cli-proxy-api-plus

# Check file permissions
ls -la /opt/cli-proxy-api-plus/
ls -la /var/lib/cli-proxy-api-plus/
```

### High Resource Usage

```bash
# Check current usage
systemctl show cli-proxy-api-plus -p MemoryCurrent
systemctl show cli-proxy-api-plus -p CPUUsageNSec

# Adjust limits if needed
sudo systemctl edit --full cli-proxy-api-plus
```

### Service Keeps Restarting

```bash
# View restart count and recent failures
systemctl show cli-proxy-api-plus -p NRestarts
sudo journalctl -u cli-proxy-api-plus -p err -n 50

# Reset restart counter
sudo systemctl reset-failed cli-proxy-api-plus
```

For more troubleshooting steps, see `systemd/INSTALL.md`.

## Documentation

- **[systemd/INSTALL.md](systemd/INSTALL.md)**: Complete installation guide
- **[systemd/QUICKREF.md](systemd/QUICKREF.md)**: Quick reference for common operations
- **[systemd/README.md](systemd/README.md)**: Overview of systemd deployment

## Migration from PM2

If you're currently using PM2:

1. **Export PM2 configuration**:
   ```bash
   pm2 env <app-id> > pm2-env.txt
   ```

2. **Convert to systemd environment file**:
   Review and convert PM2 environment variables to systemd format.

3. **Stop PM2 service**:
   ```bash
   pm2 stop cli-proxy-api-plus
   pm2 delete cli-proxy-api-plus
   pm2 save
   ```

4. **Install systemd service**:
   ```bash
   sudo ./systemd/install.sh
   ```

5. **Verify migration**:
   ```bash
   sudo systemctl status cli-proxy-api-plus
   sudo journalctl -u cli-proxy-api-plus -f
   ```

## Support

For issues with systemd deployment:

1. Check service logs: `sudo journalctl -u cli-proxy-api-plus -n 100`
2. Verify service status: `sudo systemctl status cli-proxy-api-plus`
3. Review installation guide: `systemd/INSTALL.md`
4. Check quick reference: `systemd/QUICKREF.md`

For general CLIProxyAPI issues, refer to the main [README.md](README.md).

## License

This systemd deployment is part of CLIProxyAPI Plus and follows the same MIT License. See [LICENSE](LICENSE) for details.
