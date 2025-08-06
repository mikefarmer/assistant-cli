# Development Guide

## Prerequisites

- Go 1.21 or later
- Google Cloud account with Text-to-Speech API enabled
- golangci-lint (for linting)

## Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/mikefarmer/assistant-cli.git
   cd assistant-cli
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Build the project:
   ```bash
   make build
   ```

## Development Workflow

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run a specific test
go test -run TestRootCommand ./cmd
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Run all verification steps
make verify
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Install to GOPATH/bin
make install
```

## Project Structure

```
.
├── cmd/                    # CLI commands
│   ├── root.go            # Root command setup
│   ├── login.go           # Authentication command
│   └── synthesize.go      # TTS synthesis command
├── internal/              # Private packages
│   ├── auth/             # Authentication logic
│   ├── config/           # Configuration management
│   ├── tts/              # TTS client and synthesis
│   ├── output/           # File output handling
│   └── player/           # Audio playback
├── pkg/                   # Public packages
│   └── utils/            # Utility functions
├── main.go               # Entry point
└── Makefile              # Build automation
```

## Testing Strategy

We follow a test-driven development approach:

1. Write tests first
2. Implement minimal code to pass tests
3. Refactor while keeping tests green

### Test Organization

- Unit tests: Placed alongside the code they test (`*_test.go`)
- Integration tests: In `test/integration/`
- Test utilities: In `internal/testutil/`

### Mocking

For Google Cloud API calls, use interfaces and mock implementations:

```go
type TTSClient interface {
    Synthesize(ctx context.Context, req *SynthesizeRequest) (*SynthesizeResponse, error)
}
```

## Adding New Features

1. Create a new branch
2. Write tests for the feature
3. Implement the feature
4. Ensure all tests pass
5. Update documentation
6. Submit a pull request

## Debugging

### Verbose Output

Set the `ASSISTANT_CLI_DEBUG` environment variable:

```bash
ASSISTANT_CLI_DEBUG=true ./assistant-cli synthesize
```

### Common Issues

1. **Authentication failures**: Check credentials and API enablement
2. **Build failures**: Ensure Go version is 1.21+
3. **Test failures**: Run `make deps` to ensure dependencies are up to date

## Release Process

1. Update version in Makefile
2. Run `make verify`
3. Create git tag
4. Run `make build-all`
5. Create GitHub release with binaries