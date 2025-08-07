# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based CLI tool called `assistant-cli` that serves as a personal assistant with various capabilities. The project is designed with a phased approach:

- **Phase 1** (Current): Core TTS functionality with multiple auth methods
- **Phase 2** (Future): Calendar, Gmail, and Drive integration
- **Phase 3** (Future): MCP (Model Context Protocol) server capability for AI integration

## Development Commands

### Build Commands
```bash
# Build the CLI binary
go build -o assistant-cli main.go

# Build with version information
go build -ldflags "-X main.version=1.0.0" -o assistant-cli main.go

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o assistant-cli-linux-amd64 main.go
GOOS=darwin GOARCH=amd64 go build -o assistant-cli-darwin-amd64 main.go
GOOS=windows GOARCH=amd64 go build -o assistant-cli-windows-amd64.exe main.go
```

### Testing Commands
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run a specific test
go test -run TestAuthManager ./internal/auth

# Run tests with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Development Commands
```bash
# Install dependencies
go mod download

# Update dependencies
go mod tidy

# Run linting (requires golangci-lint)
golangci-lint run

# Format code
go fmt ./...

# Run the CLI in development
go run main.go synthesize --help
```

## Architecture Overview

### Authentication System
The tool supports three authentication methods, implemented in `internal/auth/`:
1. **API Key** (`apikey.go`) - Simplest, uses environment variable or flag
2. **Service Account** (`service.go`) - JSON key file for automation
3. **OAuth2** (`oauth2.go`) - Interactive browser flow with token caching

The `AuthManager` in `internal/auth/manager.go` coordinates between these methods based on user configuration.

### Core Components

**Command Structure** (`cmd/`):
- `root.go` - Base CLI setup with Cobra
- `login.go` - Handles authentication setup and credential storage
- `synthesize.go` - Main TTS command with all voice/audio parameters

**TTS Integration** (`internal/tts/`):
- `client.go` - Wraps Google Cloud TTS client with retry logic, connection pooling, and performance monitoring
- `synthesizer.go` - Handles synthesis requests, SSML support, and audio generation
- `cache.go` - Implements voice list caching with TTL expiration and statistics
- `performance.go` - Provides comprehensive performance monitoring and benchmarking

**Configuration** (`internal/config/`):
- Uses Viper for hierarchical config (flags > env > file > defaults)
- Supports YAML configuration files
- Handles secure credential storage

**Audio Playback** (`internal/player/`):
- Platform-specific audio playback (afplay on macOS, aplay on Linux, PowerShell on Windows)
- Automatic player detection with fallback handling

### Data Flow
1. User provides text via STDIN
2. Authentication is established (using stored credentials or flags)
3. Performance monitoring begins (if enabled)
4. Voice list is retrieved (cached if available)
5. Text is validated and sent to Google Cloud TTS API
6. Audio data is received and written to file
7. Performance metrics are recorded
8. Optionally, audio is played immediately

### Key Design Decisions

**Multi-Phase Architecture**: The codebase is structured to support future expansion. Phase 1 code should be modular enough to integrate with additional Google services in Phase 2.

**Authentication Priority**: API key is the default for easy distribution, but the system is designed to seamlessly support all three auth methods.

**Error Handling**: All errors should be wrapped with context and provide actionable user messages. API errors should include retry guidance.

**Platform Support**: Code must work on Linux, macOS, and Windows. Use build tags for platform-specific code.

## Important Implementation Notes

1. **Credential Security**: Never store credentials in the binary. Use external files, environment variables, or secure token storage.

2. **SSML Support**: The input processor should detect and validate SSML markup to prevent injection attacks.

3. **Rate Limiting**: Respect Google Cloud API limits with exponential backoff and user-friendly error messages.

4. **Configuration Precedence**: Command flags override environment variables, which override config file values, which override defaults.

5. **Future MCP Integration**: Keep the service layer abstracted enough to expose via MCP protocol in Phase 3.

## Development Guidelines

- **Use the latest version of Go**: Always use the most recent stable Go version to leverage the latest language features and performance improvements.

## Testing Strategy

- Unit tests for all authentication methods with mocked Google Cloud clients
- Integration tests using test service accounts (not included in repo)
- CLI command tests using Cobra's test utilities
- Platform-specific code must be tested on all three major platforms
- **Use a test-driven approach to all work**

## Distribution

The tool will be distributed as:
- Pre-compiled binaries via GitHub Releases
- Homebrew formula for macOS
- Direct download scripts for Linux
- Users provide their own Google Cloud credentials