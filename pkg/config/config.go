package config

import (
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
	MCP         struct {
		Enabled bool   `yaml:"enabled,omitempty"`
		Port    int    `yaml:"port,omitempty"`
		Host    string `yaml:"host,omitempty"`
	} `yaml:"mcp,omitempty"`
	Docker struct {
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
	Host           string
	Port           int
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

// GenerateAgentConfigs generates agent configuration files from templates
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
			rulesFormat:  "agents.md",
			rulesDir:     "",
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
	}

	renderData := RenderData{
		Host:        c.MCP.Host,
		Port:        c.MCP.Port,
		AuthToken:   "token", // TODO: generate or configure
		ProjectName: c.Name,
		// TODO: populate RulesConfig, etc. as JSON
	}

	var gitignorePatterns []string
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
			agentRulesDir := filepath.Join(".vendatta", "agents", agent.Name)
			agentMerged := *merged
			agentMerged.Rules = make(map[string]interface{})
			for k, v := range merged.Rules {
				agentMerged.Rules[k] = v
			}

			manager := templates.NewManager(".vendatta")
			if err := manager.LoadAgentRules(agentRulesDir, &agentMerged); err == nil {
				if err := c.generateAgentRules(worktreePath, cfg.rulesFormat, cfg.rulesDir, &agentMerged); err != nil {
					return fmt.Errorf("failed to generate rules for %s: %w", agent.Name, err)
				}
			} else {
				if err := c.generateAgentRules(worktreePath, cfg.rulesFormat, cfg.rulesDir, merged); err != nil {
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
	if merged == nil || len(merged.Rules) == 0 {
		return nil
	}

	absRulesDir := filepath.Join(worktreePath, rulesDir)
	if rulesDir != "" {
		if err := os.MkdirAll(absRulesDir, 0755); err != nil {
			return err
		}
	}

	switch format {
	case "mdc":
		for name, data := range merged.Rules {
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

			outputPath := filepath.Join(absRulesDir, name+".mdc")
			if err := os.WriteFile(outputPath, []byte(builder.String()), 0644); err != nil {
				return err
			}
		}
	case "agents.md":
		var builder strings.Builder
		builder.WriteString("# PROJECT KNOWLEDGE BASE\n\n")
		builder.WriteString(fmt.Sprintf("**Generated:** %s\n", strings.ToUpper(c.Name)))
		builder.WriteString("\n")
		for name, data := range merged.Rules {
			ruleMap, ok := data.(map[string]interface{})
			if !ok {
				continue
			}
			content, _ := ruleMap["content"].(string)
			title, _ := ruleMap["title"].(string)
			if title == "" {
				title = name
			}
			builder.WriteString(fmt.Sprintf("## %s\n\n", title))
			builder.WriteString(content)
			builder.WriteString("\n\n")
		}
		outputPath := filepath.Join(worktreePath, "AGENTS.md")
		if err := os.WriteFile(outputPath, []byte(builder.String()), 0644); err != nil {
			return err
		}
	}
	return nil
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
