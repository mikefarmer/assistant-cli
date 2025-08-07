# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Added GitHub Actions CI/CD pipeline for automated testing and releases
- Enhanced distribution preparation with cross-platform builds and checksums

## [1.0.0] - 2025-08-07

### Added - Phase 1 Complete: Production-Ready TTS CLI Tool

#### 🔐 Enterprise Authentication System
- **Multi-Method Authentication**: API Key (simple), Service Account (automation), OAuth2 (interactive)
- **Auto-Detection**: Intelligent credential discovery and validation
- **Interactive Setup**: Guided login process with `assistant-cli login`
- **Security**: Secure credential storage and handling

#### 🗣️ Professional Text-to-Speech Integration  
- **Google Cloud TTS**: Full API integration with retry logic and error handling
- **Voice Customization**: Comprehensive control (voice, language, speed, pitch, volume)
- **SSML Support**: Advanced speech markup with security validation
- **Multiple Formats**: MP3, LINEAR16/WAV, OGG_OPUS, MULAW, ALAW, PCM support
- **Voice Discovery**: List and explore available voices by language

#### 🎵 Native Audio Playback
- **Cross-Platform Support**: macOS, Linux, and Windows audio playback
- **Native Integration**: macOS afplay, Linux aplay, Windows PowerShell
- **Automatic Detection**: Smart platform detection with fallback handling
- **Instant Playback**: `--play` flag for immediate audio playback

#### ⚙️ Enterprise Configuration Management
- **Hierarchical Structure**: YAML configuration with nested sections
- **Source Precedence**: Flags > environment variables > config file > defaults
- **Validation System**: Type checking, range validation, helpful error messages
- **Management Commands**: Generate, validate, and inspect configuration with `assistant-cli config`

#### ⚡ Performance Optimization & Caching
- **Connection Pooling**: Optimized gRPC connections with keep-alive settings
- **Intelligent Caching**: Voice list caching with TTL expiration and management
- **Performance Monitoring**: Real-time metrics with latency percentiles (P50/P90/P99)
- **System Monitoring**: Memory usage, GC statistics, and resource tracking

#### 🛡️ Security & Input Processing
- **SSML Security**: Injection prevention, dangerous pattern detection, tag whitelisting
- **Input Validation**: UTF-8 validation, text statistics, intelligent text splitting
- **File Safety**: Path validation, backup creation, traversal protection
- **Enterprise-Grade**: Security-first approach with comprehensive validation

#### 🧪 Comprehensive Testing Foundation
- **Test Coverage**: 45%+ overall with 150+ unit and integration tests
- **Quality Gates**: Automated coverage validation and reporting
- **Cross-Platform**: Tests work consistently across macOS, Linux, and Windows
- **Performance Benchmarks**: CLI startup and execution performance measurement

#### 📊 Core Capabilities
- **STDIN Input**: Pipe text directly with full UTF-8 support
- **Smart Output**: Automatic filename generation and safe file operations
- **Error Handling**: Comprehensive error handling with actionable messages
- **CLI Framework**: Built with Cobra for excellent user experience
- **Professional Output**: Formatted tables, progress indicators, colored output

### Architecture Highlights
- **Modular Design**: Clean separation of concerns with extensible architecture
- **Interface-Based**: Testable code with proper dependency injection
- **Enterprise-Ready**: Production-grade error handling and validation
- **Platform-Agnostic**: Works seamlessly across all major operating systems
- **Future-Proof**: Designed for Phase 2 Google services integration

### Technical Achievements
- **Go 1.23**: Latest Go features for optimal performance
- **Zero Dependencies**: Minimal external dependencies for security
- **Memory Efficient**: Optimized for low memory usage and fast startup
- **Connection Optimized**: Advanced gRPC connection management
- **Cache Intelligent**: Smart caching strategies for improved performance

---

### Phase 1 Sub-Phases Completed

- ✅ **Phase 1.1**: Project Foundation - Go module, Cobra CLI, development tooling
- ✅ **Phase 1.2**: Authentication Foundation - Multi-method auth (API Key, Service Account, OAuth2)
- ✅ **Phase 1.3**: Core TTS Integration - Google Cloud TTS with voice customization
- ✅ **Phase 1.4**: Input/Output Processing - Cross-platform audio, file handling, SSML security
- ✅ **Phase 1.5**: Configuration Management - Hierarchical YAML config with validation
- ✅ **Phase 1.6**: Performance Optimization - Caching, connection pooling, monitoring
- ✅ **Phase 1.7**: Testing Foundation - Comprehensive test coverage and reporting

### Getting Started

#### Quick Installation
```bash
# macOS Apple Silicon
curl -L -o assistant-cli https://github.com/mikefarmer/assistant-cli/releases/download/v1.0.0/assistant-cli-darwin-arm64
chmod +x assistant-cli

# Verify installation
./assistant-cli --version
```

#### Quick Usage
```bash
# Set up authentication
export ASSISTANT_CLI_API_KEY="your-google-cloud-api-key"

# Basic text-to-speech
echo "Hello, World!" | ./assistant-cli synthesize -o hello.mp3 --play

# Advanced usage
echo "Welcome to Assistant CLI" | ./assistant-cli synthesize \
  --voice en-US-Wavenet-C \
  --speed 1.2 \
  --pitch -2.0 \
  --format LINEAR16 \
  --output welcome.wav \
  --play
```

This release represents a significant milestone - Assistant-CLI is now production-ready for text-to-speech operations with enterprise-grade features, comprehensive testing, and native cross-platform support.