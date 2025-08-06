package utils

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const (
	// MaxTextLength defines the maximum allowed text length for processing
	MaxTextLength = 5000
	// BufferSize defines the buffer size for reading input
	BufferSize = 4096
)

// InputProcessor handles text input processing and validation
type InputProcessor struct {
	maxLength int
	reader    io.Reader
}

// InputError represents input-related errors
type InputError struct {
	Type    string
	Message string
	Input   string
}

func (e *InputError) Error() string {
	if e.Input != "" {
		// Truncate input in error message for readability
		input := e.Input
		if len(input) > 50 {
			input = input[:47] + "..."
		}
		return fmt.Sprintf("input %s: %s (input: %q)", e.Type, e.Message, input)
	}
	return fmt.Sprintf("input %s: %s", e.Type, e.Message)
}

// NewInputProcessor creates a new input processor with default settings
func NewInputProcessor(reader io.Reader) *InputProcessor {
	return &InputProcessor{
		maxLength: MaxTextLength,
		reader:    reader,
	}
}

// NewInputProcessorWithLimit creates a new input processor with custom length limit
func NewInputProcessorWithLimit(reader io.Reader, maxLength int) *InputProcessor {
	return &InputProcessor{
		maxLength: maxLength,
		reader:    reader,
	}
}

// ReadText reads and validates text from the input source
func (p *InputProcessor) ReadText() (string, error) {
	if p.reader == nil {
		return "", &InputError{
			Type:    "configuration",
			Message: "no input reader configured",
		}
	}
	
	// Read input with buffering
	var buffer strings.Builder
	scanner := bufio.NewScanner(p.reader)
	scanner.Buffer(make([]byte, BufferSize), p.maxLength+1)
	
	// Read all lines
	for scanner.Scan() {
		if buffer.Len() > 0 {
			buffer.WriteString("\n")
		}
		buffer.WriteString(scanner.Text())
		
		// Check length limit during reading
		if buffer.Len() > p.maxLength {
			return "", &InputError{
				Type:    "length",
				Message: fmt.Sprintf("input exceeds maximum length of %d characters", p.maxLength),
			}
		}
	}
	
	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return "", &InputError{
			Type:    "read",
			Message: fmt.Sprintf("failed to read input: %v", err),
		}
	}
	
	text := buffer.String()
	
	// Validate the resulting text
	if err := p.validateText(text); err != nil {
		return "", err
	}
	
	return text, nil
}

// ReadTextWithPrompt reads text with a user prompt (for interactive mode)
func (p *InputProcessor) ReadTextWithPrompt(prompt string) (string, error) {
	fmt.Print(prompt)
	return p.ReadText()
}

// validateText performs comprehensive text validation
func (p *InputProcessor) validateText(text string) error {
	// Check if empty
	if strings.TrimSpace(text) == "" {
		return &InputError{
			Type:    "empty",
			Message: "input text is empty or contains only whitespace",
		}
	}
	
	// Check length
	if len(text) > p.maxLength {
		return &InputError{
			Type:    "length",
			Message: fmt.Sprintf("input exceeds maximum length of %d characters", p.maxLength),
			Input:   text,
		}
	}
	
	// Validate UTF-8 encoding
	if !utf8.ValidString(text) {
		return &InputError{
			Type:    "encoding",
			Message: "input contains invalid UTF-8 characters",
			Input:   text,
		}
	}
	
	// Check for potentially problematic characters
	if err := p.checkProblematicChars(text); err != nil {
		return err
	}
	
	return nil
}

// checkProblematicChars checks for characters that might cause issues
func (p *InputProcessor) checkProblematicChars(text string) error {
	// Count control characters (excluding common ones like \n, \r, \t)
	controlCharCount := 0
	for _, r := range text {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			controlCharCount++
		}
	}
	
	// Warn if too many control characters
	if controlCharCount > 10 {
		return &InputError{
			Type:    "characters",
			Message: fmt.Sprintf("input contains %d control characters which may cause processing issues", controlCharCount),
			Input:   text,
		}
	}
	
	// Check for null bytes
	if strings.Contains(text, "\x00") {
		return &InputError{
			Type:    "characters",
			Message: "input contains null bytes which are not allowed",
			Input:   text,
		}
	}
	
	return nil
}

// CleanText performs basic text cleaning while preserving meaning
func (p *InputProcessor) CleanText(text string) string {
	// Remove null bytes
	cleaned := strings.ReplaceAll(text, "\x00", "")
	
	// Normalize line endings to Unix style
	cleaned = strings.ReplaceAll(cleaned, "\r\n", "\n")
	cleaned = strings.ReplaceAll(cleaned, "\r", "\n")
	
	// Remove excessive whitespace while preserving intentional formatting
	lines := strings.Split(cleaned, "\n")
	var cleanedLines []string
	
	for _, line := range lines {
		// Trim trailing whitespace but preserve leading whitespace for formatting
		line = strings.TrimRight(line, " \t")
		cleanedLines = append(cleanedLines, line)
	}
	
	cleaned = strings.Join(cleanedLines, "\n")
	
	// Remove excessive blank lines (more than 2 consecutive)
	for strings.Contains(cleaned, "\n\n\n\n") {
		cleaned = strings.ReplaceAll(cleaned, "\n\n\n\n", "\n\n\n")
	}
	
	// Trim leading and trailing whitespace
	cleaned = strings.TrimSpace(cleaned)
	
	return cleaned
}

// SplitByLength splits text into chunks of specified maximum length
// Attempts to split at word boundaries when possible
func (p *InputProcessor) SplitByLength(text string, maxLength int) []string {
	if len(text) <= maxLength {
		return []string{text}
	}
	
	var chunks []string
	remaining := text
	
	for len(remaining) > maxLength {
		// Find the best split point
		splitPoint := p.findSplitPoint(remaining, maxLength)
		
		chunk := remaining[:splitPoint]
		chunks = append(chunks, strings.TrimSpace(chunk))
		remaining = strings.TrimSpace(remaining[splitPoint:])
	}
	
	// Add the final chunk if there's remaining text
	if len(remaining) > 0 {
		chunks = append(chunks, remaining)
	}
	
	return chunks
}

// findSplitPoint finds the best point to split text, preferring word boundaries
func (p *InputProcessor) findSplitPoint(text string, maxLength int) int {
	if len(text) <= maxLength {
		return len(text)
	}
	
	// Try to find a good break point (sentence, then clause, then word)
	breakChars := []string{". ", "! ", "? ", "; ", ", ", " "}
	
	for _, breakChar := range breakChars {
		// Look for break character within the last 20% of the allowed length
		searchStart := maxLength - (maxLength / 5)
		if searchStart < 0 {
			searchStart = 0
		}
		
		substr := text[searchStart:maxLength]
		if idx := strings.LastIndex(substr, breakChar); idx != -1 {
			return searchStart + idx + len(breakChar)
		}
	}
	
	// If no good break point found, split at maxLength
	return maxLength
}

// GetTextStats returns statistics about the input text
func (p *InputProcessor) GetTextStats(text string) TextStats {
	lines := strings.Split(text, "\n")
	words := strings.Fields(text)
	runes := []rune(text)
	
	return TextStats{
		Characters:    len(text),
		CharactersUTF: len(runes),
		Words:         len(words),
		Lines:         len(lines),
		Bytes:         len([]byte(text)),
		IsValidUTF8:   utf8.ValidString(text),
	}
}

// TextStats contains statistics about text
type TextStats struct {
	Characters    int  `json:"characters"`     // Number of bytes
	CharactersUTF int  `json:"characters_utf"` // Number of UTF-8 characters/runes
	Words         int  `json:"words"`
	Lines         int  `json:"lines"`
	Bytes         int  `json:"bytes"`
	IsValidUTF8   bool `json:"is_valid_utf8"`
}

// String returns a human-readable representation of text statistics
func (s TextStats) String() string {
	return fmt.Sprintf("Characters: %d (UTF-8: %d), Words: %d, Lines: %d, Bytes: %d, Valid UTF-8: %t",
		s.Characters, s.CharactersUTF, s.Words, s.Lines, s.Bytes, s.IsValidUTF8)
}