package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibegear/oursky/pkg/templates"
	"gopkg.in/yaml.v3"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected *Config
		wantErr  bool
	}{
		{
			name: "valid config",
			yaml: `name: test-project
provider: docker
services:
  web:
    command: "npm run dev"
  api:
    command: "go run main.go"
agents:
  - name: cursor
    enabled: true
mcp:
  enabled: true
  port: 3001`,
			expected: &Config{
				Name:     "test-project",
				Provider: "docker",
				Services: map[string]Service{
					"web": {Command: "npm run dev"},
					"api": {Command: "go run main.go"},
				},
				Agents: []Agent{
					{Name: "cursor", Enabled: true},
				},
				MCP: struct {
					Enabled bool   `yaml:"enabled,omitempty"`
					Port    int    `yaml:"port,omitempty"`
					Host    string `yaml:"host,omitempty"`
				}{
					Enabled: true,
					Port:    3001,
				},
			},
			wantErr: false,
		},
		{
			name:    "file not found",
			yaml:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file for valid configs
			var tempFile string
			if tt.yaml != "" {
				temp, err := os.CreateTemp("", "config-*.yaml")
				require.NoError(t, err)
				defer os.Remove(temp.Name())

				_, err = temp.WriteString(tt.yaml)
				require.NoError(t, err)
				temp.Close()

				tempFile = temp.Name()
			} else {
				tempFile = "nonexistent.yaml"
			}

			config, err := LoadConfig(tempFile)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.expected.Name, config.Name)
				assert.Equal(t, tt.expected.Provider, config.Provider)
				assert.Equal(t, len(tt.expected.Services), len(config.Services))
				assert.Equal(t, len(tt.expected.Agents), len(config.Agents))
			}
		})
	}
}

func TestConfig_GetMergedTemplates(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "config-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create mock template structure
	templatesDir := filepath.Join(tempDir, ".vendatta", "templates")
	require.NoError(t, os.MkdirAll(templatesDir, 0755))

	// Create a mock template file
	templateContent := `skills:
  - name: "test-skill"
    description: "A test skill"
rules:
  - name: "test-rule"
    content: "Test rule content"`

	templateFile := filepath.Join(templatesDir, "test.yaml")
	require.NoError(t, os.WriteFile(templateFile, []byte(templateContent), 0644))

	config := &Config{}

	_, _ = config.GetMergedTemplates(tempDir)
}

func TestConfig_GenerateAgentConfigs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "agent-config-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create worktree directory
	worktreeDir := filepath.Join(tempDir, "worktree")
	require.NoError(t, os.MkdirAll(worktreeDir, 0755))

	config := &Config{
		Name: "test-project",
		MCP: struct {
			Enabled bool   `yaml:"enabled,omitempty"`
			Port    int    `yaml:"port,omitempty"`
			Host    string `yaml:"host,omitempty"`
		}{
			Enabled: true,
			Port:    3001,
			Host:    "localhost",
		},
		Agents: []Agent{
			{Name: "cursor", Enabled: true},
		},
	}

	// Mock merged templates (empty for this test)
	mergedTemplates := &templates.TemplateData{}

	err = config.GenerateAgentConfigs(worktreeDir, mergedTemplates)
	_ = err
}

func TestUpdateGitignore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gitignore-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp dir so .gitignore operations work relative to it
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Test with no existing gitignore
	err = updateGitignore([]string{".cursor/"})
	assert.NoError(t, err)

	content, err := os.ReadFile(".gitignore")
	require.NoError(t, err)
	assert.Contains(t, string(content), ".cursor/")

	// Test with existing gitignore
	require.NoError(t, os.WriteFile(".gitignore", []byte("node_modules/\n"), 0644))
	err = updateGitignore([]string{".opencode/"})
	assert.NoError(t, err)

	content, err = os.ReadFile(".gitignore")
	require.NoError(t, err)
	assert.Contains(t, string(content), "node_modules/")
	assert.Contains(t, string(content), ".opencode/")
}

func TestConfigYAMLMarshalling(t *testing.T) {
	config := &Config{
		Name:     "test-project",
		Provider: "docker",
		Services: map[string]Service{
			"web": {
				Command: "npm run dev",
				Port:    3000,
			},
			"api": {
				Command:   "go run main.go",
				DependsOn: []string{"db"},
				Env: map[string]string{
					"DEBUG": "true",
				},
			},
		},
		Agents: []Agent{
			{Name: "cursor", Enabled: true},
			{Name: "opencode", Enabled: false},
		},
		MCP: struct {
			Enabled bool   `yaml:"enabled,omitempty"`
			Port    int    `yaml:"port,omitempty"`
			Host    string `yaml:"host,omitempty"`
		}{
			Enabled: true,
			Port:    3001,
			Host:    "localhost",
		},
	}

	// Test marshalling
	data, err := yaml.Marshal(config)
	require.NoError(t, err)

	// Verify key fields are present
	yamlStr := string(data)
	assert.Contains(t, yamlStr, "test-project")
	assert.Contains(t, yamlStr, "npm run dev")
	assert.Contains(t, yamlStr, "cursor")
	assert.Contains(t, yamlStr, "3001")

	// Test unmarshalling
	var unmarshalled Config
	err = yaml.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)

	assert.Equal(t, config.Name, unmarshalled.Name)
	assert.Equal(t, config.Provider, unmarshalled.Provider)
	assert.Equal(t, len(config.Services), len(unmarshalled.Services))
	assert.Equal(t, len(config.Agents), len(unmarshalled.Agents))
	assert.Equal(t, config.MCP.Port, unmarshalled.MCP.Port)
}
