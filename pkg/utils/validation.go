package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// SSMLValidator handles SSML validation and security checks
type SSMLValidator struct {
	// Allow basic SSML tags by default
	allowedTags map[string]bool
	// Patterns for detecting potentially malicious content
	dangerousPatterns []*regexp.Regexp
}

// ValidationError represents validation-related errors
type ValidationError struct {
	Type    string
	Message string
	Input   string
	Pos     int // Position where error occurred (-1 if not applicable)
}

func (e *ValidationError) Error() string {
	if e.Pos > 0 {
		return fmt.Sprintf("validation %s at position %d: %s", e.Type, e.Pos, e.Message)
	}
	return fmt.Sprintf("validation %s: %s", e.Type, e.Message)
}

// NewSSMLValidator creates a new SSML validator with default settings
func NewSSMLValidator() *SSMLValidator {
	validator := &SSMLValidator{
		allowedTags:       make(map[string]bool),
		dangerousPatterns: make([]*regexp.Regexp, 0),
	}

	// Initialize with safe SSML tags
	validator.initializeAllowedTags()
	validator.initializeDangerousPatterns()

	return validator
}

// initializeAllowedTags sets up the list of allowed SSML tags
func (v *SSMLValidator) initializeAllowedTags() {
	// Google Cloud TTS supported SSML tags (safe subset)
	safeTags := []string{
		"speak",    // Root element
		"p",        // Paragraph
		"s",        // Sentence
		"break",    // Pause
		"emphasis", // Emphasis
		"prosody",  // Prosody (rate, pitch, volume)
		"say-as",   // Say-as (interpret-as)
		"sub",      // Substitute
		"mark",     // Mark (for timing)
		"audio",    // Audio (with restrictions)
		"desc",     // Description
	}

	for _, tag := range safeTags {
		v.allowedTags[tag] = true
	}
}

// initializeDangerousPatterns sets up patterns for detecting dangerous content
func (v *SSMLValidator) initializeDangerousPatterns() {
	// Patterns that could indicate injection attempts or malicious content
	dangerousRegexps := []string{
		// Script injection attempts
		`(?i)<script[^>]*>`,
		`(?i)javascript:`,
		`(?i)vbscript:`,
		`(?i)onload\s*=`,
		`(?i)onerror\s*=`,
		`(?i)onclick\s*=`,

		// File system access attempts
		`(?i)file://`,
		`(?i)\.\.[\\/]`,
		`(?i)[\\\/]etc[\\/]`,
		`(?i)[\\\/]proc[\\/]`,

		// Network access attempts
		`(?i)http://`,
		`(?i)https://`,
		`(?i)ftp://`,

		// System command injection
		`(?i)system\s*\(`,
		`(?i)exec\s*\(`,
		`(?i)eval\s*\(`,

		// XML External Entity (XXE) attempts
		`(?i)<!ENTITY`,
		`(?i)<!DOCTYPE.*ENTITY`,
		`(?i)&[a-zA-Z][a-zA-Z0-9]*;.*SYSTEM`,

		// Excessive nesting (potential DoS)
		`(<[^>]+>){50,}`, // More than 50 nested tags
	}

	for _, pattern := range dangerousRegexps {
		if regex, err := regexp.Compile(pattern); err == nil {
			v.dangerousPatterns = append(v.dangerousPatterns, regex)
		}
	}
}

// IsSSML determines if the input text contains SSML markup
func (v *SSMLValidator) IsSSML(text string) bool {
	// Simple check for SSML structure - any angle bracket indicates potential SSML
	return strings.Contains(text, "<") || strings.Contains(text, ">")
}

// ValidateSSML performs comprehensive SSML validation
func (v *SSMLValidator) ValidateSSML(text string) error {
	if !v.IsSSML(text) {
		// Not SSML, no validation needed
		return nil
	}

	// Check for dangerous patterns first
	if err := v.checkDangerousPatterns(text); err != nil {
		return err
	}

	// Validate SSML structure
	if err := v.validateSSMLStructure(text); err != nil {
		return err
	}

	// Validate allowed tags
	if err := v.validateAllowedTags(text); err != nil {
		return err
	}

	// Validate tag nesting and attributes
	if err := v.validateTagAttributes(text); err != nil {
		return err
	}

	return nil
}

// checkDangerousPatterns checks for potentially malicious patterns
func (v *SSMLValidator) checkDangerousPatterns(text string) error {
	for _, pattern := range v.dangerousPatterns {
		if match := pattern.FindStringIndex(text); match != nil {
			return &ValidationError{
				Type:    "security",
				Message: "input contains potentially dangerous content",
				Input:   text[match[0]:match[1]],
				Pos:     match[0],
			}
		}
	}
	return nil
}

// validateSSMLStructure validates basic SSML XML structure
func (v *SSMLValidator) validateSSMLStructure(text string) error {
	// Basic XML well-formedness check
	tagStack := make([]string, 0)
	tagRegex := regexp.MustCompile(`<(/?)([a-zA-Z][a-zA-Z0-9-]*)[^/>]*(/?)>`)

	matches := tagRegex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		isClosing := match[1] == "/"
		tagName := match[2]
		isSelfClosing := match[3] == "/"

		if isSelfClosing {
			// Self-closing tag, no stack manipulation needed
			continue
		}

		if isClosing {
			// Closing tag
			if len(tagStack) == 0 {
				return &ValidationError{
					Type:    "structure",
					Message: fmt.Sprintf("unexpected closing tag: %s", tagName),
					Input:   match[0],
				}
			}

			// Check if it matches the most recent opening tag
			if tagStack[len(tagStack)-1] != tagName {
				return &ValidationError{
					Type:    "structure",
					Message: fmt.Sprintf("mismatched tag: expected %s, got %s", tagStack[len(tagStack)-1], tagName),
					Input:   match[0],
				}
			}

			// Pop from stack
			tagStack = tagStack[:len(tagStack)-1]
		} else {
			// Opening tag
			tagStack = append(tagStack, tagName)
		}
	}

	// Check for unclosed tags
	if len(tagStack) > 0 {
		return &ValidationError{
			Type:    "structure",
			Message: fmt.Sprintf("unclosed tag: %s", tagStack[len(tagStack)-1]),
		}
	}

	return nil
}

// validateAllowedTags checks if all tags are in the allowed list
func (v *SSMLValidator) validateAllowedTags(text string) error {
	tagRegex := regexp.MustCompile(`<(?:/?([a-zA-Z][a-zA-Z0-9-]*)[^>]*)/?>`)
	matches := tagRegex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		tagName := strings.ToLower(match[1])
		if !v.allowedTags[tagName] {
			return &ValidationError{
				Type:    "tag",
				Message: fmt.Sprintf("tag not allowed: %s", tagName),
				Input:   match[0],
			}
		}
	}

	return nil
}

// validateTagAttributes validates attributes for specific tags
func (v *SSMLValidator) validateTagAttributes(text string) error {
	// Validate prosody tag attributes
	if err := v.validateProsodyAttributes(text); err != nil {
		return err
	}

	// Validate say-as tag attributes
	if err := v.validateSayAsAttributes(text); err != nil {
		return err
	}

	// Validate break tag attributes
	if err := v.validateBreakAttributes(text); err != nil {
		return err
	}

	// Validate audio tag attributes (with security restrictions)
	if err := v.validateAudioAttributes(text); err != nil {
		return err
	}

	return nil
}

// validateProsodyAttributes validates prosody tag attributes
func (v *SSMLValidator) validateProsodyAttributes(text string) error {
	prosodyRegex := regexp.MustCompile(`<prosody\s+([^>]+)>`)
	matches := prosodyRegex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if err := v.validateSingleProsodyTag(match); err != nil {
			return err
		}
	}

	return nil
}

func (v *SSMLValidator) validateSingleProsodyTag(match []string) error {
	attrs := match[1]
	tag := match[0]

	if err := v.validateProsodyRate(attrs, tag); err != nil {
		return err
	}
	if err := v.validateProsodyPitch(attrs, tag); err != nil {
		return err
	}
	return v.validateProsodyVolume(attrs, tag)
}

func (v *SSMLValidator) validateProsodyRate(attrs, tag string) error {
	if !strings.Contains(attrs, "rate=") {
		return nil
	}
	rateRegex := regexp.MustCompile(`rate=["']?([^"'\s>]+)["']?`)
	rateMatch := rateRegex.FindStringSubmatch(attrs)
	if rateMatch != nil {
		rate := rateMatch[1]
		if !v.isValidProsodyRate(rate) {
			return &ValidationError{
				Type:    "attribute",
				Message: fmt.Sprintf("invalid prosody rate: %s", rate),
				Input:   tag,
			}
		}
	}
	return nil
}

func (v *SSMLValidator) validateProsodyPitch(attrs, tag string) error {
	if !strings.Contains(attrs, "pitch=") {
		return nil
	}
	pitchRegex := regexp.MustCompile(`pitch=["']?([^"'\s>]+)["']?`)
	pitchMatch := pitchRegex.FindStringSubmatch(attrs)
	if pitchMatch != nil {
		pitch := pitchMatch[1]
		if !v.isValidProsodyPitch(pitch) {
			return &ValidationError{
				Type:    "attribute",
				Message: fmt.Sprintf("invalid prosody pitch: %s", pitch),
				Input:   tag,
			}
		}
	}
	return nil
}

func (v *SSMLValidator) validateProsodyVolume(attrs, tag string) error {
	if !strings.Contains(attrs, "volume=") {
		return nil
	}
	volumeRegex := regexp.MustCompile(`volume=["']?([^"'\s>]+)["']?`)
	volumeMatch := volumeRegex.FindStringSubmatch(attrs)
	if volumeMatch != nil {
		volume := volumeMatch[1]
		if !v.isValidProsodyVolume(volume) {
			return &ValidationError{
				Type:    "attribute",
				Message: fmt.Sprintf("invalid prosody volume: %s", volume),
				Input:   tag,
			}
		}
	}
	return nil
}

// validateSayAsAttributes validates say-as tag attributes
func (v *SSMLValidator) validateSayAsAttributes(text string) error {
	// Match say-as tags with or without attributes
	sayAsRegex := regexp.MustCompile(`<say-as(\s+[^>]+)?>`)
	matches := sayAsRegex.FindAllStringSubmatch(text, -1)

	validInterpretAs := map[string]bool{
		"characters": true,
		"spell-out":  true,
		"cardinal":   true,
		"number":     true,
		"ordinal":    true,
		"digits":     true,
		"fraction":   true,
		"unit":       true,
		"date":       true,
		"time":       true,
		"telephone":  true,
		"address":    true,
		"expletive":  true,
		"bleep":      true,
	}

	for _, match := range matches {
		attrs := match[1] // This will be empty string if no attributes

		// interpret-as is required for say-as
		interpretRegex := regexp.MustCompile(`interpret-as=["']?([^"'\s>]+)["']?`)
		interpretMatch := interpretRegex.FindStringSubmatch(attrs)

		if interpretMatch == nil {
			return &ValidationError{
				Type:    "attribute",
				Message: "say-as tag missing required interpret-as attribute",
				Input:   match[0],
			}
		}

		interpretAs := interpretMatch[1]
		if !validInterpretAs[interpretAs] {
			return &ValidationError{
				Type:    "attribute",
				Message: fmt.Sprintf("invalid interpret-as value: %s", interpretAs),
				Input:   match[0],
			}
		}
	}

	return nil
}

// validateBreakAttributes validates break tag attributes
func (v *SSMLValidator) validateBreakAttributes(text string) error {
	breakRegex := regexp.MustCompile(`<break\s+([^>]+)/?>`)
	matches := breakRegex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if err := v.validateSingleBreakTag(match); err != nil {
			return err
		}
	}

	return nil
}

func (v *SSMLValidator) validateSingleBreakTag(match []string) error {
	attrs := match[1]
	tag := match[0]

	if err := v.validateBreakTime(attrs, tag); err != nil {
		return err
	}
	return v.validateBreakStrength(attrs, tag)
}

func (v *SSMLValidator) validateBreakTime(attrs, tag string) error {
	if !strings.Contains(attrs, "time=") {
		return nil
	}
	timeRegex := regexp.MustCompile(`time=["']?([^"'\s>]+)["']?`)
	timeMatch := timeRegex.FindStringSubmatch(attrs)
	if timeMatch != nil {
		timeValue := timeMatch[1]
		if !v.isValidBreakTime(timeValue) {
			return &ValidationError{
				Type:    "attribute",
				Message: fmt.Sprintf("invalid break time: %s", timeValue),
				Input:   tag,
			}
		}
	}
	return nil
}

func (v *SSMLValidator) validateBreakStrength(attrs, tag string) error {
	if !strings.Contains(attrs, "strength=") {
		return nil
	}
	strengthRegex := regexp.MustCompile(`strength=["']?([^"'\s>]+)["']?`)
	strengthMatch := strengthRegex.FindStringSubmatch(attrs)
	if strengthMatch != nil {
		strength := strengthMatch[1]
		if !v.isValidBreakStrength(strength) {
			return &ValidationError{
				Type:    "attribute",
				Message: fmt.Sprintf("invalid break strength: %s", strength),
				Input:   tag,
			}
		}
	}
	return nil
}

func (v *SSMLValidator) isValidBreakStrength(strength string) bool {
	validStrengths := []string{"none", "x-weak", "weak", "medium", "strong", "x-strong"}
	for _, valid := range validStrengths {
		if strength == valid {
			return true
		}
	}
	return false
}

// validateAudioAttributes validates audio tag attributes with security restrictions
func (v *SSMLValidator) validateAudioAttributes(text string) error {
	// For security, we'll be very restrictive with audio tags
	audioRegex := regexp.MustCompile(`<audio\s+([^>]+)>`)
	matches := audioRegex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		// For security, reject all audio tags
		return &ValidationError{
			Type:    "security",
			Message: "audio tags are not allowed for security reasons",
			Input:   match[0],
		}
	}

	return nil
}

// Helper validation functions
func (v *SSMLValidator) isValidProsodyRate(rate string) bool {
	// Validate prosody rate values
	validRates := []string{"x-slow", "slow", "medium", "fast", "x-fast"}
	for _, valid := range validRates {
		if rate == valid {
			return true
		}
	}

	// Check for percentage values (e.g., "50%", "200%")
	percentRegex := regexp.MustCompile(`^\d+%$`)
	if percentRegex.MatchString(rate) {
		return true
	}

	// Check for relative values (e.g., "+10%", "-20%")
	relativeRegex := regexp.MustCompile(`^[+-]\d+%$`)
	return relativeRegex.MatchString(rate)
}

func (v *SSMLValidator) isValidProsodyPitch(pitch string) bool {
	// Validate prosody pitch values
	validPitches := []string{"x-low", "low", "medium", "high", "x-high"}
	for _, valid := range validPitches {
		if pitch == valid {
			return true
		}
	}

	// Check for Hz values (e.g., "200Hz")
	hzRegex := regexp.MustCompile(`^\d+Hz$`)
	if hzRegex.MatchString(pitch) {
		return true
	}

	// Check for relative values
	relativeRegex := regexp.MustCompile(`^[+-]\d+%$`)
	return relativeRegex.MatchString(pitch)
}

func (v *SSMLValidator) isValidProsodyVolume(volume string) bool {
	// Validate prosody volume values
	validVolumes := []string{"silent", "x-soft", "soft", "medium", "loud", "x-loud"}
	for _, valid := range validVolumes {
		if volume == valid {
			return true
		}
	}

	// Check for dB values (e.g., "+6dB", "-3dB")
	dbRegex := regexp.MustCompile(`^[+-]?\d+dB$`)
	return dbRegex.MatchString(volume)
}

func (v *SSMLValidator) isValidBreakTime(timeValue string) bool {
	// Validate break time values
	// Format: number followed by 's' (seconds) or 'ms' (milliseconds)
	timeRegex := regexp.MustCompile(`^\d+(?:\.\d+)?(?:s|ms)$`)
	if !timeRegex.MatchString(timeValue) {
		return false
	}

	// Additional validation: reasonable time limits (max 10 seconds)
	if strings.HasSuffix(timeValue, "ms") {
		// Milliseconds - max 10000ms (10s)
		msRegex := regexp.MustCompile(`^(\d+(?:\.\d+)?)ms$`)
		matches := msRegex.FindStringSubmatch(timeValue)
		if len(matches) > 1 {
			// Simple check - if the number part is more than 5 digits before decimal, it's too long
			numberPart := matches[1]
			beforeDecimal := strings.Split(numberPart, ".")[0]
			if len(beforeDecimal) > 5 {
				return false
			}
		}
	} else {
		// Seconds - max 10s
		sRegex := regexp.MustCompile(`^(\d+(?:\.\d+)?)s$`)
		matches := sRegex.FindStringSubmatch(timeValue)
		if len(matches) > 1 {
			numberPart := matches[1]
			beforeDecimal := strings.Split(numberPart, ".")[0]

			// Check if it's more than 10 seconds
			if len(beforeDecimal) > 2 {
				return false // More than 99 seconds
			}
			if len(beforeDecimal) == 2 {
				// Check if it's 10 or less (first digit 1 and second digit 0, or first digit < 1)
				if beforeDecimal[0] > '1' || (beforeDecimal[0] == '1' && beforeDecimal[1] > '0') {
					return false
				}
			}
		}
	}

	return true
}

// SanitizeText removes potentially dangerous content while preserving safe SSML
func (v *SSMLValidator) SanitizeText(text string) string {
	if !v.IsSSML(text) {
		// Not SSML, just clean up basic issues
		return strings.TrimSpace(text)
	}

	// Remove dangerous patterns
	sanitized := text
	for _, pattern := range v.dangerousPatterns {
		sanitized = pattern.ReplaceAllString(sanitized, "")
	}

	// Remove disallowed tags and their content
	// First handle script tags specifically (remove content too)
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	sanitized = scriptRegex.ReplaceAllString(sanitized, "")

	// Then remove other disallowed tags
	tagRegex := regexp.MustCompile(`<(/?)([a-zA-Z][a-zA-Z0-9-]*)[^>]*(/?)>`)
	sanitized = tagRegex.ReplaceAllStringFunc(sanitized, func(match string) string {
		submatch := tagRegex.FindStringSubmatch(match)
		if len(submatch) >= 3 {
			tagName := strings.ToLower(submatch[2])
			if v.allowedTags[tagName] {
				return match // Keep allowed tags
			}
		}
		return "" // Remove disallowed tags
	})

	return strings.TrimSpace(sanitized)
}
