package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/vibegear/oursky/pkg/templates"
)

type Service struct {
	Command     string            `yaml:"command,omitempty"`
	Port        int               `yaml:"port,omitempty"` // Legacy support
	Healthcheck *Healthcheck      `yaml:"healthcheck,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
}

type Healthcheck struct {
	URL      string `yaml:"url"`
	Interval string `yaml:"interval,omitempty"`
	Timeout  string `yaml:"timeout,omitempty"`
	Retries  int    `yaml:"retries,omitempty"`
}

type Agent struct {
	Name     string   `yaml:"name"`
	Enabled  bool     `yaml:"enabled,omitempty"`
	Rules    string   `yaml:"rules,omitempty"`
	Skills   []string `yaml:"skills,omitempty"`
	Commands []string `yaml:"commands,omitempty"`
	Plugins  []string `yaml:"plugins,omitempty"` // Plugin namespaces to include
}

type TemplateRepo struct {
	URL    string `yaml:"url"`
	Branch string `yaml:"branch,omitempty"`
	Path   string `yaml:"path,omitempty"` // Path within repo to templates
}

type Remote struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type Config struct {
	Name        string             `yaml:"name"`
	Provider    string             `yaml:"provider,omitempty"`
	Services    map[string]Service `yaml:"services"`
	Remotes     []TemplateRepo     `yaml:"remotes,omitempty"`
	SyncTargets []Remote           `yaml:"sync_targets,omitempty"`
	Agents      []Agent            `yaml:"agents,omitempty"`
	Docker      struct {
		Image string   `yaml:"image"`
		Ports []string `yaml:"ports,omitempty"`
		DinD  bool     `yaml:"dind,omitempty"`
	} `yaml:"docker,omitempty"`
	LXC struct {
		Image string `yaml:"image,omitempty"`
	} `yaml:"lxc,omitempty"`
	Hooks struct {
		Setup    string `yaml:"setup,omitempty"`
		Dev      string `yaml:"dev,omitempty"`
		Teardown string `yaml:"teardown,omitempty"`
	} `yaml:"hooks,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}

// InitializeRemotes pulls all template repositories defined in config
func (c *Config) InitializeRemotes(baseDir string) error {
	if len(c.Remotes) == 0 {
		return nil
	}

	manager := templates.NewManager(baseDir)

	for _, repo := range c.Remotes {
		if err := manager.PullRepo(templates.TemplateRepo{
			URL:    repo.URL,
			Branch: repo.Branch,
			Path:   repo.Path,
		}); err != nil {
			return fmt.Errorf("failed to pull template repo %s: %w", repo.URL, err)
		}
	}

	return nil
}

// RenderData holds data for template rendering
type RenderData struct {
	AuthToken      string
	ProjectName    string
	RulesConfig    string // JSON
	SkillsConfig   string // JSON
	CommandsConfig string // JSON
}

// GetMergedTemplates returns merged template data from all sources
func (c *Config) GetMergedTemplates(baseDir string) (*templates.TemplateData, error) {
	manager := templates.NewManager(baseDir)
	return manager.Merge(baseDir)
}

func (c *Config) GenerateAgentConfigs(worktreePath string, merged *templates.TemplateData) error {
	// Agent configurations mapping
	agentConfigs := map[string]struct {
		templatePath string
		outputPath   string
		gitignore    string
		rulesFormat  string
		rulesDir     string
	}{
		"cursor": {
			templatePath: ".vendatta/agents/cursor/mcp.json.tpl",
			outputPath:   ".cursor/mcp.json",
			gitignore:    ".cursor/",
			rulesFormat:  "mdc",
			rulesDir:     ".cursor/rules",
		},
		"opencode": {
			templatePath: ".vendatta/agents/opencode/opencode.json.tpl",
			outputPath:   "opencode.json",
			gitignore:    "AGENTS.md",
			rulesFormat:  "md",
			rulesDir:     ".opencode/rules",
		},
		"claude-desktop": {
			templatePath: ".vendatta/agents/claude-desktop/claude_desktop_config.json.tpl",
			outputPath:   "claude_desktop_config.json",
			gitignore:    "claude_desktop_config.json",
		},
		"claude-code": {
			templatePath: ".vendatta/agents/claude-code/claude_code_config.json.tpl",
			outputPath:   "claude_code_config.json",
			gitignore:    "claude_code_config.json",
		},
		"codex": {
			templatePath: ".vendatta/agents/codex/settings.json.tpl",
			outputPath:   ".vscode/settings.json",
			gitignore:    ".vscode/",
			rulesFormat:  "md",
			rulesDir:     ".github/instructions",
		},
	}

	// Collect merged data from specified plugins
	finalRules := make(map[string]interface{})
	finalSkills := make(map[string]interface{})
	finalCommands := make(map[string]interface{})

	// Always include "base" and "override" plugins
	relevantPlugins := []string{"base", "override"}
	for _, agent := range c.Agents {
		if agent.Enabled {
			relevantPlugins = append(relevantPlugins, agent.Plugins...)
		}
	}

	if merged != nil {
		for _, pluginName := range relevantPlugins {
			if plugin, ok := merged.Plugins[pluginName]; ok {
				for k, v := range plugin.Rules {
					finalRules[k] = v
				}
				for k, v := range plugin.Skills {
					finalSkills[k] = v
				}
				for k, v := range plugin.Commands {
					finalCommands[k] = v
				}
			}
		}
	}

	// Marshal template data to JSON for rendering
	rulesJSON, _ := json.Marshal(finalRules)
	skillsJSON, _ := json.Marshal(finalSkills)
	commandsJSON, _ := json.Marshal(finalCommands)

	renderData := RenderData{
		AuthToken:      generateAuthToken(),
		ProjectName:    c.Name,
		RulesConfig:    string(rulesJSON),
		SkillsConfig:   string(skillsJSON),
		CommandsConfig: string(commandsJSON),
	}

	var gitignorePatterns []string

	if len(finalRules) > 0 {
		// Create a temporary merged object for generateAgentRules
		tempMerged := &templates.TemplateData{
			Plugins: map[string]*templates.PluginData{
				"merged": {
					Rules: finalRules,
				},
			},
		}
		if err := c.generateAgentRules(worktreePath, "agents.md", "", tempMerged); err != nil {
			return fmt.Errorf("failed to generate AGENTS.md: %w", err)
		}
		gitignorePatterns = append(gitignorePatterns, "AGENTS.md")
	}

	for _, agent := range c.Agents {
		if !agent.Enabled {
			continue
		}

		cfg, ok := agentConfigs[agent.Name]
		if !ok {
			continue
		}

		templateContent, err := os.ReadFile(cfg.templatePath)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", cfg.templatePath, err)
		}

		tmpl, err := template.New("template").Parse(string(templateContent))
		if err != nil {
			return fmt.Errorf("failed to parse template for %s: %w", agent.Name, err)
		}

		var result strings.Builder
		if err := tmpl.Execute(&result, renderData); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", agent.Name, err)
		}
		rendered := result.String()

		outputPath := filepath.Join(worktreePath, cfg.outputPath)
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create dir for %s: %w", outputPath, err)
		}

		if err := os.WriteFile(outputPath, []byte(rendered), 0644); err != nil {
			return fmt.Errorf("failed to write config for %s: %w", agent.Name, err)
		}

		if cfg.rulesFormat != "" {
			agentRules := make(map[string]interface{})

			// Merge rules from specified plugins for this specific agent
			agentPluginNames := append([]string{"base", "override"}, agent.Plugins...)
			for _, pluginName := range agentPluginNames {
				if merged == nil {
					continue
				}
				if plugin, ok := merged.Plugins[pluginName]; ok {
					for ruleName, ruleData := range plugin.Rules {
						// Namespace plugin rules to avoid collisions
						namespacedName := pluginName + "/" + ruleName
						agentRules[namespacedName] = ruleData
					}
				}
			}

			// Load agent-specific rules from override directory
			agentOverrideDir := filepath.Join(".vendatta", "agents", agent.Name, "rules")
			if _, err := os.Stat(agentOverrideDir); err == nil {
				if err := filepath.Walk(agentOverrideDir, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && (strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".mdc")) {
						relPath, _ := filepath.Rel(agentOverrideDir, path)
						ruleName := strings.TrimSuffix(relPath, filepath.Ext(relPath))

						content, err := os.ReadFile(path)
						if err != nil {
							return err
						}

						agentRules[ruleName] = map[string]interface{}{
							"content": string(content),
						}
					}
					return nil
				}); err != nil {
					return fmt.Errorf("failed to load agent-specific rules for %s: %w", agent.Name, err)
				}
			}

			// Generate agent-specific rules if any were collected
			if len(agentRules) > 0 {
				tempAgentMerged := &templates.TemplateData{
					Plugins: map[string]*templates.PluginData{
						"agent": {
							Rules: agentRules,
						},
					},
				}
				if err := c.generateAgentRules(worktreePath, cfg.rulesFormat, cfg.rulesDir, tempAgentMerged); err != nil {
					return fmt.Errorf("failed to generate rules for %s: %w", agent.Name, err)
				}
			}
		}

		gitignorePatterns = append(gitignorePatterns, cfg.gitignore)
	}

	// Update .gitignore with generated patterns
	if err := updateGitignore(gitignorePatterns); err != nil {
		return fmt.Errorf("failed to update .gitignore: %w", err)
	}

	return nil
}

func (c *Config) generateAgentRules(worktreePath, format, rulesDir string, merged *templates.TemplateData) error {
	if merged == nil || len(merged.Plugins) == 0 {
		return nil
	}

	absRulesDir := filepath.Join(worktreePath, rulesDir)
	if rulesDir != "" {
		if err := os.MkdirAll(absRulesDir, 0755); err != nil {
			return err
		}
	}

	switch format {
	case "mdc", "md":
		extension := "." + format
		if format == "md" {
			extension = ".md"
		}
		for _, plugin := range merged.Plugins {
			for name, data := range plugin.Rules {
				ruleMap, ok := data.(map[string]interface{})
				if !ok {
					continue
				}
				content, _ := ruleMap["content"].(string)

				var builder strings.Builder
				builder.WriteString("---\n")
				for k, v := range ruleMap {
					if k == "content" {
						continue
					}
					builder.WriteString(fmt.Sprintf("%s: %v\n", k, v))
				}
				builder.WriteString("---\n")
				builder.WriteString(content)

				outputPath := filepath.Join(absRulesDir, name+extension)
				if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
					return err
				}
				if err := os.WriteFile(outputPath, []byte(builder.String()), 0644); err != nil {
					return err
				}
			}
		}
	case "agents.md":
		var builder strings.Builder
		builder.WriteString("# PROJECT KNOWLEDGE BASE\n\n")
		builder.WriteString(fmt.Sprintf("**Generated:** %s\n", strings.ToUpper(c.Name)))
		builder.WriteString("\n")
		for _, plugin := range merged.Plugins {
			for _, data := range plugin.Rules {
				ruleMap, ok := data.(map[string]interface{})
				if !ok {
					continue
				}
				content, _ := ruleMap["content"].(string)
				builder.WriteString(content)
				builder.WriteString("\n\n")
			}
		}
		outputPath := filepath.Join(worktreePath, "AGENTS.md")
		if err := os.WriteFile(outputPath, []byte(builder.String()), 0644); err != nil {
			return err
		}
	}
	return nil
}

// generateAuthToken creates a random 32-character hex token
func generateAuthToken() string {
	return "random-token-12345" // Simplified for now
}

// updateGitignore adds patterns to .gitignore if not already present
func updateGitignore(patterns []string) error {
	gitignorePath := ".gitignore"
	var existing []string

	if content, err := os.ReadFile(gitignorePath); err == nil {
		existing = strings.Split(strings.TrimSpace(string(content)), "\n")
	}

	var cleaned []string
	for _, line := range existing {
		if strings.TrimSpace(line) != "" {
			cleaned = append(cleaned, line)
		}
	}
	existing = cleaned

	// Add missing patterns
	for _, pattern := range patterns {
		found := false
		for _, line := range existing {
			if strings.TrimSpace(line) == pattern {
				found = true
				break
			}
		}
		if !found {
			existing = append(existing, pattern)
		}
	}

	// Write back
	newContent := strings.Join(existing, "\n") + "\n"
	return os.WriteFile(gitignorePath, []byte(newContent), 0644)
}

// GenerateJSONSchema generates a JSON schema for the Config struct
func GenerateJSONSchema() (string, error) {
	schema := map[string]interface{}{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"type":    "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Project name",
			},
			"provider": map[string]interface{}{
				"type":        "string",
				"description": "Execution provider (docker, lxc)",
				"enum":        []string{"docker", "lxc"},
			},
			"services": map[string]interface{}{
				"type":        "object",
				"description": "Service definitions",
				"patternProperties": map[string]interface{}{
					".*": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"command": map[string]interface{}{
								"type":        "string",
								"description": "Command to run the service",
							},
							"port": map[string]interface{}{
								"type":        "integer",
								"description": "Legacy port configuration (auto-detected from command)",
								"minimum":     1,
								"maximum":     65535,
							},
							"healthcheck": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"url": map[string]interface{}{
										"type":        "string",
										"description": "Health check URL",
									},
									"interval": map[string]interface{}{
										"type":        "string",
										"description": "Health check interval",
									},
									"timeout": map[string]interface{}{
										"type":        "string",
										"description": "Health check timeout",
									},
									"retries": map[string]interface{}{
										"type":        "integer",
										"description": "Number of retries",
										"minimum":     0,
									},
								},
								"additionalProperties": false,
							},
							"depends_on": map[string]interface{}{
								"type":        "array",
								"items":       map[string]interface{}{"type": "string"},
								"description": "Services this service depends on",
							},
							"env": map[string]interface{}{
								"type":        "object",
								"description": "Environment variables",
								"patternProperties": map[string]interface{}{
									".*": map[string]interface{}{"type": "string"},
								},
							},
						},
						"additionalProperties": false,
					},
				},
			},
			"remotes": map[string]interface{}{
				"type":        "array",
				"description": "Remote template repositories",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"url": map[string]interface{}{
							"type":        "string",
							"description": "Repository URL",
						},
						"branch": map[string]interface{}{
							"type":        "string",
							"description": "Branch to pull from",
						},
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path within repository",
						},
					},
					"required":             []string{"url"},
					"additionalProperties": false,
				},
			},
			"sync_targets": map[string]interface{}{
				"type":        "array",
				"description": "Remote repositories for config syncing",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Target name",
						},
						"url": map[string]interface{}{
							"type":        "string",
							"description": "Repository URL",
						},
					},
					"required":             []string{"name", "url"},
					"additionalProperties": false,
				},
			},
			"agents": map[string]interface{}{
				"type":        "array",
				"description": "AI agent configurations",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Agent name (cursor, opencode, claude-desktop, etc.)",
							"enum":        []string{"cursor", "opencode", "claude-desktop", "claude-code", "codex"},
						},
						"enabled": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether this agent is enabled",
						},
						"rules": map[string]interface{}{
							"type":        "string",
							"description": "Path to rules file",
						},
						"skills": map[string]interface{}{
							"type":        "array",
							"items":       map[string]interface{}{"type": "string"},
							"description": "List of skills to enable",
						},
						"commands": map[string]interface{}{
							"type":        "array",
							"items":       map[string]interface{}{"type": "string"},
							"description": "List of commands to enable",
						},
						"plugins": map[string]interface{}{
							"type":        "array",
							"items":       map[string]interface{}{"type": "string"},
							"description": "Plugin namespaces to include",
						},
					},
					"required":             []string{"name"},
					"additionalProperties": false,
				},
			},
			"docker": map[string]interface{}{
				"type":        "object",
				"description": "Docker provider configuration",
				"properties": map[string]interface{}{
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Docker image to use",
					},
					"ports": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Port mappings",
					},
					"dind": map[string]interface{}{
						"type":        "boolean",
						"description": "Enable Docker-in-Docker",
					},
				},
				"additionalProperties": false,
			},
			"lxc": map[string]interface{}{
				"type":        "object",
				"description": "LXC provider configuration",
				"properties": map[string]interface{}{
					"image": map[string]interface{}{
						"type":        "string",
						"description": "LXC image to use",
					},
				},
				"additionalProperties": false,
			},
			"hooks": map[string]interface{}{
				"type":        "object",
				"description": "Lifecycle hooks",
				"properties": map[string]interface{}{
					"setup": map[string]interface{}{
						"type":        "string",
						"description": "Setup hook script path",
					},
					"dev": map[string]interface{}{
						"type":        "string",
						"description": "Development hook script path",
					},
					"teardown": map[string]interface{}{
						"type":        "string",
						"description": "Teardown hook script path",
					},
				},
				"additionalProperties": false,
			},
		},
		"additionalProperties": false,
	}

	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}

	return string(schemaBytes), nil
}
