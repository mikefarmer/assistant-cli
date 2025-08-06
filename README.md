# Assistant-CLI

A personal assistant command-line interface tool with comprehensive authentication and text-to-speech capabilities.

## Overview

Assistant-CLI is a Go-based personal assistant tool designed with a phased approach. **Phase 1** (currently in progress) focuses on core text-to-speech functionality using Google Cloud Text-to-Speech API with robust authentication. Future phases will add Calendar, Gmail, Drive integration, and MCP server capabilities.

## Current Status: Phase 1.3 Complete ✅

- ✅ **Phase 1.1**: Project foundation with Go module, Cobra CLI, and directory structure
- ✅ **Phase 1.2**: Complete authentication system with API Key, Service Account, and OAuth2 support
- ✅ **Phase 1.3**: Core TTS integration with Google Cloud Text-to-Speech API
- 🚧 **Phase 1.4**: Audio playback and file management (next)

## Features

### Authentication (✅ Complete)
- **API Key Authentication**: Simplest method using Google Cloud API keys
- **Service Account Authentication**: JSON file-based auth for automation
- **OAuth2 Authentication**: Interactive browser flow with token caching and refresh
- **Auto-detection**: Automatically selects best auth method based on available credentials
- **Interactive Setup**: Guided authentication process with `assistant-cli login`

### Text-to-Speech (✅ Complete - Phase 1.3)
- **STDIN Input**: Pipe text directly into the tool with UTF-8 support
- **Voice Customization**: Comprehensive voice settings (voice, language, speed, pitch, volume)
- **Multiple Audio Formats**: MP3, LINEAR16/WAV, OGG_OPUS, MULAW, ALAW, PCM support
- **SSML Support**: Advanced speech markup language with security validation
- **Voice Discovery**: List available voices by language
- **Robust Error Handling**: Retry logic and comprehensive validation

### Platform & Configuration (✅ Complete)
- **Cross-Platform**: Works on macOS, Linux, and Windows
- **Configuration Support**: YAML config files with environment variable support
- **CLI Framework**: Built with Cobra for excellent user experience

## Installation

### Prerequisites
- Google Cloud account with Text-to-Speech API enabled
- API credentials (API key, service account, or OAuth2 client credentials)

### From Source

```bash
# Clone the repository
git clone https://github.com/mikefarmer/assistant-cli.git
cd assistant-cli

# Build the binary
go build -o assistant-cli main.go
```

### Pre-built Binaries

Download the appropriate binary for your platform from the [Releases](https://github.com/mikefarmer/assistant-cli/releases) page (coming soon).

## Quick Start

### 1. Authentication Setup

Choose one of the three authentication methods:

#### Option A: API Key (Simplest)
```bash
# Set environment variable
export ASSISTANT_CLI_API_KEY="your-google-cloud-api-key"

# Or use interactive login
./assistant-cli login --method apikey
```

#### Option B: Service Account (For Automation)
```bash
# Set environment variable
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"

# Or use interactive login
./assistant-cli login --method serviceaccount --service-account /path/to/key.json
```

#### Option C: OAuth2 (Interactive)
```bash
# Interactive browser-based authentication
./assistant-cli login --method oauth2 --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET
```

### 2. Verify Authentication

```bash
./assistant-cli login --validate
```

### 3. Text-to-Speech Usage (✅ Available Now)

```bash
# Basic text-to-speech
echo "Hello, World!" | ./assistant-cli synthesize -o hello.mp3

# Advanced voice customization
cat story.txt | ./assistant-cli synthesize \
  --voice en-US-Wavenet-C \
  --speed 1.2 \
  --pitch -2.0 \
  --volume 3.0 \
  --format LINEAR16 \
  --output speech.wav

# SSML support for advanced speech control
echo "<speak>Hello <break time='1s'/> <emphasis>World!</emphasis></speak>" | \
  ./assistant-cli synthesize --format MP3 -o greeting.mp3

# List available voices
./assistant-cli synthesize --list-voices --language en-US
```

## Authentication Methods

The assistant-cli supports three robust authentication methods with auto-detection and validation:

### 1. API Key Authentication (Simplest)

**Best for**: Quick start, personal use, simple scripts

**Setup**:
1. Create an API key in [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Enable Text-to-Speech API for your project  
3. Restrict the key to Text-to-Speech API (recommended)
4. Use one of these methods:

```bash
# Environment variable (recommended)
export ASSISTANT_CLI_API_KEY="your-api-key-here"

# Interactive login
./assistant-cli login --method apikey

# Direct flag (less secure)
./assistant-cli login --method apikey --api-key "your-api-key"
```

### 2. Service Account Authentication (For Automation)

**Best for**: Server deployments, CI/CD, automation, production environments

**Setup**:
1. Create a service account in Google Cloud Console
2. Grant it "Cloud Text-to-Speech User" role
3. Download the JSON key file
4. Use one of these methods:

```bash
# Environment variable (recommended)  
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"

# Interactive login
./assistant-cli login --method serviceaccount

# Direct flag
./assistant-cli login --method serviceaccount --service-account /path/to/key.json
```

### 3. OAuth2 Authentication (Interactive)

**Best for**: Desktop applications, user consent workflows, development

**Setup**:
1. Create OAuth2 credentials in Google Cloud Console
2. Set redirect URI to `http://localhost:8080/callback`
3. Use one of these methods:

```bash
# Environment variables (recommended)
export ASSISTANT_CLI_OAUTH2_CLIENT_ID="your-client-id"
export ASSISTANT_CLI_OAUTH2_CLIENT_SECRET="your-client-secret"

# Interactive login (will open browser)
./assistant-cli login --method oauth2

# Direct flags  
./assistant-cli login --method oauth2 --client-id ID --client-secret SECRET
```

### Authentication Management

```bash
# Check current authentication status
./assistant-cli login --validate

# Force re-authentication
./assistant-cli login --force

# Auto-detect and use best available method
./assistant-cli login

# Get help with authentication
./assistant-cli login --help
```

## Available Commands

### Authentication Commands (✅ Available Now)

```bash
# Interactive authentication setup
./assistant-cli login

# Specific authentication method
./assistant-cli login --method apikey|serviceaccount|oauth2

# Validate current authentication
./assistant-cli login --validate

# View help
./assistant-cli --help
./assistant-cli login --help
```

### Text-to-Speech Commands (✅ Available Now)

```bash
# Basic text-to-speech
echo "Hello, World!" | ./assistant-cli synthesize -o output.mp3

# Custom voice parameters
cat text.txt | ./assistant-cli synthesize \
  --voice en-US-Wavenet-C \
  --speed 1.2 \
  --pitch -2.0 \
  --volume 3.0 \
  --format LINEAR16 \
  --output speech.wav

# SSML markup for advanced control
echo "<speak>Hello <break time='500ms'/> World!</speak>" | \
  ./assistant-cli synthesize -o advanced.mp3

# List available voices for a language
./assistant-cli synthesize --list-voices --language en-US

# Using configuration file
echo "Welcome" | ./assistant-cli synthesize --config ~/.assistant-cli.yaml

# Multiple audio format support
echo "Test" | ./assistant-cli synthesize --format OGG_OPUS -o test.ogg
```

## Configuration

The assistant-cli uses a hierarchical configuration system: **CLI flags** > **Environment variables** > **Config file** > **Defaults**

### Configuration File

Create a configuration file at `~/.assistant-cli.yaml`:

```yaml
# Authentication settings (Phase 1.2 ✅)
auth:
  method: "apikey"  # apikey, serviceaccount, or oauth2
  service_account_file: "/path/to/key.json"  # Only for serviceaccount method
  # Note: Sensitive credentials (API keys, OAuth secrets) should use environment variables

# Text-to-Speech settings (Phase 1.3 ✅)  
tts:
  voice: "en-US-Wavenet-D"
  language: "en-US" 
  speaking_rate: 1.0
  pitch: 0.0
  volume_gain: 0.0

# Output settings (Phase 1.3 ✅)
output:
  default_path: "./output"
  format: "MP3"
  overwrite: true

# Playback settings (Phase 1.4 🚧)
playback:
  auto_play: false
```

### Environment Variables

```bash
# Authentication
export ASSISTANT_CLI_API_KEY="your-api-key"
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
export ASSISTANT_CLI_OAUTH2_CLIENT_ID="your-client-id"
export ASSISTANT_CLI_OAUTH2_CLIENT_SECRET="your-client-secret"
export ASSISTANT_CLI_OAUTH2_TOKEN_FILE="/custom/token/path.json"

# TTS and output settings (Phase 1.3 ✅)
export ASSISTANT_CLI_VOICE="en-US-Wavenet-C"
export ASSISTANT_CLI_OUTPUT_PATH="./speech-files"
export ASSISTANT_CLI_SPEAKING_RATE="1.2"
export ASSISTANT_CLI_PITCH="0.0"
export ASSISTANT_CLI_VOLUME_GAIN="0.0"
```

## Development

### Prerequisites

- Go 1.23.0 or later (project uses latest Go features)
- Google Cloud account with Text-to-Speech API enabled
- Git for version control

### Building

```bash
# Clone and build
git clone https://github.com/mikefarmer/assistant-cli.git
cd assistant-cli

# Install dependencies
go mod download

# Build for current platform
go build -o assistant-cli main.go

# Or use the Makefile
make build

# Cross-platform builds
make build-all
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Test specific package
go test ./internal/auth
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Update dependencies
go mod tidy
```

## Architecture

The project follows a clean, modular architecture designed for extensibility:

```
assistant-cli/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command and config
│   └── login.go           # Authentication commands ✅
├── internal/              # Private application code
│   ├── auth/              # Authentication system ✅
│   │   ├── manager.go     # Auth coordinator
│   │   ├── apikey.go      # API key provider
│   │   ├── service.go     # Service account provider
│   │   └── oauth2.go      # OAuth2 provider
│   ├── tts/               # TTS integration ✅
│   │   ├── client.go      # Google Cloud TTS client wrapper
│   │   └── synthesizer.go # Speech synthesis engine
│   ├── config/            # Configuration management 🚧
│   ├── output/            # File output handling 🚧
│   └── player/            # Audio playback 🚧
└── pkg/                   # Public/shared utilities
    └── utils/             # Common utilities 🚧
```

## Troubleshooting

### Authentication Issues

```bash
# Check authentication status
./assistant-cli login --validate

# Common issues:
# 1. API key invalid or expired
# 2. Service account file permissions
# 3. OAuth2 client credentials incorrect
# 4. Text-to-Speech API not enabled

# Force re-authentication
./assistant-cli login --force
```

### Build Issues

```bash
# Update to Go 1.23+
go version

# Clean module cache
go clean -modcache
go mod download

# Rebuild
go build -o assistant-cli main.go
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! This project follows a phased development approach:

1. **Phase 1.2** ✅ - Authentication system (complete)
2. **Phase 1.3** 🚧 - TTS integration (next priority)  
3. **Phase 1.4** 📋 - Input/output processing
4. **Phase 2** 📋 - Google services integration
5. **Phase 3** 📋 - MCP server capability

See [phase-1-tasks.md](phase-1-tasks.md) for detailed implementation status.

## Support

- **Issues**: [GitHub Issues](https://github.com/mikefarmer/assistant-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/mikefarmer/assistant-cli/discussions)
- **Documentation**: Check the `/docs` directory

## Project Status & Roadmap

### ✅ **Phase 1.1**: Project Foundation (Complete)
- Go module setup with latest Go 1.23
- Cobra CLI framework integration  
- Project structure and development tooling
- Cross-platform build configuration

### ✅ **Phase 1.2**: Authentication Foundation (Complete)  
- Multi-method authentication (API Key, Service Account, OAuth2)
- Interactive login command with validation
- Auto-detection and credential management
- Comprehensive error handling and user guidance

### ✅ **Phase 1.3**: Core TTS Integration (Complete)
- Google Cloud Text-to-Speech API integration with retry logic
- Voice synthesis with comprehensive customization (voice, speed, pitch, volume)
- SSML support with security validation for advanced speech control
- Multiple audio format support (MP3, LINEAR16/WAV, OGG_OPUS, MULAW, ALAW, PCM)
- Voice discovery and listing functionality
- Robust input validation and error handling

### 🚧 **Phase 1.4**: Audio Playback and Enhanced I/O (Next)
- Cross-platform audio playback functionality
- Enhanced file output management
- Configuration system improvements
- Performance optimizations

### 📋 **Future Phases**
- **Phase 1.5-1.10**: Testing, distribution, and polish
- **Phase 2**: Google services integration (Calendar, Gmail, Drive)  
- **Phase 3**: MCP server capability for AI assistant integration

---

*Assistant-CLI is actively developed with a focus on security, usability, and extensibility. Both authentication and core TTS functionality are production-ready. Audio playback is the next priority.*