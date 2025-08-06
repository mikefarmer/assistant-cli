package player

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// AudioPlayer handles cross-platform audio playback
type AudioPlayer struct {
	// platform-specific player command
	player   string
	args     []string
	fallback bool
}

// PlayerError represents playback-related errors
type PlayerError struct {
	Operation string
	Err       error
	Platform  string
}

func (e *PlayerError) Error() string {
	return fmt.Sprintf("audio playback error on %s during %s: %v", e.Platform, e.Operation, e.Err)
}

func (e *PlayerError) Unwrap() error {
	return e.Err
}

// NewAudioPlayer creates a new audio player with platform detection
func NewAudioPlayer() (*AudioPlayer, error) {
	player := &AudioPlayer{}
	
	if err := player.detectPlayer(); err != nil {
		return nil, &PlayerError{
			Operation: "initialization",
			Err:       err,
			Platform:  runtime.GOOS,
		}
	}
	
	return player, nil
}

// detectPlayer detects the appropriate audio player for the current platform
func (p *AudioPlayer) detectPlayer() error {
	switch runtime.GOOS {
	case "darwin": // macOS
		if p.commandExists("afplay") {
			p.player = "afplay"
			p.args = []string{}
			return nil
		}
		// Fallback to system open command
		if p.commandExists("open") {
			p.player = "open"
			p.args = []string{"-a", "QuickTime Player"}
			p.fallback = true
			return nil
		}
		return fmt.Errorf("no suitable audio player found on macOS")
		
	case "linux":
		// Try common Linux audio players in order of preference
		players := []struct {
			cmd  string
			args []string
		}{
			{"aplay", []string{}},           // ALSA
			{"paplay", []string{}},          // PulseAudio
			{"ffplay", []string{"-nodisp", "-autoexit"}}, // ffmpeg
			{"mpv", []string{"--no-video"}}, // mpv
			{"mplayer", []string{}},         // mplayer
		}
		
		for _, player := range players {
			if p.commandExists(player.cmd) {
				p.player = player.cmd
				p.args = player.args
				if player.cmd != "aplay" && player.cmd != "paplay" {
					p.fallback = true
				}
				return nil
			}
		}
		return fmt.Errorf("no suitable audio player found on Linux")
		
	case "windows":
		// Use PowerShell's media capabilities
		if p.commandExists("powershell") {
			p.player = "powershell"
			p.args = []string{"-Command"}
			return nil
		}
		return fmt.Errorf("PowerShell not found on Windows")
		
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// commandExists checks if a command exists in PATH
func (p *AudioPlayer) commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// Play plays an audio file
func (p *AudioPlayer) Play(filePath string) error {
	if p.player == "" {
		return &PlayerError{
			Operation: "play",
			Err:       fmt.Errorf("no audio player configured"),
			Platform:  runtime.GOOS,
		}
	}
	
	// Validate file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &PlayerError{
			Operation: "play",
			Err:       fmt.Errorf("audio file does not exist: %s", filePath),
			Platform:  runtime.GOOS,
		}
	}
	
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		cmd = p.buildMacOSCommand(filePath)
	case "linux":
		cmd = p.buildLinuxCommand(filePath)
	case "windows":
		cmd = p.buildWindowsCommand(filePath)
	default:
		return &PlayerError{
			Operation: "play",
			Err:       fmt.Errorf("unsupported platform: %s", runtime.GOOS),
			Platform:  runtime.GOOS,
		}
	}
	
	// Execute the command
	if err := cmd.Run(); err != nil {
		return &PlayerError{
			Operation: "play",
			Err:       fmt.Errorf("failed to play audio: %v", err),
			Platform:  runtime.GOOS,
		}
	}
	
	return nil
}

// buildMacOSCommand builds the command for macOS
func (p *AudioPlayer) buildMacOSCommand(filePath string) *exec.Cmd {
	if p.player == "afplay" {
		return exec.Command(p.player, filePath)
	}
	// Fallback to open command
	args := append(p.args, filePath)
	return exec.Command(p.player, args...)
}

// buildLinuxCommand builds the command for Linux
func (p *AudioPlayer) buildLinuxCommand(filePath string) *exec.Cmd {
	args := append(p.args, filePath)
	return exec.Command(p.player, args...)
}

// buildWindowsCommand builds the command for Windows using PowerShell
func (p *AudioPlayer) buildWindowsCommand(filePath string) *exec.Cmd {
	// Convert to absolute path for PowerShell
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath // fallback to original path
	}
	
	// PowerShell script to play audio
	script := fmt.Sprintf(`
		Add-Type -AssemblyName presentationCore
		$mediaPlayer = New-Object system.windows.media.mediaplayer
		$mediaPlayer.open([uri]'%s')
		$mediaPlayer.Play()
		Start-Sleep 1
		do {
			Start-Sleep 1
		} while($mediaPlayer.NaturalDuration.HasTimeSpan -eq $false)
		Start-Sleep $mediaPlayer.NaturalDuration.TimeSpan.TotalSeconds
	`, absPath)
	
	args := append(p.args, script)
	return exec.Command(p.player, args...)
}

// GetPlayerInfo returns information about the detected audio player
func (p *AudioPlayer) GetPlayerInfo() PlayerInfo {
	return PlayerInfo{
		Command:  p.player,
		Args:     p.args,
		Platform: runtime.GOOS,
		Fallback: p.fallback,
	}
}

// PlayerInfo contains information about the audio player
type PlayerInfo struct {
	Command  string   `json:"command"`
	Args     []string `json:"args"`
	Platform string   `json:"platform"`
	Fallback bool     `json:"fallback"`
}

// IsSupported checks if audio playback is supported on the current platform
func IsSupported() bool {
	player, err := NewAudioPlayer()
	if err != nil {
		return false
	}
	return player.player != ""
}

// PlayFile is a convenience function to play an audio file
func PlayFile(filename string) error {
	player, err := NewAudioPlayer()
	if err != nil {
		return err
	}
	
	return player.Play(filename)
}