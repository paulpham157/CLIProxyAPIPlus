# Systemd Installation Guide for CLIProxyAPI Plus

This guide explains how to install and configure CLIProxyAPI Plus as a systemd service for production Linux servers.

## Prerequisites

- Linux system with systemd (most modern distributions)
- Root or sudo access
- Built CLIProxyAPI Plus binary

## Installation Steps

### 1. Build the Application

```bash
go build -o cli-proxy-api-plus ./cmd/server
```

### 2. Create System User

Create a dedicated user for running the service:

```bash
sudo useradd --system --no-create-home --shell /bin/false cliproxy
```

### 3. Create Directory Structure

```bash
# Application directory
sudo mkdir -p /opt/cli-proxy-api-plus

# Data directory (for tokens, config, logs)
sudo mkdir -p /var/lib/cli-proxy-api-plus

# Log directory
sudo mkdir -p /var/log/cli-proxy-api-plus

# Configuration directory
sudo mkdir -p /etc/cli-proxy-api-plus
```

### 4. Install Application Files

```bash
# Copy binary
sudo cp cli-proxy-api-plus /opt/cli-proxy-api-plus/

# Copy configuration files
sudo cp config.example.yaml /opt/cli-proxy-api-plus/config.yaml
# or if you have a custom config:
# sudo cp config.yaml /opt/cli-proxy-api-plus/

# Set ownership
sudo chown -R cliproxy:cliproxy /opt/cli-proxy-api-plus
sudo chown -R cliproxy:cliproxy /var/lib/cli-proxy-api-plus
sudo chown -R cliproxy:cliproxy /var/log/cli-proxy-api-plus
```

### 5. Configure Environment Variables (Optional)

If you need environment variables (e.g., for Postgres, Git, or Object Store):

```bash
# Copy and edit environment file
sudo cp systemd/environment.example /etc/cli-proxy-api-plus/environment
sudo nano /etc/cli-proxy-api-plus/environment

# Secure the file (may contain secrets)
sudo chmod 600 /etc/cli-proxy-api-plus/environment
sudo chown root:root /etc/cli-proxy-api-plus/environment
```

Alternatively, you can place a `.env` file in the working directory:

```bash
sudo cp .env.example /opt/cli-proxy-api-plus/.env
sudo nano /opt/cli-proxy-api-plus/.env
sudo chown cliproxy:cliproxy /opt/cli-proxy-api-plus/.env
sudo chmod 600 /opt/cli-proxy-api-plus/.env
```

### 6. Configure the Application

Edit the configuration file:

```bash
sudo nano /opt/cli-proxy-api-plus/config.yaml
```

Key configuration options:
- `host`: Set to `""` for all interfaces or `"127.0.0.1"` for localhost only
- `port`: Default is `8317`
- `auth-dir`: Set to `/var/lib/cli-proxy-api-plus` for production
- `api-keys`: Set your API keys
- `logging-to-file`: Set to `true` for file-based logging
- `remote-management.secret-key`: Set a strong management password

For file-based logging, update `auth-dir` in config:

```yaml
auth-dir: "/var/lib/cli-proxy-api-plus"
logging-to-file: true
```

### 7. Install Systemd Service

```bash
# Copy service file
sudo cp cli-proxy-api-plus.service /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload
```

### 8. Enable and Start Service

```bash
# Enable service to start on boot
sudo systemctl enable cli-proxy-api-plus

# Start the service
sudo systemctl start cli-proxy-api-plus

# Check status
sudo systemctl status cli-proxy-api-plus
```

## Service Management

### View Logs

```bash
# Follow live logs
sudo journalctl -u cli-proxy-api-plus -f

# View recent logs
sudo journalctl -u cli-proxy-api-plus -n 100

# View logs with timestamps
sudo journalctl -u cli-proxy-api-plus -o short-iso
```

### Control Service

```bash
# Stop service
sudo systemctl stop cli-proxy-api-plus

# Restart service
sudo systemctl restart cli-proxy-api-plus

# Reload configuration (if supported)
sudo systemctl reload cli-proxy-api-plus

# Check if service is running
sudo systemctl is-active cli-proxy-api-plus

# Disable auto-start on boot
sudo systemctl disable cli-proxy-api-plus
```

### Monitor Service Health

```bash
# Check service status
sudo systemctl status cli-proxy-api-plus

# View service dependencies
systemctl list-dependencies cli-proxy-api-plus

# Check restart history
sudo journalctl -u cli-proxy-api-plus | grep "Started\|Stopped"
```

## Configuration Options

### Customizing Resource Limits

Edit `/etc/systemd/system/cli-proxy-api-plus.service` and adjust:

```ini
[Service]
# Memory limits
MemoryMax=4G          # Maximum memory
MemoryHigh=3G         # Soft limit before throttling

# CPU limits
CPUQuota=400%         # 400% = 4 CPU cores

# File descriptor limit
LimitNOFILE=131072    # Increase if handling many connections

# Process limit
LimitNPROC=1024
TasksMax=1024
```

After editing, reload systemd:

```bash
sudo systemctl daemon-reload
sudo systemctl restart cli-proxy-api-plus
```

### Using Different Storage Backends

#### PostgreSQL Store

1. Edit environment file:

```bash
sudo nano /etc/cli-proxy-api-plus/environment
```

Add:

```bash
PGSTORE_DSN=postgresql://user:pass@localhost:5432/cliproxy
PGSTORE_SCHEMA=public
PGSTORE_LOCAL_PATH=/var/lib/cli-proxy-api-plus/pgstore
```

2. Create PostgreSQL database:

```sql
CREATE DATABASE cliproxy;
```

3. Restart service:

```bash
sudo systemctl restart cli-proxy-api-plus
```

#### Git-Backed Store

1. Edit environment file:

```bash
sudo nano /etc/cli-proxy-api-plus/environment
```

Add:

```bash
GITSTORE_GIT_URL=https://github.com/your-org/cli-proxy-config.git
GITSTORE_GIT_USERNAME=git-user
GITSTORE_GIT_TOKEN=ghp_your_personal_access_token
GITSTORE_LOCAL_PATH=/var/lib/cli-proxy-api-plus/gitstore
```

2. Restart service:

```bash
sudo systemctl restart cli-proxy-api-plus
```

#### Object Store (S3-Compatible)

1. Edit environment file:

```bash
sudo nano /etc/cli-proxy-api-plus/environment
```

Add:

```bash
OBJECTSTORE_ENDPOINT=https://s3.your-cloud.example.com
OBJECTSTORE_BUCKET=cli-proxy-config
OBJECTSTORE_ACCESS_KEY=your_access_key
OBJECTSTORE_SECRET_KEY=your_secret_key
OBJECTSTORE_LOCAL_PATH=/var/lib/cli-proxy-api-plus/objectstore
```

2. Restart service:

```bash
sudo systemctl restart cli-proxy-api-plus
```

## Troubleshooting

### Service Won't Start

Check logs for errors:

```bash
sudo journalctl -u cli-proxy-api-plus -n 50
```

Common issues:
- **Permission denied**: Check file ownership and permissions
- **Port already in use**: Change port in `config.yaml` or stop conflicting service
- **Config file errors**: Validate YAML syntax in `config.yaml`

### Permission Issues

Ensure correct ownership:

```bash
sudo chown -R cliproxy:cliproxy /opt/cli-proxy-api-plus
sudo chown -R cliproxy:cliproxy /var/lib/cli-proxy-api-plus
sudo chown -R cliproxy:cliproxy /var/log/cli-proxy-api-plus
```

### Service Keeps Restarting

Check for:
- Application crashes in logs
- Configuration errors
- Resource exhaustion (memory, disk space)
- Network issues (if using remote storage)

View restart count:

```bash
systemctl show cli-proxy-api-plus | grep NRestarts
```

### High Memory Usage

Adjust memory limits in service file:

```ini
MemoryMax=1G
MemoryHigh=768M
```

### Database Connection Issues

For PostgreSQL store, verify:
- Database is running: `sudo systemctl status postgresql`
- Credentials are correct in environment file
- Network connectivity to database
- Database exists and user has permissions

## Security Considerations

1. **Run as non-root user**: Service runs as `cliproxy` user by default
2. **File permissions**: Config and environment files should be readable only by service user
3. **Firewall**: Configure firewall to allow only necessary connections
4. **TLS/HTTPS**: Enable TLS in config for production deployments
5. **API keys**: Use strong, randomly generated API keys
6. **Management API**: Restrict management access with `allow-remote: false`

## Updating the Application

```bash
# Build new version
go build -o cli-proxy-api-plus ./cmd/server

# Stop service
sudo systemctl stop cli-proxy-api-plus

# Backup current binary
sudo cp /opt/cli-proxy-api-plus/cli-proxy-api-plus /opt/cli-proxy-api-plus/cli-proxy-api-plus.backup

# Install new binary
sudo cp cli-proxy-api-plus /opt/cli-proxy-api-plus/
sudo chown cliproxy:cliproxy /opt/cli-proxy-api-plus/cli-proxy-api-plus

# Start service
sudo systemctl start cli-proxy-api-plus

# Verify
sudo systemctl status cli-proxy-api-plus
```

## Uninstalling

```bash
# Stop and disable service
sudo systemctl stop cli-proxy-api-plus
sudo systemctl disable cli-proxy-api-plus

# Remove service file
sudo rm /etc/systemd/system/cli-proxy-api-plus.service
sudo systemctl daemon-reload

# Remove application files
sudo rm -rf /opt/cli-proxy-api-plus
sudo rm -rf /etc/cli-proxy-api-plus

# Remove data (optional - backup first!)
# sudo rm -rf /var/lib/cli-proxy-api-plus
# sudo rm -rf /var/log/cli-proxy-api-plus

# Remove user (optional)
# sudo userdel cliproxy
```

## Migration from PM2

If you're currently using PM2:

1. Export PM2 environment variables:

```bash
pm2 env <app-id> > pm2-env.txt
```

2. Convert to systemd environment file format

3. Stop PM2 application:

```bash
pm2 stop cli-proxy-api-plus
pm2 delete cli-proxy-api-plus
pm2 save
```

4. Follow installation steps above

5. Start systemd service:

```bash
sudo systemctl start cli-proxy-api-plus
```

## Additional Resources

- [systemd.service man page](https://www.freedesktop.org/software/systemd/man/systemd.service.html)
- [systemd.exec man page](https://www.freedesktop.org/software/systemd/man/systemd.exec.html)
- [CLIProxyAPI Documentation](https://github.com/router-for-me/CLIProxyAPI)
