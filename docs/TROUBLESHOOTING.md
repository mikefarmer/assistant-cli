# Troubleshooting Guide

This guide helps resolve common issues with the Assistant-CLI tool.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Authentication Problems](#authentication-problems)
- [Text-to-Speech Issues](#text-to-speech-issues)
- [Audio Playback Problems](#audio-playback-problems)
- [Configuration Issues](#configuration-issues)
- [Performance Issues](#performance-issues)
- [Build and Development Issues](#build-and-development-issues)
- [Getting Help](#getting-help)

## Installation Issues

### Binary Not Found After Installation

**Problem**: `assistant-cli: command not found`

**Solutions**:

1. **Check if binary is in PATH**:
   ```bash
   which assistant-cli
   ```

2. **Add to PATH** (if installed locally):
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export PATH="/path/to/assistant-cli:$PATH"
   ```

3. **Use absolute path**:
   ```bash
   /usr/local/bin/assistant-cli --help
   ```

### Permission Denied

**Problem**: `permission denied: ./assistant-cli`

**Solutions**:

1. **Make executable**:
   ```bash
   chmod +x assistant-cli
   ```

2. **Check file permissions**:
   ```bash
   ls -la assistant-cli
   ```

### macOS Security Warning

**Problem**: "assistant-cli cannot be opened because it is from an unidentified developer"

**Solutions**:

1. **Allow in System Preferences**:
   - Go to System Preferences → Security & Privacy → General
   - Click "Allow Anyway" for assistant-cli

2. **Command line bypass**:
   ```bash
   sudo xattr -rd com.apple.quarantine assistant-cli
   ```

3. **Manual verification**:
   ```bash
   spctl --assess --type execute assistant-cli
   ```

## Authentication Problems

### API Key Issues

**Problem**: "Invalid API key" or authentication failures

**Solutions**:

1. **Verify API key format**:
   ```bash
   echo $ASSISTANT_CLI_API_KEY
   # Should be: AIzaSy...
   ```

2. **Check API key permissions**:
   - Ensure Text-to-Speech API is enabled
   - Verify API key restrictions in Google Cloud Console

3. **Test authentication**:
   ```bash
   assistant-cli login --validate
   ```

4. **Regenerate API key**:
   - Go to Google Cloud Console → APIs & Services → Credentials
   - Delete old key and create new one

### Service Account Issues

**Problem**: "Service account authentication failed"

**Solutions**:

1. **Verify JSON file**:
   ```bash
   cat /path/to/service-account.json | jq .
   ```

2. **Check file permissions**:
   ```bash
   ls -la /path/to/service-account.json
   chmod 600 /path/to/service-account.json
   ```

3. **Validate service account**:
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
   assistant-cli login --validate
   ```

4. **Check service account roles**:
   - Ensure "Cloud Text-to-Speech User" role is assigned
   - Verify service account is enabled

### OAuth2 Issues

**Problem**: OAuth2 flow fails or tokens are invalid

**Solutions**:

1. **Clear cached tokens**:
   ```bash
   rm ~/.assistant-cli-token.json
   assistant-cli login --method oauth2 --force
   ```

2. **Check OAuth2 credentials**:
   ```bash
   echo $ASSISTANT_CLI_OAUTH2_CLIENT_ID
   echo $ASSISTANT_CLI_OAUTH2_CLIENT_SECRET
   ```

3. **Verify redirect URI**:
   - Must be: `http://localhost:8080/callback`
   - Check Google Cloud Console OAuth2 settings

4. **Port conflicts**:
   ```bash
   # Check if port 8080 is in use
   lsof -i :8080
   ```

## Text-to-Speech Issues

### API Request Failures

**Problem**: "Request failed" or timeout errors

**Solutions**:

1. **Check API quotas**:
   - Google Cloud Console → APIs & Services → Quotas
   - Verify daily/monthly limits

2. **Test with simple input**:
   ```bash
   echo "Hello" | assistant-cli synthesize -o test.mp3
   ```

3. **Check network connectivity**:
   ```bash
   ping texttospeech.googleapis.com
   ```

4. **Enable verbose logging**:
   ```bash
   echo "Test" | assistant-cli synthesize -o test.mp3 --verbose
   ```

### Voice Not Available

**Problem**: "Voice not found" or invalid voice errors

**Solutions**:

1. **List available voices**:
   ```bash
   assistant-cli synthesize --list-voices --language en-US
   ```

2. **Check voice format**:
   ```bash
   # Correct format
   assistant-cli synthesize --voice en-US-Wavenet-D
   
   # Incorrect format
   assistant-cli synthesize --voice wavenet-d
   ```

3. **Try default voice**:
   ```bash
   echo "Test" | assistant-cli synthesize -o test.mp3
   ```

### SSML Validation Errors

**Problem**: "Invalid SSML" or markup errors

**Solutions**:

1. **Test with plain text first**:
   ```bash
   echo "Plain text" | assistant-cli synthesize -o test.mp3
   ```

2. **Check SSML syntax**:
   ```bash
   echo '<speak>Hello <break time="1s"/> World</speak>' | assistant-cli synthesize -o test.mp3
   ```

3. **Enable SSML validation debug**:
   ```bash
   echo '<speak>Test</speak>' | assistant-cli synthesize -o test.mp3 --debug
   ```

## Audio Playback Problems

### No Audio Output

**Problem**: Audio file created but no sound during playback

**Solutions**:

1. **Check audio file**:
   ```bash
   file output.mp3
   ls -la output.mp3
   ```

2. **Test manual playback**:
   ```bash
   # macOS
   afplay output.mp3
   
   # Linux
   aplay output.mp3
   
   # Windows
   powershell -c "(New-Object Media.SoundPlayer 'output.mp3').PlaySync()"
   ```

3. **Check system volume**:
   - Ensure system volume is not muted
   - Test with other audio files

### Audio Player Not Found

**Problem**: "Audio player not found" errors

**Solutions**:

1. **macOS**: Ensure afplay is available:
   ```bash
   which afplay
   ```

2. **Linux**: Install audio player:
   ```bash
   # Ubuntu/Debian
   sudo apt install alsa-utils
   
   # RHEL/CentOS
   sudo yum install alsa-utils
   ```

3. **Manual player specification**:
   ```bash
   echo "Test" | assistant-cli synthesize -o test.mp3
   # Then manually play with your preferred player
   ```

## Configuration Issues

### Config File Not Found

**Problem**: Configuration file not loaded

**Solutions**:

1. **Check config file location**:
   ```bash
   assistant-cli config show --show-sources
   ```

2. **Create default config**:
   ```bash
   assistant-cli config generate ~/.assistant-cli.yaml
   ```

3. **Specify config file**:
   ```bash
   assistant-cli --config /path/to/config.yaml synthesize --help
   ```

### Invalid Configuration

**Problem**: Configuration validation errors

**Solutions**:

1. **Validate config file**:
   ```bash
   assistant-cli config validate ~/.assistant-cli.yaml
   ```

2. **Check YAML syntax**:
   ```bash
   cat ~/.assistant-cli.yaml | yaml-lint
   ```

3. **Reset to defaults**:
   ```bash
   mv ~/.assistant-cli.yaml ~/.assistant-cli.yaml.backup
   assistant-cli config generate ~/.assistant-cli.yaml
   ```

## Performance Issues

### Slow Startup

**Problem**: CLI takes long time to start

**Solutions**:

1. **Check disk space**:
   ```bash
   df -h
   ```

2. **Clear cache**:
   ```bash
   rm -rf ~/.cache/assistant-cli/
   ```

3. **Profile startup**:
   ```bash
   time assistant-cli --help
   ```

### High Memory Usage

**Problem**: Excessive memory consumption

**Solutions**:

1. **Monitor memory**:
   ```bash
   # During operation
   ps aux | grep assistant-cli
   ```

2. **Process smaller chunks**:
   ```bash
   # Instead of large text files, process in smaller pieces
   head -100 largefile.txt | assistant-cli synthesize -o part1.mp3
   ```

3. **Check for memory leaks**:
   ```bash
   # Run multiple operations and monitor
   for i in {1..10}; do
     echo "Test $i" | assistant-cli synthesize -o test$i.mp3
   done
   ```

## Build and Development Issues

### Build Failures

**Problem**: `go build` fails

**Solutions**:

1. **Check Go version**:
   ```bash
   go version
   # Requires Go 1.23+
   ```

2. **Update dependencies**:
   ```bash
   go mod download
   go mod tidy
   ```

3. **Clean build cache**:
   ```bash
   go clean -modcache
   go clean -cache
   ```

4. **Verbose build**:
   ```bash
   go build -v -x main.go
   ```

### Test Failures

**Problem**: `go test` fails

**Solutions**:

1. **Run specific test**:
   ```bash
   go test -v ./internal/auth
   ```

2. **Skip integration tests**:
   ```bash
   go test -short ./...
   ```

3. **Check test coverage**:
   ```bash
   go test -cover ./...
   ```

### Import Issues

**Problem**: Import path errors

**Solutions**:

1. **Check module path**:
   ```bash
   grep module go.mod
   ```

2. **Update imports**:
   ```bash
   go mod tidy
   ```

3. **Clear module cache**:
   ```bash
   go clean -modcache
   ```

## Getting Help

### Enable Debug Mode

For any issue, enable verbose/debug output:

```bash
# Add --verbose flag to any command
assistant-cli --verbose synthesize --help
echo "test" | assistant-cli --verbose synthesize -o test.mp3
```

### Check Version and Build Info

```bash
assistant-cli --version
assistant-cli --help
```

### Collect System Information

```bash
# System info
uname -a
go version

# Assistant-CLI info
assistant-cli --version
assistant-cli config show

# Authentication status
assistant-cli login --validate
```

### Common Diagnostic Commands

```bash
# Test basic functionality
echo "Hello World" | assistant-cli synthesize -o test.mp3

# Test authentication
assistant-cli login --validate

# Test configuration
assistant-cli config show

# Test voice listing
assistant-cli synthesize --list-voices --language en-US

# Test audio playback
assistant-cli synthesize --help | grep -i play
```

### Log Files and Debugging

1. **Enable verbose mode**: Use `--verbose` flag
2. **Check error messages**: Read full error output
3. **Test incrementally**: Start with simple commands
4. **Isolate issues**: Test each component separately

### Reporting Issues

When reporting issues, please include:

1. **Version information**:
   ```bash
   assistant-cli --version
   ```

2. **Operating system**:
   ```bash
   uname -a
   ```

3. **Command that failed**:
   ```bash
   # Exact command with --verbose flag
   echo "test" | assistant-cli --verbose synthesize -o test.mp3
   ```

4. **Full error output**
5. **Configuration** (sanitized, no secrets):
   ```bash
   assistant-cli config show
   ```

### Support Channels

- **GitHub Issues**: [Report bugs and feature requests](https://github.com/mikefarmer/assistant-cli/issues)
- **GitHub Discussions**: [Ask questions and get help](https://github.com/mikefarmer/assistant-cli/discussions)
- **Documentation**: Check the `/docs` directory for additional guides

---

## Quick Fixes Summary

| Issue | Quick Fix |
|-------|-----------|
| Command not found | `chmod +x assistant-cli` |
| API key invalid | `assistant-cli login --validate` |
| Config not found | `assistant-cli config generate` |
| No audio output | Check system volume, test with `afplay output.mp3` |
| Build fails | `go mod tidy && go build` |
| Tests fail | `go test -short ./...` |
| Slow startup | Clear cache, check disk space |
| SSML errors | Test with plain text first |

---

*For additional help, check the main [README.md](../README.md) and [DEVELOPMENT.md](DEVELOPMENT.md) guides.*