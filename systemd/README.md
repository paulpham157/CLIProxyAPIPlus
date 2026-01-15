# Systemd Service for CLIProxyAPI Plus

This directory contains systemd service unit files and related configuration for running CLIProxyAPI Plus as a production service on Linux servers.

## Contents

- **`cli-proxy-api-plus.service`** (in repo root): Main systemd service unit file
- **`environment.example`**: Example environment variable configuration
- **`INSTALL.md`**: Complete installation and configuration guide
- **`install.sh`**: Quick installation script
- **`cli-proxy-api-plus-maintenance.service`**: Optional maintenance service unit (for scheduled tasks)
- **`cli-proxy-api-plus-maintenance.timer`**: Optional systemd timer for maintenance tasks

## Features

- ✅ **Automatic restart on failure** with exponential backoff
- ✅ **Resource limits** (CPU, memory, file descriptors)
- ✅ **Security hardening** (sandboxing, restricted permissions)
- ✅ **Boot-time startup** enabled by default
- ✅ **Systemd journal integration** for centralized logging
- ✅ **Graceful shutdown** with configurable timeout
- ✅ **Support for all storage backends** (file, PostgreSQL, Git, Object Store)

## Quick Start

For a standard installation:

```bash
# Run the installation script
sudo ./systemd/install.sh

# Or follow the manual installation guide
cat systemd/INSTALL.md
```

## Service Configuration

The service runs with the following characteristics:

- **User/Group**: `cliproxy` (non-root, system user)
- **Working Directory**: `/opt/cli-proxy-api-plus`
- **Data Directory**: `/var/lib/cli-proxy-api-plus`
- **Log Directory**: `/var/log/cli-proxy-api-plus`
- **Config Directory**: `/etc/cli-proxy-api-plus`

## Resource Limits (Default)

- **Memory**: 2GB max, 1.5GB high watermark
- **CPU**: 200% (2 cores)
- **File Descriptors**: 65,536
- **Processes**: 512

These can be customized by editing the service file.

## Security Features

The service includes multiple security hardening measures:

- Runs as unprivileged system user
- Private `/tmp` namespace
- Read-only system directories
- No new privileges
- Restricted system calls
- No SUID/SGID execution
- Locked personality
- Protected kernel tunables/modules

## Storage Backend Support

The service supports all CLIProxyAPI storage backends:

### File-Based (Default)
- No additional configuration needed
- Data stored in `/var/lib/cli-proxy-api-plus`

### PostgreSQL
- Set `PGSTORE_DSN` in environment file
- Recommended for production multi-instance deployments

### Git-Backed
- Set `GITSTORE_GIT_URL` and credentials in environment file
- Useful for GitOps workflows

### Object Store (S3-Compatible)
- Set `OBJECTSTORE_ENDPOINT` and credentials in environment file
- Compatible with MinIO, AWS S3, and other S3-compatible services

## Environment Variables

Environment variables can be set in two locations:

1. **System-wide**: `/etc/cli-proxy-api-plus/environment` (preferred for production)
2. **Application directory**: `/opt/cli-proxy-api-plus/.env`

See `environment.example` for available options.

## Common Operations

```bash
# Start service
sudo systemctl start cli-proxy-api-plus

# Stop service
sudo systemctl stop cli-proxy-api-plus

# Restart service
sudo systemctl restart cli-proxy-api-plus

# View status
sudo systemctl status cli-proxy-api-plus

# View logs
sudo journalctl -u cli-proxy-api-plus -f

# Enable auto-start on boot
sudo systemctl enable cli-proxy-api-plus

# Disable auto-start on boot
sudo systemctl disable cli-proxy-api-plus
```

## Monitoring

View service metrics:

```bash
# Check if service is active
sudo systemctl is-active cli-proxy-api-plus

# View restart count
systemctl show cli-proxy-api-plus -p NRestarts

# View memory usage
systemctl show cli-proxy-api-plus -p MemoryCurrent

# View full service status
sudo systemctl status cli-proxy-api-plus
```

## Advantages Over PM2

1. **Native Integration**: Deep integration with Linux system management
2. **Resource Control**: Built-in cgroups support for CPU/memory limits
3. **Security**: Advanced sandboxing and security features
4. **Logging**: Centralized journal with structured logging
5. **Dependencies**: Proper service dependency management
6. **Reliability**: Battle-tested process supervision
7. **No Additional Dependencies**: No need for Node.js/PM2

## Documentation

- **[INSTALL.md](INSTALL.md)**: Complete installation guide with troubleshooting
- **[QUICKREF.md](QUICKREF.md)**: Quick reference for common systemd operations
- **[environment.example](environment.example)**: Environment variable examples

## Support

For issues specific to systemd deployment, check:

1. Service logs: `sudo journalctl -u cli-proxy-api-plus -n 100`
2. Service status: `sudo systemctl status cli-proxy-api-plus`
3. Permissions: Verify ownership of `/var/lib/cli-proxy-api-plus` and `/opt/cli-proxy-api-plus`

For application issues, refer to the main CLIProxyAPI documentation.
