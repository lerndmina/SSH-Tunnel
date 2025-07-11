# SSH Tunnel Manager

A comprehensive, cross-platform SSH tunnel management tool written in Go. This application provides a modern replacement for traditional SSH tunnel scripts with enhanced features including multi-tunnel management, real-time monitoring, and cross-platform service integration.

## 🚀 Features

- **Cross-Platform Support**: Works on Linux, macOS, and Windows (ARM64 & x86_64)
- **Multi-Tunnel Management**: Manage multiple SSH tunnels simultaneously
- **Service Integration**: Native system service support (systemd, launchd, Windows Services)
- **Real-time Monitoring**: Interactive dashboard with live tunnel status
- **Configuration Templates**: Pre-built templates for common use cases
- **Backup & Restore**: Complete configuration backup and restore functionality
- **SSH Key Management**: Automated SSH key generation and distribution
- **Performance Optimization**: Built-in performance tuning and diagnostics
- **Analytics & Logging**: Comprehensive logging and usage analytics
- **Interactive Setup**: User-friendly setup wizard
- **CLI & TUI Interface**: Both command-line and terminal user interface

## 📦 Installation

### One-line Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/yourusername/ssh-tunnel-manager/main/install.sh | bash
```

### Manual Installation

1. Download the latest release for your platform from [Releases](https://github.com/yourusername/ssh-tunnel-manager/releases)
2. Extract the binary to your PATH:

```bash
# Linux/macOS
sudo mv ssh-tunnel /usr/local/bin/

# Windows (PowerShell as Administrator)
Move-Item ssh-tunnel.exe C:\Windows\System32\
```

### Build from Source

```bash
git clone https://github.com/yourusername/ssh-tunnel-manager.git
cd ssh-tunnel-manager
make build
```

### Go Install

```bash
go install github.com/yourusername/ssh-tunnel-manager/cmd/cli@latest
```

## 🔧 Quick Start

### 1. Setup Your First Tunnel

```bash
ssh-tunnel setup
```

The interactive setup wizard will guide you through:
- Tunnel configuration
- SSH key setup
- Cloud server connection
- Service installation

### 2. Manage Tunnels

```bash
# List all tunnels
ssh-tunnel list

# Start a tunnel
ssh-tunnel start my-tunnel

# Check status
ssh-tunnel status

# Stop a tunnel
ssh-tunnel stop my-tunnel

# View logs
ssh-tunnel logs my-tunnel -f
```

### 3. Monitor Tunnels

```bash
# Real-time monitoring dashboard
ssh-tunnel monitor

# Run diagnostics
ssh-tunnel diagnostics my-tunnel
```

## 📖 Usage

### Basic Commands

```bash
# Setup new tunnel
ssh-tunnel setup

# List all tunnels
ssh-tunnel list

# Start/stop tunnels
ssh-tunnel start [tunnel-name]
ssh-tunnel stop [tunnel-name]
ssh-tunnel restart [tunnel-name]

# Check status
ssh-tunnel status [tunnel-name]

# View logs
ssh-tunnel logs [tunnel-name] --follow

# Configuration management
ssh-tunnel config list
ssh-tunnel config show [tunnel-name]
ssh-tunnel config edit [tunnel-name]

# Templates
ssh-tunnel template list
ssh-tunnel template apply home-server my-home

# Backup operations
ssh-tunnel backup create
ssh-tunnel backup restore [backup-file]

# Monitoring and diagnostics
ssh-tunnel monitor
ssh-tunnel diagnostics [tunnel-name]
```

### Configuration File

Configuration files are stored in:
- Linux/macOS: `~/.ssh-tunnel-manager/`
- Windows: `%USERPROFILE%\.ssh-tunnel-manager\`

Example configuration:

```yaml
tunnel_name: "my-tunnel"
cloud_server:
  ip: "203.0.113.1"
  port: 22
  user: "ubuntu"
  home_dir: "/home/ubuntu"
local_server:
  user: "localuser"
  reverse_port: 2222
  socks_port: 1080
ssh:
  private_key_path: "/home/user/.ssh/cloud_server_key"
  natted_key_path: "/home/user/.ssh/natted_server_key"
  compression: true
service:
  name: "ssh-tunnel-my-tunnel"
  auto_reconnect: true
  restart_sec: 5
performance:
  keep_alive_interval: 30
  keep_alive_count_max: 3
  connect_timeout: 10
```

## 🏗️ Architecture

```
ssh-tunnel-manager/
├── cmd/cli/           # CLI application entry point
├── internal/
│   ├── config/        # Configuration management
│   ├── tunnel/        # Core tunnel functionality
│   ├── service/       # System service management
│   ├── ssh/           # SSH operations
│   └── ...
├── pkg/
│   ├── logger/        # Structured logging
│   └── ...
└── scripts/           # Installation and build scripts
```

### Core Components

- **Configuration Manager**: Handles tunnel configurations with validation
- **Tunnel Manager**: Manages SSH tunnel lifecycle
- **Service Manager**: Cross-platform system service integration
- **SSH Manager**: SSH key generation, validation, and connection testing
- **Monitor**: Real-time tunnel monitoring and health checks

## 🔒 Security Features

- **Key Management**: Secure SSH key generation and storage
- **Permission Validation**: Automatic file permission checks
- **Fingerprint Verification**: SSH host fingerprint validation
- **Encrypted Storage**: Configuration encryption at rest
- **Audit Logging**: Comprehensive security event logging

## 🛠️ Development

### Prerequisites

- Go 1.21 or later
- Make (for build automation)

### Setup Development Environment

```bash
git clone https://github.com/yourusername/ssh-tunnel-manager.git
cd ssh-tunnel-manager
make dev-setup
```

### Build and Test

```bash
# Run all checks
make ci

# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Format and lint
make fmt vet lint
```

### Cross-Platform Building

The project supports building for multiple platforms:

```bash
# Build for all supported platforms
make build-all

# Create release packages
make release
```

Supported platforms:
- Linux: amd64, arm64, arm
- macOS: amd64, arm64
- Windows: amd64

## 📊 Monitoring & Analytics

### Real-time Dashboard

Access the monitoring dashboard with:

```bash
ssh-tunnel monitor
```

Features:
- Live tunnel status
- Connection metrics
- Performance graphs
- Error tracking
- Uptime statistics

### Diagnostics

Run comprehensive diagnostics:

```bash
ssh-tunnel diagnostics my-tunnel --performance
```

Includes:
- Network connectivity tests
- SSH authentication validation
- Performance measurements
- Service health checks

## 🔄 Migration from Bash Script

To migrate from the original bash script:

1. **Export existing configuration**:
   ```bash
   # If you have the old script config
   ssh-tunnel template apply migration my-existing-tunnel
   ```

2. **Import SSH keys**:
   ```bash
   # The setup wizard will help import existing keys
   ssh-tunnel setup --import-keys
   ```

3. **Verify and test**:
   ```bash
   ssh-tunnel diagnostics my-existing-tunnel
   ```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go conventions and best practices
- Add tests for new functionality
- Update documentation for user-facing changes
- Ensure cross-platform compatibility

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Original bash script inspiration
- Go SSH library contributors
- Cross-platform service management libraries
- Community feedback and contributions

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/ssh-tunnel-manager/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/ssh-tunnel-manager/discussions)
- **Documentation**: [Wiki](https://github.com/yourusername/ssh-tunnel-manager/wiki)

## 🗺️ Roadmap

- [ ] Web-based dashboard
- [ ] Docker container support
- [ ] Kubernetes operator
- [ ] Mobile app for monitoring
- [ ] Advanced load balancing
- [ ] VPN integration
- [ ] Cloud provider integration (AWS, GCP, Azure)
- [ ] Terraform provider

---

**Made with ❤️ for the SSH tunneling community**
