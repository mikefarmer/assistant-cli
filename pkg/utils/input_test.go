package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInputProcessor(t *testing.T) {
	reader := strings.NewReader("test input")
	processor := NewInputProcessor(reader)
	
	assert.NotNil(t, processor)
	assert.Equal(t, MaxTextLength, processor.maxLength)
	assert.Equal(t, reader, processor.reader)
}

func TestNewInputProcessorWithLimit(t *testing.T) {
	reader := strings.NewReader("test input")
	customLimit := 1000
	processor := NewInputProcessorWithLimit(reader, customLimit)
	
	assert.NotNil(t, processor)
	assert.Equal(t, customLimit, processor.maxLength)
	assert.Equal(t, reader, processor.reader)
}

func TestInputProcessor_ReadText_Success(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple text", "Hello, World!", "Hello, World!"},
		{"multiline text", "Line 1\nLine 2\nLine 3", "Line 1\nLine 2\nLine 3"},
		{"text with spaces", "  Text with spaces  ", "  Text with spaces  "},
		{"unicode text", "Hello 世界! 🌍", "Hello 世界! 🌍"},
		{"empty lines", "Line 1\n\nLine 3", "Line 1\n\nLine 3"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.input)
			processor := NewInputProcessor(reader)
			
			result, err := processor.ReadText()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestInputProcessor_ReadText_EmptyInput(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"completely empty", ""},
		{"only whitespace", "   \n\t  \n  "},
		{"only newlines", "\n\n\n"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.input)
			processor := NewInputProcessor(reader)
			
			result, err := processor.ReadText()
			require.Error(t, err)
			
			var inputErr *InputError
			assert.ErrorAs(t, err, &inputErr)
			assert.Equal(t, "empty", inputErr.Type)
			assert.Empty(t, result)
		})
	}
}

func TestInputProcessor_ReadText_TooLong(t *testing.T) {
	// Create input longer than default limit
	longInput := strings.Repeat("a", MaxTextLength+1)
	reader := strings.NewReader(longInput)
	processor := NewInputProcessor(reader)
	
	result, err := processor.ReadText()
	require.Error(t, err)
	
	var inputErr *InputError
	assert.ErrorAs(t, err, &inputErr)
	assert.Equal(t, "length", inputErr.Type)
	assert.Empty(t, result)
}

func TestInputProcessor_ReadText_InvalidUTF8(t *testing.T) {
	// Create input with invalid UTF-8 sequences
	invalidUTF8 := string([]byte{0xFF, 0xFE, 0xFD})
	reader := strings.NewReader("valid text" + invalidUTF8)
	processor := NewInputProcessor(reader)
	
	result, err := processor.ReadText()
	require.Error(t, err)
	
	var inputErr *InputError
	assert.ErrorAs(t, err, &inputErr)
	assert.Equal(t, "encoding", inputErr.Type)
	assert.Empty(t, result)
}

func TestInputProcessor_ReadText_NilReader(t *testing.T) {
	processor := NewInputProcessor(nil)
	
	result, err := processor.ReadText()
	require.Error(t, err)
	
	var inputErr *InputError
	assert.ErrorAs(t, err, &inputErr)
	assert.Equal(t, "configuration", inputErr.Type)
	assert.Empty(t, result)
}

func TestInputProcessor_CleanText(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"remove null bytes", "Hello\x00World", "HelloWorld"},
		{"normalize line endings", "Line1\r\nLine2\rLine3\n", "Line1\nLine2\nLine3"},
		{"trim trailing spaces", "Line1  \t\nLine2\t  ", "Line1\nLine2"},
		{"reduce excessive blank lines", "Line1\n\n\n\n\nLine2", "Line1\n\n\nLine2"},
		{"trim overall whitespace", "  \n  Hello World  \n  ", "Hello World"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader("")
			processor := NewInputProcessor(reader)
			
			result := processor.CleanText(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestInputProcessor_SplitByLength(t *testing.T) {
	reader := strings.NewReader("")
	processor := NewInputProcessor(reader)
	
	testCases := []struct {
		name      string
		input     string
		maxLength int
		expected  []string
	}{
		{
			"short text",
			"Hello World",
			50,
			[]string{"Hello World"},
		},
		{
			"split at word boundary",
			"This is a long sentence that needs to be split",
			20,
			[]string{"This is a long", "sentence that needs", "to be split"},
		},
		{
			"split at sentence boundary",
			"First sentence. Second sentence. Third sentence.",
			20,
			[]string{"First sentence. ", "Second sentence. ", "Third sentence."},
		},
		{
			"force split long word",
			"supercalifragilisticexpialidocious",
			10,
			[]string{"supercalif", "ragilistic", "expialidoc", "ious"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := processor.SplitByLength(tc.input, tc.maxLength)
			assert.Equal(t, len(tc.expected), len(result), "Number of chunks should match")
			
			for i, expected := range tc.expected {
				assert.Equal(t, expected, result[i], "Chunk %d should match", i)
				assert.LessOrEqual(t, len(result[i]), tc.maxLength, "Chunk %d should not exceed max length", i)
			}
		})
	}
}

func TestInputProcessor_GetTextStats(t *testing.T) {
	reader := strings.NewReader("")
	processor := NewInputProcessor(reader)
	
	testCases := []struct {
		name           string
		input          string
		expectedChars  int
		expectedWords  int
		expectedLines  int
		expectedUTF    bool
	}{
		{
			"simple text",
			"Hello World",
			11, 2, 1, true,
		},
		{
			"multiline text",
			"Line 1\nLine 2\nLine 3",
			20, 6, 3, true,
		},
		{
			"unicode text",
			"Hello 世界!",
			9, 2, 1, true,
		},
		{
			"empty text",
			"",
			0, 0, 1, true, // empty string has 1 "line"
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stats := processor.GetTextStats(tc.input)
			
			assert.Equal(t, tc.expectedChars, stats.Characters)
			assert.Equal(t, tc.expectedWords, stats.Words)
			assert.Equal(t, tc.expectedLines, stats.Lines)
			assert.Equal(t, tc.expectedUTF, stats.IsValidUTF8)
			assert.Equal(t, len([]byte(tc.input)), stats.Bytes)
			
			// Test string representation
			str := stats.String()
			assert.Contains(t, str, "Characters:")
			assert.Contains(t, str, "Words:")
			assert.Contains(t, str, "Lines:")
		})
	}
}

func TestInputProcessor_checkProblematicChars(t *testing.T) {
	reader := strings.NewReader("")
	processor := NewInputProcessor(reader)
	
	testCases := []struct {
		name        string
		input       string
		expectError bool
		errorType   string
	}{
		{"clean text", "Hello World", false, ""},
		{"with tabs and newlines", "Hello\tWorld\n", false, ""},
		{"with null byte", "Hello\x00World", true, "characters"},
		{"excessive control chars", string([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}), true, "characters"},
		{"few control chars", "Hello\x01World", false, ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := processor.checkProblematicChars(tc.input)
			
			if tc.expectError {
				require.Error(t, err)
				var inputErr *InputError
				assert.ErrorAs(t, err, &inputErr)
				assert.Equal(t, tc.errorType, inputErr.Type)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputProcessor_validateText(t *testing.T) {
	reader := strings.NewReader("")
	processor := NewInputProcessor(reader)
	
	testCases := []struct {
		name        string
		input       string
		expectError bool
		errorType   string
	}{
		{"valid text", "Hello World", false, ""},
		{"empty text", "", true, "empty"},
		{"whitespace only", "   \n\t  ", true, "empty"},
		{"too long", strings.Repeat("a", MaxTextLength+1), true, "length"},
		{"invalid utf8", "Hello\xFF\xFEWorld", true, "encoding"},
		{"null byte", "Hello\x00World", true, "characters"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := processor.validateText(tc.input)
			
			if tc.expectError {
				require.Error(t, err)
				var inputErr *InputError
				assert.ErrorAs(t, err, &inputErr)
				assert.Equal(t, tc.errorType, inputErr.Type)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInputError_Error(t *testing.T) {
	testCases := []struct {
		name     string
		err      *InputError
		expected string
	}{
		{
			"error without input",
			&InputError{Type: "test", Message: "test message"},
			"input test: test message",
		},
		{
			"error with short input",
			&InputError{Type: "test", Message: "test message", Input: "short"},
			"input test: test message (input: \"short\")",
		},
		{
			"error with long input",
			&InputError{Type: "test", Message: "test message", Input: strings.Repeat("a", 100)},
			"input test: test message (input: \"" + strings.Repeat("a", 47) + "...\")",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.err.Error()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestInputProcessor_findSplitPoint(t *testing.T) {
	reader := strings.NewReader("")
	processor := NewInputProcessor(reader)
	
	testCases := []struct {
		name      string
		input     string
		maxLength int
		expected  int
	}{
		{"text shorter than max", "Hello", 10, 5},
		{"split at period", "Hello. World and more text", 10, 7},
		{"split at space", "Hello World and more", 12, 12},
		{"no good split point", "Verylongwordwithoutspaces", 10, 10},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := processor.findSplitPoint(tc.input, tc.maxLength)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkInputProcessor_ReadText(b *testing.B) {
	text := "Hello World! This is a test input for benchmarking the input processor."
	
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		processor := NewInputProcessor(reader)
		_, _ = processor.ReadText()
	}
}

func BenchmarkInputProcessor_CleanText(b *testing.B) {
	text := "Hello\r\nWorld\x00!\tThis\r\nis\n\n\n\na\n\n\ntest  \n  "
	reader := strings.NewReader("")
	processor := NewInputProcessor(reader)
	
	for i := 0; i < b.N; i++ {
		_ = processor.CleanText(text)
	}
}

func BenchmarkInputProcessor_GetTextStats(b *testing.B) {
	text := "Hello World! This is a test input with multiple lines\nand some unicode characters: 世界 🌍\nand various symbols."
	reader := strings.NewReader("")
	processor := NewInputProcessor(reader)
	
	for i := 0; i < b.N; i++ {
		_ = processor.GetTextStats(text)
	}
}