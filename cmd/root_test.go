package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		wantOutput  string
	}{
		{
			name:        "help flag",
			args:        []string{"--help"},
			wantErr:     false,
			wantOutput:  "A personal assistant CLI tool with various capabilities",
		},
		{
			name:        "version flag",
			args:        []string{"--version"},
			wantErr:     false,
			wantOutput:  "assistant-cli version",
		},
		{
			name:        "no args shows help",
			args:        []string{},
			wantErr:     false,
			wantOutput:  "A personal assistant CLI tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			rootCmd := NewRootCmd()
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			output := buf.String()
			assert.Contains(t, output, tt.wantOutput)
		})
	}
}

func TestInitConfig(t *testing.T) {
	// Test that config initialization doesn't panic
	assert.NotPanics(t, func() {
		initConfig()
	})
}

func TestRootCommandStructure(t *testing.T) {
	rootCmd := NewRootCmd()
	
	// Test command properties
	assert.Equal(t, "assistant-cli", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "personal assistant")
	assert.NotEmpty(t, rootCmd.Long)
	
	// Test persistent flags
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "config file (default is $HOME/.assistant-cli.yaml)", configFlag.Usage)
	
	// Test that version is set
	assert.True(t, rootCmd.Version != "")
}