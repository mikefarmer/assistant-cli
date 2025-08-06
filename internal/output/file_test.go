package output

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileHandler(t *testing.T) {
	handler := NewFileHandler()
	
	assert.NotNil(t, handler)
	assert.Equal(t, ".", handler.baseDir)
	assert.True(t, handler.createDirs)
	assert.Equal(t, OverwriteBackup, handler.overwriteMode)
	assert.Equal(t, os.FileMode(0644), handler.filePermissions)
	assert.Equal(t, os.FileMode(0755), handler.dirPermissions)
}

func TestNewFileHandlerWithOptions(t *testing.T) {
	baseDir := "/tmp/test"
	handler := NewFileHandlerWithOptions(baseDir, false, OverwriteAlways)
	
	assert.NotNil(t, handler)
	assert.Equal(t, baseDir, handler.baseDir)
	assert.False(t, handler.createDirs)
	assert.Equal(t, OverwriteAlways, handler.overwriteMode)
}

func TestFileHandler_SetPermissions(t *testing.T) {
	handler := NewFileHandler()
	
	filePerms := os.FileMode(0600)
	dirPerms := os.FileMode(0700)
	
	handler.SetPermissions(filePerms, dirPerms)
	
	assert.Equal(t, filePerms, handler.filePermissions)
	assert.Equal(t, dirPerms, handler.dirPermissions)
}

func TestFileHandler_WriteFile_Success(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewFileHandlerWithOptions(tempDir, true, OverwriteAlways)
	
	testData := []byte("Hello, World!")
	filename := "test.txt"
	
	info, err := handler.WriteFile(filename, testData)
	require.NoError(t, err)
	require.NotNil(t, info)
	
	expectedPath := filepath.Join(tempDir, filename)
	assert.Equal(t, expectedPath, info.Path)
	assert.Equal(t, int64(len(testData)), info.Size)
	assert.False(t, info.Overwritten)
	assert.Empty(t, info.BackupPath)
	
	// Verify file was actually written
	writtenData, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, testData, writtenData)
}

func TestFileHandler_WriteFile_CreateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewFileHandlerWithOptions(tempDir, true, OverwriteAlways)
	
	testData := []byte("test data")
	filename := "subdir1/subdir2/test.txt"
	
	info, err := handler.WriteFile(filename, testData)
	require.NoError(t, err)
	require.NotNil(t, info)
	
	expectedPath := filepath.Join(tempDir, filename)
	assert.Equal(t, expectedPath, info.Path)
	
	// Verify directories were created
	assert.DirExists(t, filepath.Join(tempDir, "subdir1"))
	assert.DirExists(t, filepath.Join(tempDir, "subdir1", "subdir2"))
	
	// Verify file was written
	writtenData, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, testData, writtenData)
}

func TestFileHandler_WriteFile_NoCreateDirectories(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewFileHandlerWithOptions(tempDir, false, OverwriteAlways)
	
	testData := []byte("test data")
	filename := "nonexistent/test.txt"
	
	info, err := handler.WriteFile(filename, testData)
	require.Error(t, err)
	assert.Nil(t, info)
	
	var fileErr *FileError
	assert.ErrorAs(t, err, &fileErr)
	assert.Equal(t, "write", fileErr.Operation)
}

func TestFileHandler_WriteFile_OverwriteNever(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewFileHandlerWithOptions(tempDir, true, OverwriteNever)
	
	filename := "test.txt"
	filepath := filepath.Join(tempDir, filename)
	
	// Create existing file
	err := os.WriteFile(filepath, []byte("existing"), 0644)
	require.NoError(t, err)
	
	// Try to overwrite
	testData := []byte("new data")
	info, err := handler.WriteFile(filename, testData)
	
	require.Error(t, err)
	assert.Nil(t, info)
	
	var fileErr *FileError
	assert.ErrorAs(t, err, &fileErr)
	assert.Equal(t, "overwrite_check", fileErr.Operation)
	assert.Contains(t, err.Error(), "already exists")
}

func TestFileHandler_WriteFile_OverwriteBackup(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewFileHandlerWithOptions(tempDir, true, OverwriteBackup)
	
	filename := "test.txt"
	filePath := filepath.Join(tempDir, filename)
	originalData := []byte("original data")
	
	// Create existing file
	err := os.WriteFile(filePath, originalData, 0644)
	require.NoError(t, err)
	
	// Wait a moment to ensure different timestamps
	time.Sleep(100 * time.Millisecond)
	
	// Overwrite with backup
	newData := []byte("new data")
	info, err := handler.WriteFile(filename, newData)
	
	require.NoError(t, err)
	require.NotNil(t, info)
	
	assert.True(t, info.Overwritten)
	assert.NotEmpty(t, info.BackupPath)
	
	// Verify new file contents
	writtenData, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, newData, writtenData)
	
	// Verify backup file exists with original contents
	backupData, err := os.ReadFile(info.BackupPath)
	require.NoError(t, err)
	assert.Equal(t, originalData, backupData)
}

func TestFileHandler_validatePath(t *testing.T) {
	handler := NewFileHandler()
	
	testCases := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{"valid filename", "test.txt", false, ""},
		{"empty filename", "", true, "filename cannot be empty"},
		{"path traversal", "../../../etc/passwd", true, "path traversal not allowed"},
		{"windows path traversal", "..\\..\\windows\\system32", true, "path traversal not allowed"},
		{"executable extension", "test.exe", true, "file extension not allowed"},
		{"batch file", "script.bat", true, "file extension not allowed"},
		{"javascript file", "script.js", true, "file extension not allowed"},
		{"valid nested path", "audio/output.mp3", false, ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := handler.validatePath(tc.input)
			
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestFileHandler_validatePathSecurity_SystemDirectories(t *testing.T) {
	handler := NewFileHandler()
	
	testCases := []struct {
		path        string
		expectedMsg string
	}{
		{"/etc/passwd", "system directory not allowed"},
		{"/bin/sh", "system directory not allowed"},
		{"/usr/bin/something", "system directory not allowed"},
		{"C:\\Windows\\System32\\cmd.exe", "file extension not allowed"}, // .exe extension checked first
		{"C:\\Program Files\\malware.exe", "file extension not allowed"}, // .exe extension checked first
	}
	
	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			err := handler.validatePathSecurity(tc.path)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedMsg)
		})
	}
}

func TestGetSafeFilename(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		extension string
		expected  string
	}{
		{"simple text", "Hello World", "mp3", "Hello_World.mp3"},
		{"with special chars", "Hello, World! How are you?", "wav", "Hello_World_How_are_you.wav"},
		{"very long text", string(make([]byte, 200)), "txt", "output.txt"}, // Should fallback to "output"
		{"empty input", "", "mp3", "output.mp3"},
		{"only special chars", "!@#$%^&*()", "mp3", "output.mp3"},
		{"mixed valid/invalid", "Hello123_world-test.final", "mp3", "Hello123_world-test.final.mp3"},
		{"extension with dot", "test", ".mp3", "test.mp3"},
		{"no extension", "test", "", "test"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetSafeFilename(tc.input, tc.extension)
			assert.Equal(t, tc.expected, result)
			
			// Verify result doesn't contain problematic characters
			assert.NotContains(t, result, "/")
			assert.NotContains(t, result, "\\")
			assert.NotContains(t, result, ":")
			assert.NotContains(t, result, "*")
			assert.NotContains(t, result, "?")
			assert.NotContains(t, result, "\"")
			assert.NotContains(t, result, "<")
			assert.NotContains(t, result, ">")
			assert.NotContains(t, result, "|")
		})
	}
}

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test existing file
	existingFile := filepath.Join(tempDir, "exists.txt")
	err := os.WriteFile(existingFile, []byte("test"), 0644)
	require.NoError(t, err)
	
	assert.True(t, FileExists(existingFile))
	
	// Test non-existing file
	nonExistingFile := filepath.Join(tempDir, "does_not_exist.txt")
	assert.False(t, FileExists(nonExistingFile))
}

func TestGetFileSize(t *testing.T) {
	tempDir := t.TempDir()
	
	testData := []byte("Hello, World! This is a test file.")
	testFile := filepath.Join(tempDir, "test.txt")
	
	err := os.WriteFile(testFile, testData, 0644)
	require.NoError(t, err)
	
	size, err := GetFileSize(testFile)
	require.NoError(t, err)
	assert.Equal(t, int64(len(testData)), size)
	
	// Test non-existing file
	_, err = GetFileSize(filepath.Join(tempDir, "nonexistent.txt"))
	assert.Error(t, err)
}

func TestGenerateUniqueFilename(t *testing.T) {
	tempDir := t.TempDir()
	
	// Test with non-existing file
	basePath := filepath.Join(tempDir, "test.txt")
	result := GenerateUniqueFilename(basePath)
	assert.Equal(t, basePath, result)
	
	// Create the base file
	err := os.WriteFile(basePath, []byte("test"), 0644)
	require.NoError(t, err)
	
	// Generate unique filename
	uniquePath := GenerateUniqueFilename(basePath)
	assert.NotEqual(t, basePath, uniquePath)
	assert.Contains(t, uniquePath, "test_1.txt")
	
	// Create the first alternative
	err = os.WriteFile(uniquePath, []byte("test"), 0644)
	require.NoError(t, err)
	
	// Generate another unique filename
	uniquePath2 := GenerateUniqueFilename(basePath)
	assert.NotEqual(t, basePath, uniquePath2)
	assert.NotEqual(t, uniquePath, uniquePath2)
	assert.Contains(t, uniquePath2, "test_2.txt")
}

func TestFileHandler_WriteFileStream(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewFileHandlerWithOptions(tempDir, true, OverwriteAlways)
	
	testData := []byte("Stream test data")
	filename := "stream_test.txt"
	
	info, err := handler.WriteFileStream(filename, testData, false)
	require.NoError(t, err)
	require.NotNil(t, info)
	
	expectedPath := filepath.Join(tempDir, filename)
	assert.Equal(t, expectedPath, info.Path)
	
	// Verify file contents
	writtenData, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, testData, writtenData)
}

func TestFileHandler_WriteFileStream_Append(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewFileHandlerWithOptions(tempDir, true, OverwriteAlways)
	
	filename := "append_test.txt"
	expectedPath := filepath.Join(tempDir, filename)
	
	// Write initial data
	initialData := []byte("Initial data\n")
	_, err := handler.WriteFileStream(filename, initialData, false)
	require.NoError(t, err)
	
	// Append more data
	appendData := []byte("Appended data\n")
	info, err := handler.WriteFileStream(filename, appendData, true)
	require.NoError(t, err)
	require.NotNil(t, info)
	
	// Verify combined contents
	expectedContent := append(initialData, appendData...)
	writtenData, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, expectedContent, writtenData)
}

func TestFileError_Error(t *testing.T) {
	err := &FileError{
		Operation: "write",
		Path:      "/test/path/file.txt",
		Err:       assert.AnError,
	}
	
	result := err.Error()
	assert.Contains(t, result, "file write error for /test/path/file.txt")
	assert.Equal(t, assert.AnError, err.Unwrap())
}

func TestFileHandler_ensureDirectoryExists(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewFileHandlerWithOptions(tempDir, true, OverwriteAlways)
	
	testCases := []struct {
		name        string
		dir         string
		expectError bool
	}{
		{"empty dir", "", false},
		{"current dir", ".", false},
		{"simple subdir", filepath.Join(tempDir, "subdir"), false},
		{"nested subdirs", filepath.Join(tempDir, "sub1", "sub2", "sub3"), false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := handler.ensureDirectoryExists(tc.dir)
			
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.dir != "" && tc.dir != "." {
					assert.DirExists(t, tc.dir)
				}
			}
		})
	}
}

func TestFileHandler_ensureDirectoryExists_FileExists(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewFileHandlerWithOptions(tempDir, true, OverwriteAlways)
	
	// Create a file where we want to create a directory
	filePath := filepath.Join(tempDir, "not_a_directory")
	err := os.WriteFile(filePath, []byte("test"), 0644)
	require.NoError(t, err)
	
	err = handler.ensureDirectoryExists(filePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path exists but is not a directory")
}

// Benchmark tests
func BenchmarkFileHandler_WriteFile(b *testing.B) {
	tempDir := b.TempDir()
	handler := NewFileHandlerWithOptions(tempDir, true, OverwriteAlways)
	testData := []byte("Benchmark test data for file writing performance")
	
	for i := 0; i < b.N; i++ {
		filename := filepath.Join("bench", "file_"+string(rune(i))+".txt")
		_, _ = handler.WriteFile(filename, testData)
	}
}

func BenchmarkGetSafeFilename(b *testing.B) {
	input := "Hello, World! This is a test filename with special characters: @#$%^&*()"
	
	for i := 0; i < b.N; i++ {
		_ = GetSafeFilename(input, "txt")
	}
}