# Technical Design Document: Go CLI Assistant Tool

## 1. Overview

### 1.1 Purpose
This document outlines the design for a Go-based command-line interface (CLI) tool that serves as a personal assistant with various capabilities. Currently, it provides text-to-speech functionality using Google Cloud Text-to-Speech API, with plans to expand into additional Google services integration.

### 1.2 Scope
- CLI tool written in Go using Cobra framework
- Integration with Google Cloud Text-to-Speech API
- STDIN text input processing
- MP3 audio file output
- Multiple authentication methods for easy distribution
- Configuration management for API credentials and voice settings
- Foundation for future expansion into a comprehensive Google services CLI

## 2. Project Phases

### 2.1 Phase 1: Core TTS Tool (Current Focus)
**Goal**: Build a robust, standalone text-to-speech CLI tool
- Complete text-to-speech functionality
- Multiple authentication methods (API key, service account, OAuth2)
- Audio file generation and playback
- Configuration management
- macOS distribution

### 2.2 Phase 2: Google Services Integration (Future)
**Goal**: Expand assistant-cli with multiple Google services
- **New Services**:
  - Google Calendar integration (read events, daily summaries)
  - Gmail integration (read emails, inbox summaries)
  - Google Drive integration (folder-scoped document access)
- **Architecture**: Service pipeline for chaining commands
- **Authentication**: Unified OAuth2 for all services

### 2.3 Phase 3: MCP Server Capability (Future)
**Goal**: Enable AI assistants to use the tool programmatically
- Implement Model Context Protocol (MCP) server
- Expose all services as MCP tools and resources
- JSON-RPC 2.0 communication protocol
- Integration with Claude, OpenAI, and other LLMs
- Enable AI-driven automation workflows

## 3. System Architecture

### 2.1 High-Level Architecture
```
┌─────────────┐    ┌──────────────┐    ┌─────────────────┐    ┌─────────────┐
│    STDIN    │───▶│  CLI Tool    │───▶│ Google Cloud    │───▶│ MP3 Output  │
│    Text     │    │   (Go)       │    │ Text-to-Speech  │    │    File     │
└─────────────┘    └──────────────┘    └─────────────────┘    └─────────────┘
```

### 2.2 Component Architecture
```
assistant-cli/
├── cmd/
│   ├── root.go          # Root command definition
│   └── synthesize.go    # Main synthesize command
├── internal/
│   ├── config/
│   │   ├── config.go    # Configuration management
│   │   └── auth.go      # Multi-method authentication
│   ├── auth/
│   │   ├── apikey.go    # API key authentication
│   │   ├── service.go   # Service account authentication
│   │   └── oauth2.go    # OAuth2 authentication
│   ├── tts/
│   │   ├── client.go    # Google Cloud TTS client wrapper
│   │   └── synthesizer.go # Speech synthesis logic
│   ├── output/
│   │   └── file.go      # File output handling
│   └── player/
│       └── audio.go     # Audio playback functionality
├── pkg/
│   └── utils/
│       ├── input.go     # STDIN input processing
│       └── validation.go # Input validation
├── main.go              # Application entry point
├── go.mod               # Go module definition
└── go.sum               # Dependency checksums
```

## 4. Detailed Component Design

### 3.1 CLI Interface Design

**Command Structure:**
```bash
assistant-cli [flags]
assistant-cli synthesize [flags]
```

**Primary Commands:**
- `login` - Authenticate with Google Cloud services
- `synthesize` - Convert text to speech

**Login Command Flags:**
- `--method` - Authentication method: "api_key", "service_account", "oauth2" (default: "oauth2")
- `--api-key` - API key for authentication (when method=api_key)
- `--credentials` - Service account file path (when method=service_account)
- `--no-browser` - Disable automatic browser opening for OAuth2
- `--port` - Local port for OAuth2 callback (default: 8080)

**Synthesize Command Flags:**
- `--output, -o` - Output file path (default: "output.mp3")
- `--voice` - Voice name (default: "en-US-Wavenet-D")
- `--language` - Language code (default: "en-US")
- `--speed` - Speaking rate (default: 1.0)
- `--pitch` - Voice pitch (default: 0.0)
- `--format` - Audio format (default: "MP3")
- `--play, -p` - Play audio immediately after generation (default: false)
- `--auth-method` - Authentication method: "api_key", "service_account", "oauth2" (default: "api_key")
- `--api-key` - Google Cloud API key for Text-to-Speech
- `--credentials` - Service account credentials file path
- `--config` - Config file path

### 3.2 Input Processing

**STDIN Handler:**
```go
type InputProcessor struct {
    MaxLength int
    Encoding  string
}

func (ip *InputProcessor) ReadFromSTDIN() (string, error)
func (ip *InputProcessor) ValidateInput(text string) error
```

**Features:**
- Read text from STDIN with configurable limits
- UTF-8 encoding support
- Input validation (length, character encoding)
- SSML support detection and handling

### 3.3 Authentication System

**Authentication Manager:**
```go
type AuthManager struct {
    Method AuthMethod
    Config *AuthConfig
}

type AuthMethod string

const (
    AuthMethodAPIKey         AuthMethod = "api_key"
    AuthMethodServiceAccount AuthMethod = "service_account"
    AuthMethodOAuth2         AuthMethod = "oauth2"
)

type AuthConfig struct {
    APIKey          string `yaml:"api_key"`
    CredentialsFile string `yaml:"credentials_file"`
    TokenCachePath  string `yaml:"token_cache_path"`
}

func (am *AuthManager) CreateClient(ctx context.Context) (*texttospeech.Client, error)
```

### 3.4 Google Cloud Integration

**TTS Client Wrapper:**
```go
type TTSClient struct {
    client      *texttospeech.Client
    authManager *AuthManager
    config      *Config
}

func NewTTSClient(ctx context.Context, config *Config) (*TTSClient, error)
func (t *TTSClient) Synthesize(ctx context.Context, request *SynthesizeRequest) (*SynthesizeResponse, error)
```

**Request/Response Models:**
```go
type SynthesizeRequest struct {
    Text        string
    Voice       VoiceConfig
    AudioConfig AudioConfig
}

type VoiceConfig struct {
    LanguageCode string
    Name         string
    Gender       texttospeechpb.SsmlVoiceGender
}

type AudioConfig struct {
    AudioEncoding   texttospeechpb.AudioEncoding
    SpeakingRate    float64
    Pitch           float64
    VolumeGainDb    float64
}
```

### 3.5 Configuration Management

**Configuration Structure:**
```go
type Config struct {
    Auth        AuthConfig        `yaml:"auth"`
    GoogleCloud GoogleCloudConfig `yaml:"google_cloud"`
    Audio       AudioConfig       `yaml:"audio"`
    Output      OutputConfig      `yaml:"output"`
    Playback    PlaybackConfig    `yaml:"playback"`
}

type GoogleCloudConfig struct {
    ProjectID string `yaml:"project_id"`
    Region    string `yaml:"region"`
}

type PlaybackConfig struct {
    AutoPlay bool   `yaml:"auto_play"`
    Command  string `yaml:"command"`  // Override default audio player
}
```

**Configuration Sources (Priority Order):**
1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values

### 3.6 Output Management

**File Output Handler:**
```go
type FileWriter struct {
    OutputPath string
    Overwrite  bool
}

func (fw *FileWriter) WriteAudio(data []byte, format string) error
func (fw *FileWriter) ValidateOutputPath() error
```

### 3.7 Audio Playback

**Audio Player:**
```go
type AudioPlayer struct {
    Command string // Platform-specific playback command
}

func NewAudioPlayer() *AudioPlayer
func (ap *AudioPlayer) Play(filepath string) error
func (ap *AudioPlayer) DetectCommand() string
```

**Platform-specific playback commands:**
- macOS: `afplay`

## 5. Data Flow

### 4.1 Processing Pipeline
1. **Input Stage**: Read text from STDIN
2. **Validation Stage**: Validate input text and configuration
3. **Authentication Stage**: Authenticate using selected method (API key, service account, or OAuth2)
4. **Synthesis Stage**: Send request to Text-to-Speech API
5. **Output Stage**: Write MP3 data to file
6. **Playback Stage**: Optionally play audio file (if --play flag is set)
7. **Cleanup Stage**: Close connections and resources

### 4.2 Error Handling
- Input validation errors
- Authentication failures
- API rate limiting and quota errors
- Network connectivity issues
- File system errors
- Audio playback failures (missing player, unsupported format)

## 6. Security Considerations

### 5.1 Authentication
- Support for multiple authentication methods:
  - **API Key** (Primary - Easiest for distribution)
    - Restricted to Text-to-Speech API only
    - Can be limited by IP addresses
    - Easy to rotate and revoke
  - **Service Account** (Secondary - For advanced users)
    - JSON key file with minimal permissions
    - Suitable for automation and CI/CD
  - **OAuth2** (Optional - For interactive use)
    - Browser-based authentication flow
    - Tokens cached locally with encryption
    - Automatic token refresh
- Environment variable support:
  - `ASSISTANT_CLI_API_KEY` for API key
  - `GOOGLE_APPLICATION_CREDENTIALS` for service account
- Secure credential storage and handling

### 5.2 Input Validation
- Text length limits to prevent abuse
- Character encoding validation
- SSML injection prevention

### 5.3 Output Security
- Safe file path handling
- Directory traversal prevention
- File permission management

### 5.4 Authentication Security Best Practices
- Never embed credentials in binary
- Use least-privilege principle for all authentication methods
- Implement rate limiting for API key usage
- Provide clear documentation on secure key storage
- Support credential rotation without code changes

## 7. Performance Considerations

### 6.1 Optimization Strategies
- Connection pooling for Google Cloud client
- Streaming for large text inputs
- Configurable timeouts
- Memory-efficient audio processing

### 6.2 Rate Limiting
- Built-in respect for Google Cloud API limits
- Configurable retry mechanisms
- Exponential backoff implementation

## 8. Dependencies

### 7.1 Core Dependencies
```go
require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2
    cloud.google.com/go/texttospeech v1.12.1
    google.golang.org/api v0.152.0
    google.golang.org/grpc v1.60.1
)
```

### 7.2 Development Dependencies
- Testing: `github.com/stretchr/testify`
- Mocking: `github.com/golang/mock`
- Linting: `golangci-lint`

## 9. Testing Strategy

### 8.1 Unit Tests
- Input processing validation
- Configuration management
- Audio output handling
- Error handling scenarios

### 8.2 Integration Tests
- Google Cloud API integration
- End-to-end CLI functionality
- File system operations

### 8.3 Test Coverage
- Target: 85% code coverage
- Critical path testing
- Error condition testing

## 10. Deployment and Distribution

### 9.1 Build Process
- macOS compilation
- Static binary generation
- Version embedding
- No embedded credentials

### 9.2 Distribution Methods
- **GitHub Releases**: Primary distribution method with semantic versioning
- **Semantic Versioning**: Following semver (MAJOR.MINOR.PATCH) for version management
- **Automated Releases**: GitHub Actions workflow for automated binary builds and releases
- **Release Assets**: Pre-compiled macOS binaries attached to each release
- **Changelog Generation**: Automated changelog generation from commit history

### 9.3 Authentication Setup Guide

**For End Users:**
1. **API Key Setup** (Recommended):
   ```bash
   # Set via environment variable
   export ASSISTANT_CLI_API_KEY="your-api-key-here"
   
   # Or use command flag
   echo "Hello" | assistant-cli synthesize --api-key "your-api-key-here"
   ```

2. **Service Account Setup**:
   ```bash
   # Download service account key from Google Cloud Console
   echo "Hello" | assistant-cli synthesize \
     --auth-method service_account \
     --credentials ~/path/to/service-account.json
   ```

3. **OAuth2 Setup** (First time only):
   ```bash
   # Will open browser for authentication
   echo "Hello" | assistant-cli synthesize --auth-method oauth2
   ```

## 11. Future Enhancements

### 11.1 Phase 1 Enhancements
- Multiple voice support in single request
- Batch processing capabilities
- Audio format options (WAV, OGG)
- SSML advanced features support
- Performance optimizations

### 11.2 Phase 2: Google Services Integration

#### 11.2.1 Architecture Evolution
```
assistant-cli/
├── cmd/
│   ├── root.go
│   ├── tts.go          # TTS commands
│   ├── calendar.go     # Calendar commands
│   ├── gmail.go        # Gmail commands
│   └── drive.go        # Drive commands
├── internal/
│   ├── services/
│   │   ├── tts/        # Existing TTS service
│   │   ├── calendar/   # Calendar integration
│   │   ├── gmail/      # Gmail integration
│   │   └── drive/      # Drive integration
│   └── pipeline/       # Service chaining logic
```

#### 11.2.2 Example Usage
```bash
# Calendar to speech
assistant-cli calendar today | assistant-cli tts synthesize --play

# Email summary
assistant-cli gmail unread --summarize | assistant-cli tts synthesize -o inbox.mp3

# Daily briefing workflow
assistant-cli assistant daily-briefing --output daily.mp3
```

#### 11.2.3 OAuth2 Scopes
```go
const (
    ScopeCalendar = "https://www.googleapis.com/auth/calendar.readonly"
    ScopeGmail    = "https://www.googleapis.com/auth/gmail.readonly"
    ScopeDrive    = "https://www.googleapis.com/auth/drive.file"
    ScopeTTS      = "https://www.googleapis.com/auth/cloud-platform"
)
```

### 11.3 Phase 3: MCP Server Implementation

#### 11.3.1 MCP Architecture
```go
// MCP server implementation
type MCPServer struct {
    services map[string]Service
    transport Transport
}

// Exposed MCP tools
func (m *MCPServer) GetTools() []mcp.Tool {
    return []mcp.Tool{
        {
            Name: "synthesize_speech",
            Description: "Convert text to speech",
            InputSchema: SynthesizeSchema,
        },
        {
            Name: "get_calendar_events",
            Description: "Retrieve calendar events",
            InputSchema: CalendarSchema,
        },
        {
            Name: "read_emails",
            Description: "Read and summarize emails",
            InputSchema: GmailSchema,
        },
    }
}

// MCP resources for data access
func (m *MCPServer) GetResources() []mcp.Resource {
    return []mcp.Resource{
        {
            Name: "calendar/events",
            Description: "Access calendar events",
            MimeType: "application/json",
        },
        {
            Name: "drive/documents",
            Description: "Access Drive documents",
            MimeType: "text/plain",
        },
    }
}
```

#### 11.3.2 AI Assistant Integration
```bash
# Start MCP server
assistant-cli mcp serve --port 3000

# AI assistants can now:
# - Generate daily briefings
# - Create audio summaries of emails
# - Convert calendar events to speech
# - Access and read Drive documents
```

#### 11.3.3 Benefits
- Direct integration with Claude, ChatGPT, and other LLMs
- Enables complex automation workflows
- Provides secure, controlled access to Google services
- Supports both tools (actions) and resources (data)

## 12. Example Usage

### 11.1 Basic Usage
```bash
echo "Hello, World!" | assistant-cli synthesize -o hello.mp3
```

### 11.2 Advanced Usage
```bash
# With custom voice and parameters
cat long-text.txt | assistant-cli synthesize \
  --voice en-US-Wavenet-C \
  --speed 1.2 \
  --pitch -2.0 \
  --output speech.mp3

# Using API key authentication
echo "Hello, World!" | assistant-cli synthesize \
  --api-key "AIza..." \
  --play

# Using service account authentication
echo "Important notification" | assistant-cli synthesize \
  --auth-method service_account \
  --credentials ~/.gcp/service-account.json \
  --output notification.mp3 \
  --play

# Using OAuth2 authentication
echo "Welcome" | assistant-cli synthesize \
  --auth-method oauth2 \
  -o welcome.mp3

# Using configuration file
echo "Welcome to our service" | assistant-cli synthesize \
  --config ~/.assistant-cli/config.yaml \
  -o welcome.mp3
```

### 11.3 Configuration File Example
```yaml
auth:
  method: "api_key"  # Options: api_key, service_account, oauth2
  api_key: ""  # Can be set here or via ASSISTANT_CLI_API_KEY env var
  credentials_file: "~/.gcp/service-account.json"
  token_cache_path: "~/.assistant-cli/token-cache.json"

google_cloud:
  project_id: "my-project-id"  # Optional for API key auth
  region: "us-central1"

audio:
  voice: "en-US-Wavenet-D"
  language: "en-US"
  speaking_rate: 1.0
  pitch: 0.0
  volume_gain_db: 0.0

output:
  default_path: "./output"
  format: "MP3"
  overwrite: true

playback:
  auto_play: false
  command: ""  # Leave empty for auto-detection
```

## 13. Implementation Timeline

### Phase 1: Core Implementation (Week 1-2)
- Project setup and structure
- Basic CLI framework
- Multi-method authentication system
- Google Cloud integration
- Simple text-to-speech conversion

### Phase 2: Enhanced Features (Week 3-4)
- Configuration management
- Advanced voice parameters
- Error handling and logging
- Unit tests

### Phase 3: Polish and Distribution (Week 5)
- Integration tests
- Documentation
- Build automation
- Release preparation