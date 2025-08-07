package output

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileHandler manages safe file output operations
type FileHandler struct {
	baseDir         string
	createDirs      bool
	overwriteMode   OverwriteMode
	filePermissions fs.FileMode
	dirPermissions  fs.FileMode
}

// OverwriteMode defines how to handle existing files
type OverwriteMode int

const (
	OverwriteNever  OverwriteMode = iota // Never overwrite existing files
	OverwriteAlways                      // Always overwrite existing files
	OverwritePrompt                      // Prompt user for confirmation
	OverwriteBackup                      // Create backup before overwriting
)

// FileError represents file operation errors
type FileError struct {
	Operation string
	Path      string
	Err       error
}

func (e *FileError) Error() string {
	return fmt.Sprintf("file %s error for %s: %v", e.Operation, e.Path, e.Err)
}

func (e *FileError) Unwrap() error {
	return e.Err
}

// FileInfo contains information about a written file
type FileInfo struct {
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Created     time.Time `json:"created"`
	Overwritten bool      `json:"overwritten"`
	BackupPath  string    `json:"backup_path,omitempty"`
	Permissions string    `json:"permissions"`
}

// NewFileHandler creates a new file handler with default settings
func NewFileHandler() *FileHandler {
	return &FileHandler{
		baseDir:         ".",
		createDirs:      true,
		overwriteMode:   OverwriteBackup,
		filePermissions: 0644,
		dirPermissions:  0755,
	}
}

// NewFileHandlerWithOptions creates a new file handler with custom options
func NewFileHandlerWithOptions(baseDir string, createDirs bool, mode OverwriteMode) *FileHandler {
	return &FileHandler{
		baseDir:         baseDir,
		createDirs:      createDirs,
		overwriteMode:   mode,
		filePermissions: 0644,
		dirPermissions:  0755,
	}
}

// SetPermissions sets file and directory permissions
func (h *FileHandler) SetPermissions(filePerms, dirPerms fs.FileMode) {
	h.filePermissions = filePerms
	h.dirPermissions = dirPerms
}

// WriteFile writes data to a file with safety checks
func (h *FileHandler) WriteFile(filename string, data []byte) (*FileInfo, error) {
	// Validate and sanitize filename
	safePath, err := h.validatePath(filename)
	if err != nil {
		return nil, &FileError{
			Operation: "validation",
			Path:      filename,
			Err:       err,
		}
	}

	// Create directories if needed
	if h.createDirs {
		if dirErr := h.ensureDirectoryExists(filepath.Dir(safePath)); dirErr != nil {
			return nil, &FileError{
				Operation: "directory_creation",
				Path:      safePath,
				Err:       dirErr,
			}
		}
	}

	// Handle existing file
	info, err := h.handleExistingFile(safePath)
	if err != nil {
		return nil, err
	}

	// Write the file
	if writeErr := os.WriteFile(safePath, data, h.filePermissions); writeErr != nil {
		return nil, &FileError{
			Operation: "write",
			Path:      safePath,
			Err:       writeErr,
		}
	}

	// Get file stats
	stat, err := os.Stat(safePath)
	if err != nil {
		// File was written but we can't stat it
		return &FileInfo{
			Path:        safePath,
			Size:        int64(len(data)),
			Created:     time.Now(),
			Overwritten: info.Overwritten,
			BackupPath:  info.BackupPath,
			Permissions: h.filePermissions.String(),
		}, nil
	}

	return &FileInfo{
		Path:        safePath,
		Size:        stat.Size(),
		Created:     stat.ModTime(),
		Overwritten: info.Overwritten,
		BackupPath:  info.BackupPath,
		Permissions: stat.Mode().String(),
	}, nil
}

// WriteFileStream writes data from a stream to a file
func (h *FileHandler) WriteFileStream(filename string, data []byte, append bool) (*FileInfo, error) {
	// Validate path
	safePath, err := h.validatePath(filename)
	if err != nil {
		return nil, &FileError{
			Operation: "validation",
			Path:      filename,
			Err:       err,
		}
	}

	// Create directories if needed
	if h.createDirs {
		if dirCreateErr := h.ensureDirectoryExists(filepath.Dir(safePath)); dirCreateErr != nil {
			return nil, &FileError{
				Operation: "directory_creation",
				Path:      safePath,
				Err:       dirCreateErr,
			}
		}
	}

	flags := os.O_CREATE | os.O_WRONLY
	if append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC

		// Handle existing file for non-append mode
		info, handleErr := h.handleExistingFile(safePath)
		if handleErr != nil {
			return nil, handleErr
		}
		// Store backup info for later use
		_ = info
	}

	file, err := os.OpenFile(safePath, flags, h.filePermissions)
	if err != nil {
		return nil, &FileError{
			Operation: "open",
			Path:      safePath,
			Err:       err,
		}
	}
	defer file.Close()

	written, err := file.Write(data)
	if err != nil {
		return nil, &FileError{
			Operation: "write",
			Path:      safePath,
			Err:       err,
		}
	}

	// Get file stats
	stat, err := file.Stat()
	if err != nil {
		return &FileInfo{
			Path:        safePath,
			Size:        int64(written),
			Created:     time.Now(),
			Permissions: h.filePermissions.String(),
		}, nil
	}

	return &FileInfo{
		Path:        safePath,
		Size:        stat.Size(),
		Created:     stat.ModTime(),
		Permissions: stat.Mode().String(),
	}, nil
}

// validatePath validates and sanitizes file path
func (h *FileHandler) validatePath(filename string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("filename cannot be empty")
	}

	// Clean the path to remove any .. or . components
	cleaned := filepath.Clean(filename)

	// Prevent directory traversal attacks
	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("path traversal not allowed: %s", filename)
	}

	// Make it relative to base directory if it's not absolute
	if !filepath.IsAbs(cleaned) {
		cleaned = filepath.Join(h.baseDir, cleaned)
	}

	// Additional security checks
	if err := h.validatePathSecurity(cleaned); err != nil {
		return "", err
	}

	return cleaned, nil
}

// validatePathSecurity performs additional security validation
func (h *FileHandler) validatePathSecurity(path string) error {
	// Check for prohibited file extensions or patterns
	prohibitedExtensions := []string{
		".exe", ".bat", ".cmd", ".com", ".scr", ".pif",
		".vbs", ".vbe", ".js", ".jse", ".wsf", ".wsh",
		".msc", ".cpl", ".dll", ".sys",
	}

	ext := strings.ToLower(filepath.Ext(path))
	for _, prohibited := range prohibitedExtensions {
		if ext == prohibited {
			return fmt.Errorf("file extension not allowed: %s", ext)
		}
	}

	// Check for system directories (Unix/Linux)
	prohibitedPaths := []string{
		"/etc/", "/bin/", "/sbin/", "/usr/bin/", "/usr/sbin/",
		"/var/log/", "/proc/", "/sys/", "/dev/",
	}

	for _, prohibited := range prohibitedPaths {
		if strings.HasPrefix(path, prohibited) {
			return fmt.Errorf("access to system directory not allowed: %s", prohibited)
		}
	}

	// Windows system directories
	if len(path) >= 3 && path[1] == ':' {
		winProhibited := []string{
			"C:\\Windows\\", "C:\\Program Files\\", "C:\\Program Files (x86)\\",
			"C:\\System32\\", "C:\\SysWOW64\\",
		}

		upperPath := strings.ToUpper(path)
		for _, prohibited := range winProhibited {
			if strings.HasPrefix(upperPath, strings.ToUpper(prohibited)) {
				return fmt.Errorf("access to system directory not allowed: %s", prohibited)
			}
		}
	}

	return nil
}

// ensureDirectoryExists creates directory if it doesn't exist
func (h *FileHandler) ensureDirectoryExists(dir string) error {
	if dir == "" || dir == "." {
		return nil
	}

	// Check if directory already exists
	if stat, err := os.Stat(dir); err == nil {
		if !stat.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", dir)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check directory: %v", err)
	}

	// Create directory with appropriate permissions
	if err := os.MkdirAll(dir, h.dirPermissions); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	return nil
}

// handleExistingFile handles existing files based on overwrite mode
func (h *FileHandler) handleExistingFile(path string) (*FileInfo, error) {
	info := &FileInfo{Path: path}

	// Check if file exists
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		// File doesn't exist, nothing to handle
		return info, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check file: %v", err)
	}

	// File exists, handle based on mode
	switch h.overwriteMode {
	case OverwriteNever:
		return nil, &FileError{
			Operation: "overwrite_check",
			Path:      path,
			Err:       fmt.Errorf("file already exists and overwrite is disabled"),
		}

	case OverwriteAlways:
		info.Overwritten = true
		return info, nil

	case OverwritePrompt:
		// For CLI tools, we'll treat this as "never" for safety
		// In a more sophisticated implementation, you could add actual prompting
		return nil, &FileError{
			Operation: "overwrite_check",
			Path:      path,
			Err:       fmt.Errorf("file already exists, user confirmation required"),
		}

	case OverwriteBackup:
		backupPath, err := h.createBackup(path, stat)
		if err != nil {
			return nil, &FileError{
				Operation: "backup",
				Path:      path,
				Err:       err,
			}
		}
		info.Overwritten = true
		info.BackupPath = backupPath
		return info, nil

	default:
		return nil, &FileError{
			Operation: "overwrite_check",
			Path:      path,
			Err:       fmt.Errorf("unknown overwrite mode"),
		}
	}
}

// createBackup creates a backup of an existing file
func (h *FileHandler) createBackup(originalPath string, stat fs.FileInfo) (string, error) {
	timestamp := stat.ModTime().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.backup_%s", originalPath, timestamp)

	// Ensure backup path doesn't exist (avoid collisions)
	counter := 1
	for {
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			break
		}
		backupPath = fmt.Sprintf("%s.backup_%s_%d", originalPath, timestamp, counter)
		counter++
		if counter > 1000 {
			return "", fmt.Errorf("too many backup files, cannot create backup")
		}
	}

	// Copy original file to backup location
	originalData, err := os.ReadFile(originalPath)
	if err != nil {
		return "", fmt.Errorf("failed to read original file for backup: %v", err)
	}

	if err := os.WriteFile(backupPath, originalData, stat.Mode()); err != nil {
		return "", fmt.Errorf("failed to create backup file: %v", err)
	}

	return backupPath, nil
}

// GetSafeFilename generates a safe filename from input text
func GetSafeFilename(input, extension string) string {
	// Remove or replace problematic characters
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_' || r == '.':
			return r
		case r == ' ':
			return '_'
		default:
			return -1 // Remove character
		}
	}, input)

	// Trim and limit length
	safe = strings.Trim(safe, "_.-")
	if len(safe) > 100 {
		safe = safe[:100]
	}

	// Ensure we have something
	if safe == "" {
		safe = "output"
	}

	// Add extension if provided
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	return safe + extension
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) (int64, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

// GenerateUniqueFilename generates a unique filename by appending a number
func GenerateUniqueFilename(basePath string) string {
	if !FileExists(basePath) {
		return basePath
	}

	dir := filepath.Dir(basePath)
	filename := filepath.Base(basePath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	counter := 1
	for {
		newFilename := fmt.Sprintf("%s_%d%s", nameWithoutExt, counter, ext)
		newPath := filepath.Join(dir, newFilename)

		if !FileExists(newPath) {
			return newPath
		}

		counter++
		if counter > 10000 {
			// Safety valve to prevent infinite loops
			return filepath.Join(dir, fmt.Sprintf("%s_%d%s", nameWithoutExt, int(time.Now().Unix()), ext))
		}
	}
}
