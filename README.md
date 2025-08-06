# Assistant-CLI

A personal assistant command-line interface tool with various capabilities including text-to-speech conversion.

## Overview

Assistant-CLI is a Go-based personal assistant tool that currently provides text-to-speech functionality using Google Cloud Text-to-Speech API. It accepts text input via STDIN and converts it to speech, outputting the result as an MP3 file. The tool will be extended with additional features for Calendar, Gmail, and Drive integration.

## Features

- **Multiple Authentication Methods**: API Key, Service Account, and OAuth2
- **STDIN Input**: Pipe text directly into the tool
- **Voice Customization**: Adjust voice, language, speed, and pitch
- **Audio Playback**: Optionally play generated audio immediately
- **Cross-Platform**: Works on macOS, Linux, and Windows
- **Configuration Support**: Use config files for persistent settings

## Installation

### From Source

```bash
go install github.com/mikefarmer/assistant-cli@latest
```

### Pre-built Binaries

Download the appropriate binary for your platform from the [Releases](https://github.com/mikefarmer/assistant-cli/releases) page.

## Quick Start

1. Set up authentication (choose one method):

   ```bash
   # API Key (simplest)
   export ASSISTANT_CLI_API_KEY="your-api-key-here"
   
   # Or use the login command
   assistant-cli login --method api_key --api-key "your-api-key-here"
   ```

2. Convert text to speech:

   ```bash
   echo "Hello, World!" | assistant-cli synthesize -o hello.mp3
   ```

3. Play immediately after generation:

   ```bash
   echo "Hello, World!" | assistant-cli synthesize --play
   ```

## Authentication

### API Key (Recommended for Quick Start)

1. Create an API key in Google Cloud Console
2. Restrict it to Text-to-Speech API
3. Use it via environment variable or command flag

### Service Account (For Automation)

```bash
assistant-cli login --method service_account --credentials path/to/key.json
```

### OAuth2 (Interactive)

```bash
assistant-cli login --method oauth2
```

## Usage Examples

### Basic Usage

```bash
echo "Hello, World!" | assistant-cli synthesize -o output.mp3
```

### Custom Voice Parameters

```bash
cat text.txt | assistant-cli synthesize \
  --voice en-US-Wavenet-C \
  --speed 1.2 \
  --pitch -2.0 \
  --output speech.mp3
```

### Using Configuration File

```bash
echo "Welcome" | assistant-cli synthesize --config ~/.assistant-cli/config.yaml
```

## Configuration

Create a configuration file at `~/.assistant-cli/config.yaml`:

```yaml
auth:
  method: "api_key"
  api_key: ""  # Can be set here or via ASSISTANT_CLI_API_KEY env var

audio:
  voice: "en-US-Wavenet-D"
  language: "en-US"
  speaking_rate: 1.0
  pitch: 0.0

output:
  default_path: "./output"
  format: "MP3"
  overwrite: true

playback:
  auto_play: false
```

## Development

### Prerequisites

- Go 1.21 or later
- Google Cloud account with Text-to-Speech API enabled

### Building

```bash
# Build for current platform
go build -o assistant-cli main.go

# Cross-platform builds
make build-all
```

### Testing

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please read our [Contributing Guidelines](CONTRIBUTING.md) first.

## Support

- Documentation: [Wiki](https://github.com/mikefarmer/assistant-cli/wiki)
- Issues: [GitHub Issues](https://github.com/mikefarmer/assistant-cli/issues)

## Roadmap

- **Phase 1** (Current): Core TTS functionality
- **Phase 2**: Google services integration (Calendar, Gmail, Drive)
- **Phase 3**: MCP server capability for AI assistants