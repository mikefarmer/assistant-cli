package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSSMLValidator(t *testing.T) {
	validator := NewSSMLValidator()
	
	assert.NotNil(t, validator)
	assert.NotEmpty(t, validator.allowedTags)
	assert.NotEmpty(t, validator.dangerousPatterns)
	
	// Check some expected allowed tags
	assert.True(t, validator.allowedTags["speak"])
	assert.True(t, validator.allowedTags["p"])
	assert.True(t, validator.allowedTags["break"])
	assert.True(t, validator.allowedTags["emphasis"])
}

func TestSSMLValidator_IsSSML(t *testing.T) {
	validator := NewSSMLValidator()
	
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"plain text", "Hello World", false},
		{"SSML with speak tag", "<speak>Hello World</speak>", true},
		{"SSML with break", "Hello <break time='1s'/> World", true},
		{"HTML-like but not SSML", "<div>Hello</div>", true}, // Still contains < and >
		{"no angle brackets", "Hello World without tags", false},
		{"only opening bracket", "Hello < World", true},
		{"only closing bracket", "Hello > World", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.IsSSML(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSSMLValidator_ValidateSSML_PlainText(t *testing.T) {
	validator := NewSSMLValidator()
	
	// Plain text should pass validation (no SSML to validate)
	err := validator.ValidateSSML("Hello World")
	assert.NoError(t, err)
}

func TestSSMLValidator_ValidateSSML_ValidSSML(t *testing.T) {
	validator := NewSSMLValidator()
	
	validSSMLCases := []string{
		"<speak>Hello World</speak>",
		"<speak><p>Hello</p><p>World</p></speak>",
		"<speak>Hello <break time='1s'/> World</speak>",
		"<speak><emphasis level='strong'>Important</emphasis></speak>",
		"<speak><prosody rate='slow'>Slow speech</prosody></speak>",
		"<speak><say-as interpret-as='digits'>123</say-as></speak>",
		"<speak><sub alias='World Wide Web'>WWW</sub></speak>",
	}
	
	for _, ssml := range validSSMLCases {
		t.Run(ssml, func(t *testing.T) {
			err := validator.ValidateSSML(ssml)
			assert.NoError(t, err)
		})
	}
}

func TestSSMLValidator_ValidateSSML_DangerousPatterns(t *testing.T) {
	validator := NewSSMLValidator()
	
	dangerousCases := []struct {
		name  string
		input string
	}{
		{"script tag", "<speak><script>alert('xss')</script>Hello</speak>"},
		{"javascript protocol", "<speak onclick='javascript:alert()'>Hello</speak>"},
		{"file protocol", "<audio src='file:///etc/passwd'>Hello</audio>"},
		{"http URL", "<audio src='http://evil.com/malware'>Hello</audio>"},
		{"XXE attempt", "<!ENTITY xxe SYSTEM 'file:///etc/passwd'><speak>&xxe;</speak>"},
		{"system command", "<speak>system('rm -rf /')</speak>"},
	}
	
	for _, tc := range dangerousCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateSSML(tc.input)
			require.Error(t, err)
			
			var validationErr *ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "security", validationErr.Type)
		})
	}
}

func TestSSMLValidator_ValidateSSML_StructureErrors(t *testing.T) {
	validator := NewSSMLValidator()
	
	structureErrorCases := []struct {
		name  string
		input string
	}{
		{"unclosed tag", "<speak>Hello World"},
		{"mismatched tags", "<speak><p>Hello</emphasis></speak>"},
		{"unexpected closing tag", "</speak>Hello World"},
		{"wrong nesting", "<speak><p>Hello<break></p>World</speak>"},
	}
	
	for _, tc := range structureErrorCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateSSML(tc.input)
			require.Error(t, err)
			
			var validationErr *ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "structure", validationErr.Type)
		})
	}
}

func TestSSMLValidator_ValidateSSML_DisallowedTags(t *testing.T) {
	validator := NewSSMLValidator()
	
	disallowedCases := []struct {
		name  string
		input string
	}{
		{"script tag", "<speak><script>alert('xss')</script></speak>"},
		{"div tag", "<speak><div>Hello</div></speak>"},
		{"iframe tag", "<speak><iframe src='evil.com'></iframe></speak>"},
		{"img tag", "<speak><img src='malware.jpg'/></speak>"},
	}
	
	for _, tc := range disallowedCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateSSML(tc.input)
			require.Error(t, err)
			
			var validationErr *ValidationError
			assert.ErrorAs(t, err, &validationErr)
			// Could be either "security" or "tag" error
			assert.True(t, validationErr.Type == "security" || validationErr.Type == "tag")
		})
	}
}

func TestSSMLValidator_ValidateSSML_AudioTag(t *testing.T) {
	validator := NewSSMLValidator()
	
	// Audio tags should be rejected for security
	err := validator.ValidateSSML("<speak><audio src='test.mp3'>Hello</audio></speak>")
	require.Error(t, err)
	
	var validationErr *ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "security", validationErr.Type)
	assert.Contains(t, err.Error(), "audio tags are not allowed")
}

func TestSSMLValidator_validateProsodyAttributes(t *testing.T) {
	validator := NewSSMLValidator()
	
	validCases := []string{
		"<prosody rate='slow'>Hello</prosody>",
		"<prosody rate='50%'>Hello</prosody>",
		"<prosody rate='+10%'>Hello</prosody>",
		"<prosody pitch='high'>Hello</prosody>",
		"<prosody pitch='200Hz'>Hello</prosody>",
		"<prosody volume='loud'>Hello</prosody>",
		"<prosody volume='+6dB'>Hello</prosody>",
	}
	
	for _, ssml := range validCases {
		t.Run(ssml, func(t *testing.T) {
			err := validator.validateProsodyAttributes(ssml)
			assert.NoError(t, err)
		})
	}
	
	invalidCases := []struct {
		name  string
		input string
	}{
		{"invalid rate", "<prosody rate='invalid'>Hello</prosody>"},
		{"invalid pitch", "<prosody pitch='invalid'>Hello</prosody>"},
		{"invalid volume", "<prosody volume='invalid'>Hello</prosody>"},
	}
	
	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.validateProsodyAttributes(tc.input)
			require.Error(t, err)
			
			var validationErr *ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "attribute", validationErr.Type)
		})
	}
}

func TestSSMLValidator_validateSayAsAttributes(t *testing.T) {
	validator := NewSSMLValidator()
	
	validCases := []string{
		"<say-as interpret-as='digits'>123</say-as>",
		"<say-as interpret-as='cardinal'>123</say-as>",
		"<say-as interpret-as='date'>2024-01-01</say-as>",
		"<say-as interpret-as='time'>12:30</say-as>",
	}
	
	for _, ssml := range validCases {
		t.Run(ssml, func(t *testing.T) {
			err := validator.validateSayAsAttributes(ssml)
			assert.NoError(t, err)
		})
	}
	
	invalidCases := []struct {
		name  string
		input string
	}{
		{"missing interpret-as", "<say-as>123</say-as>"},
		{"invalid interpret-as", "<say-as interpret-as='invalid'>123</say-as>"},
	}
	
	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.validateSayAsAttributes(tc.input)
			require.Error(t, err)
			
			var validationErr *ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "attribute", validationErr.Type)
		})
	}
}

func TestSSMLValidator_validateBreakAttributes(t *testing.T) {
	validator := NewSSMLValidator()
	
	validCases := []string{
		"<break time='1s'/>",
		"<break time='500ms'/>",
		"<break time='2.5s'/>",
		"<break strength='weak'/>",
		"<break strength='x-strong'/>",
		"<break/>", // no attributes is valid
	}
	
	for _, ssml := range validCases {
		t.Run(ssml, func(t *testing.T) {
			err := validator.validateBreakAttributes(ssml)
			assert.NoError(t, err)
		})
	}
	
	invalidCases := []struct {
		name  string
		input string
	}{
		{"invalid time format", "<break time='invalid'/>"},
		{"time too long", "<break time='100s'/>"},
		{"invalid strength", "<break strength='invalid'/>"},
	}
	
	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.validateBreakAttributes(tc.input)
			require.Error(t, err)
			
			var validationErr *ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "attribute", validationErr.Type)
		})
	}
}

func TestSSMLValidator_isValidProsodyRate(t *testing.T) {
	validator := NewSSMLValidator()
	
	validRates := []string{
		"x-slow", "slow", "medium", "fast", "x-fast",
		"50%", "200%", "100%",
		"+10%", "-20%", "+50%",
	}
	
	for _, rate := range validRates {
		t.Run(rate, func(t *testing.T) {
			assert.True(t, validator.isValidProsodyRate(rate))
		})
	}
	
	invalidRates := []string{
		"invalid", "very-slow", "50", "+50", "200%%", "abc%",
	}
	
	for _, rate := range invalidRates {
		t.Run(rate, func(t *testing.T) {
			assert.False(t, validator.isValidProsodyRate(rate))
		})
	}
}

func TestSSMLValidator_isValidProsodyPitch(t *testing.T) {
	validator := NewSSMLValidator()
	
	validPitches := []string{
		"x-low", "low", "medium", "high", "x-high",
		"200Hz", "440Hz", "100Hz",
		"+10%", "-20%", "+50%",
	}
	
	for _, pitch := range validPitches {
		t.Run(pitch, func(t *testing.T) {
			assert.True(t, validator.isValidProsodyPitch(pitch))
		})
	}
	
	invalidPitches := []string{
		"invalid", "very-high", "200", "Hz", "200HZ", "+Hz",
	}
	
	for _, pitch := range invalidPitches {
		t.Run(pitch, func(t *testing.T) {
			assert.False(t, validator.isValidProsodyPitch(pitch))
		})
	}
}

func TestSSMLValidator_isValidBreakTime(t *testing.T) {
	validator := NewSSMLValidator()
	
	validTimes := []string{
		"1s", "2.5s", "0.5s", "10s",
		"100ms", "1000ms", "2500ms",
	}
	
	for _, time := range validTimes {
		t.Run(time, func(t *testing.T) {
			assert.True(t, validator.isValidBreakTime(time))
		})
	}
	
	invalidTimes := []string{
		"invalid", "1", "s", "1sec", "100milliseconds",
		"100000ms", "100s", // too long
	}
	
	for _, time := range invalidTimes {
		t.Run(time, func(t *testing.T) {
			assert.False(t, validator.isValidBreakTime(time))
		})
	}
}

func TestSSMLValidator_SanitizeText(t *testing.T) {
	validator := NewSSMLValidator()
	
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain text", "Hello World", "Hello World"},
		{"valid SSML", "<speak>Hello</speak>", "<speak>Hello</speak>"},
		{"remove script tag", "<speak><script>evil</script>Hello</speak>", "<speak>evilHello</speak>"},
		{"remove disallowed tag", "<speak><div>Hello</div></speak>", "<speak>Hello</speak>"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.SanitizeText(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	testCases := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			"error without position",
			&ValidationError{Type: "test", Message: "test message"},
			"validation test: test message",
		},
		{
			"error with position",
			&ValidationError{Type: "test", Message: "test message", Pos: 10},
			"validation test at position 10: test message",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.err.Error()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Integration test with complex SSML
func TestSSMLValidator_ComplexSSML(t *testing.T) {
	validator := NewSSMLValidator()
	
	complexSSML := `<speak>
		<p>Welcome to our <emphasis level='strong'>text-to-speech</emphasis> service.</p>
		<break time='1s'/>
		<p><prosody rate='slow' pitch='low'>This is spoken slowly and in a low pitch.</prosody></p>
		<p>The number is <say-as interpret-as='digits'>12345</say-as>.</p>
		<p><sub alias='World Wide Web'>WWW</sub> stands for World Wide Web.</p>
	</speak>`
	
	err := validator.ValidateSSML(complexSSML)
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkSSMLValidator_ValidateSSML_PlainText(b *testing.B) {
	validator := NewSSMLValidator()
	text := "Hello World! This is a plain text message without any SSML markup."
	
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateSSML(text)
	}
}

func BenchmarkSSMLValidator_ValidateSSML_SimpleSSML(b *testing.B) {
	validator := NewSSMLValidator()
	ssml := "<speak>Hello <break time='1s'/> World!</speak>"
	
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateSSML(ssml)
	}
}

func BenchmarkSSMLValidator_ValidateSSML_ComplexSSML(b *testing.B) {
	validator := NewSSMLValidator()
	ssml := `<speak><p><prosody rate='slow'>Hello</prosody> <emphasis>World</emphasis></p>` +
		`<break time='2s'/><say-as interpret-as='digits'>123</say-as></speak>`
	
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateSSML(ssml)
	}
}