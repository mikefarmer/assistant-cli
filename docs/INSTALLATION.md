# Installation Guide

This guide provides detailed installation instructions for Assistant-CLI across different platforms and use cases.

## Table of Contents

- [System Requirements](#system-requirements)
- [Quick Installation](#quick-installation)
- [Platform-Specific Installation](#platform-specific-installation)
- [Installation Methods](#installation-methods)
- [Google Cloud Setup](#google-cloud-setup)
- [Verification](#verification)
- [Uninstallation](#uninstallation)

## System Requirements

### Minimum Requirements
- **Operating System**: macOS 10.15+, Linux (Ubuntu 18.04+, RHEL 7+), Windows 10+
- **Architecture**: AMD64 (x86_64) or ARM64
- **Memory**: 50MB RAM minimum, 100MB recommended
- **Disk Space**: 20MB for binary, 100MB for cache and temporary files
- **Network**: Internet connection for Google Cloud TTS API

### Recommended Requirements
- **Memory**: 200MB+ RAM for optimal performance
- **Disk Space**: 500MB+ for audio file storage
- **Network**: Stable broadband connection for best TTS quality

### Dependencies
- **No additional dependencies** - Assistant-CLI is a single static binary
- **Audio Playback** (optional): 
  - macOS: afplay (built-in)
  - Linux: aplay (usually pre-installed)
  - Windows: PowerShell (built-in)

## Quick Installation

### One-Line Installation

#### macOS
```bash
# Detect architecture and install
curl -s https://raw.githubusercontent.com/mikefarmer/assistant-cli/main/scripts/install.sh | bash
```

#### Linux
```bash
# Detect architecture and install
curl -s https://raw.githubusercontent.com/mikefarmer/assistant-cli/main/scripts/install.sh | bash
```

#### Windows (PowerShell)
```powershell
# Download and install
Invoke-Expression (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/mikefarmer/assistant-cli/main/scripts/install.ps1" -UseBasicParsing).Content
```

### Manual Quick Install

#### macOS
```bash
# Apple Silicon (M1/M2/M3)
curl -L -o assistant-cli https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-darwin-arm64

# Intel
curl -L -o assistant-cli https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-darwin-amd64

chmod +x assistant-cli
sudo mv assistant-cli /usr/local/bin/
```

#### Linux
```bash
# AMD64
curl -L -o assistant-cli https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-linux-amd64

# ARM64
curl -L -o assistant-cli https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-linux-arm64

chmod +x assistant-cli
sudo mv assistant-cli /usr/local/bin/
```

#### Windows
```powershell
# Download to current directory
Invoke-WebRequest -Uri "https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-windows-amd64.exe" -OutFile "assistant-cli.exe"

# Optional: Add to PATH
$env:PATH += ";$PWD"
```

## Platform-Specific Installation

### macOS Installation

#### Option 1: Homebrew (Recommended)
```bash
# Add tap (when available)
brew tap mikefarmer/assistant-cli
brew install assistant-cli
```

#### Option 2: Direct Download
```bash
# Detect architecture automatically
ARCH=$(uname -m)
if [ "$ARCH" = "arm64" ]; then
    BINARY_URL="https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-darwin-arm64"
else
    BINARY_URL="https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-darwin-amd64"
fi

curl -L -o assistant-cli $BINARY_URL
chmod +x assistant-cli
sudo mv assistant-cli /usr/local/bin/

# Verify installation
assistant-cli --version
```

#### Option 3: Local User Installation
```bash
# Install to user directory (no sudo required)
mkdir -p ~/bin
curl -L -o ~/bin/assistant-cli https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-darwin-arm64
chmod +x ~/bin/assistant-cli

# Add to PATH in ~/.zshrc or ~/.bash_profile
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### Linux Installation

#### Option 1: System-wide Installation
```bash
# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        BINARY_URL="https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-linux-amd64"
        ;;
    aarch64|arm64)
        BINARY_URL="https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-linux-arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

sudo curl -L -o /usr/local/bin/assistant-cli $BINARY_URL
sudo chmod +x /usr/local/bin/assistant-cli

# Verify installation
assistant-cli --version
```

#### Option 2: Package Manager Installation (Future)
```bash
# Ubuntu/Debian (when available)
curl -s https://packagecloud.io/install/repositories/mikefarmer/assistant-cli/script.deb.sh | sudo bash
sudo apt install assistant-cli

# RHEL/CentOS (when available)
curl -s https://packagecloud.io/install/repositories/mikefarmer/assistant-cli/script.rpm.sh | sudo bash
sudo yum install assistant-cli
```

#### Option 3: User Directory Installation
```bash
# Install to user directory
mkdir -p ~/.local/bin
curl -L -o ~/.local/bin/assistant-cli https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-linux-amd64
chmod +x ~/.local/bin/assistant-cli

# Add to PATH in ~/.bashrc
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Windows Installation

#### Option 1: PowerShell Installation
```powershell
# Create installation directory
New-Item -ItemType Directory -Force -Path "$env:LOCALAPPDATA\assistant-cli"

# Download binary
Invoke-WebRequest -Uri "https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-windows-amd64.exe" -OutFile "$env:LOCALAPPDATA\assistant-cli\assistant-cli.exe"

# Add to user PATH
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
$newPath = "$userPath;$env:LOCALAPPDATA\assistant-cli"
[Environment]::SetEnvironmentVariable("PATH", $newPath, "User")

# Restart PowerShell or use full path
& "$env:LOCALAPPDATA\assistant-cli\assistant-cli.exe" --version
```

#### Option 2: Chocolatey (Future)
```powershell
# When package is available
choco install assistant-cli
```

#### Option 3: Manual Installation
```powershell
# Download to current directory
Invoke-WebRequest -Uri "https://github.com/mikefarmer/assistant-cli/releases/latest/download/assistant-cli-windows-amd64.exe" -OutFile "assistant-cli.exe"

# Run directly
.\assistant-cli.exe --version

# Or move to preferred location
Move-Item assistant-cli.exe C:\Tools\assistant-cli.exe
```

## Installation Methods

### Method 1: Pre-built Binaries (Recommended)

**Pros**: 
- Fastest installation
- No build dependencies required
- Officially tested binaries

**Cons**: 
- Limited to supported platforms
- Requires internet connection

**Use cases**: 
- Production deployments
- Quick testing
- End users

### Method 2: Build from Source

**Pros**: 
- Latest development features
- Custom build configurations
- Full control over dependencies

**Cons**: 
- Requires Go 1.23+
- Longer setup time
- Build dependencies needed

**Installation**:
```bash
# Install Go 1.23+ first
git clone https://github.com/mikefarmer/assistant-cli.git
cd assistant-cli
go build -o assistant-cli main.go
sudo mv assistant-cli /usr/local/bin/
```

### Method 3: Container Installation

**Pros**: 
- Isolated environment
- Consistent across platforms
- Easy cleanup

**Cons**: 
- Requires Docker
- Larger resource usage

**Installation**:
```bash
# Docker run (when available)
docker run --rm -i mikefarmer/assistant-cli:latest --help

# Docker alias for easy use
echo 'alias assistant-cli="docker run --rm -i -v $(pwd):/workspace mikefarmer/assistant-cli:latest"' >> ~/.bashrc
```

## Google Cloud Setup

Assistant-CLI requires Google Cloud Text-to-Speech API access. Choose one authentication method:

### Method 1: API Key (Simplest)

1. **Create API Key**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
   - Click "Create Credentials" → "API Key"
   - Copy the generated key

2. **Enable Text-to-Speech API**:
   - Go to [APIs & Services](https://console.cloud.google.com/apis/library)
   - Search for "Cloud Text-to-Speech API"
   - Click "Enable"

3. **Set Environment Variable**:
   ```bash
   export ASSISTANT_CLI_API_KEY="your-api-key-here"
   
   # Make persistent
   echo 'export ASSISTANT_CLI_API_KEY="your-api-key-here"' >> ~/.bashrc
   source ~/.bashrc
   ```

### Method 2: Service Account (Production)

1. **Create Service Account**:
   - Go to [IAM & Admin → Service Accounts](https://console.cloud.google.com/iam-admin/serviceaccounts)
   - Click "Create Service Account"
   - Add "Cloud Text-to-Speech User" role

2. **Download Key File**:
   - Click on service account
   - Go to "Keys" tab
   - Click "Add Key" → "Create new key" → "JSON"
   - Download the JSON file

3. **Set Environment Variable**:
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
   
   # Make persistent
   echo 'export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"' >> ~/.bashrc
   source ~/.bashrc
   ```

### Method 3: OAuth2 (Interactive)

1. **Create OAuth2 Credentials**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
   - Click "Create Credentials" → "OAuth client ID"
   - Choose "Desktop application"
   - Set redirect URI: `http://localhost:8080/callback`

2. **Set Environment Variables**:
   ```bash
   export ASSISTANT_CLI_OAUTH2_CLIENT_ID="your-client-id"
   export ASSISTANT_CLI_OAUTH2_CLIENT_SECRET="your-client-secret"
   ```

3. **Run Interactive Login**:
   ```bash
   assistant-cli login --method oauth2
   ```

## Verification

### Basic Verification
```bash
# Check installation
assistant-cli --version
assistant-cli --help

# Test authentication
assistant-cli login --validate

# Test basic functionality
echo "Hello, World!" | assistant-cli synthesize -o test.mp3

# Test audio playback (optional)
echo "Hello, World!" | assistant-cli synthesize -o test.mp3 --play
```

### Complete Verification
```bash
# Test all major features
assistant-cli --version
assistant-cli config generate test-config.yaml
assistant-cli config validate test-config.yaml
assistant-cli synthesize --list-voices --language en-US
echo "Testing Assistant CLI installation" | assistant-cli synthesize -o verification.mp3 --play
```

### Performance Verification
```bash
# Test performance
time echo "Performance test" | assistant-cli synthesize -o perf-test.mp3

# Check binary size
ls -lh $(which assistant-cli)

# Test memory usage
echo "Memory test" | assistant-cli synthesize -o memory-test.mp3 &
ps aux | grep assistant-cli
```

## Troubleshooting Installation

### Common Issues

1. **Permission Denied**:
   ```bash
   chmod +x assistant-cli
   ```

2. **Command Not Found**:
   ```bash
   # Check PATH
   echo $PATH
   which assistant-cli
   
   # Add to PATH
   export PATH="/usr/local/bin:$PATH"
   ```

3. **macOS Security Warning**:
   ```bash
   # Remove quarantine attribute
   sudo xattr -rd com.apple.quarantine /usr/local/bin/assistant-cli
   ```

4. **Windows Execution Policy**:
   ```powershell
   # Allow execution
   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
   ```

### Verification Commands
```bash
# System info
uname -a
which assistant-cli
assistant-cli --version

# Test basic functionality
assistant-cli --help
echo "test" | assistant-cli synthesize -o test.mp3
```

## Uninstallation

### Remove Binary
```bash
# System-wide installation
sudo rm /usr/local/bin/assistant-cli

# User installation
rm ~/.local/bin/assistant-cli
rm ~/bin/assistant-cli
```

### Remove Configuration
```bash
# Remove config files
rm ~/.assistant-cli.yaml
rm -rf ~/.config/assistant-cli/

# Remove cache
rm -rf ~/.cache/assistant-cli/
```

### Remove from PATH
```bash
# Edit ~/.bashrc, ~/.zshrc, or ~/.profile
# Remove lines containing assistant-cli PATH exports
```

### Complete Cleanup
```bash
# Remove everything
sudo rm -f /usr/local/bin/assistant-cli
rm -f ~/.local/bin/assistant-cli
rm -f ~/bin/assistant-cli
rm -f ~/.assistant-cli.yaml
rm -rf ~/.config/assistant-cli/
rm -rf ~/.cache/assistant-cli/
rm -f ~/.assistant-cli-token.json

# Remove PATH entries (manual edit required)
echo "Please manually remove assistant-cli PATH entries from your shell config files"
```

## Next Steps

After installation:

1. **Set up authentication**: Follow [Authentication Guide](AUTHENTICATION.md)
2. **Basic usage**: See [Quick Start Guide](../README.md#quick-start)
3. **Configuration**: Read [Configuration Guide](CONFIGURATION.md)
4. **Troubleshooting**: Check [Troubleshooting Guide](TROUBLESHOOTING.md)

---

## Support

- **Issues**: [GitHub Issues](https://github.com/mikefarmer/assistant-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mikefarmer/assistant-cli/discussions)
- **Documentation**: [Full Documentation](../README.md)

---

*For additional installation methods and platform-specific instructions, check the [main documentation](../README.md) or create an issue for your specific use case.*