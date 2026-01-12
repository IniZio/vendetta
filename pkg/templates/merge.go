package templates

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateData represents the merged template data
type TemplateData struct {
	Plugins map[string]*PluginData `yaml:"plugins"`
}

type PluginData struct {
	Skills   map[string]interface{} `yaml:"skills"`
	Rules    map[string]interface{} `yaml:"rules"`
	Commands map[string]interface{} `yaml:"commands"`
}

type PluginManifest struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
	Conditions  []struct {
		File string `yaml:"file"`
	} `yaml:"conditions,omitempty"`
}

func (m *Manager) Merge(baseDir string, enabledPlugins []string, extends []string) (*TemplateData, error) {
	data := &TemplateData{
		Plugins: make(map[string]*PluginData),
	}

	// Load extends as base configs
	if err := m.loadExtends(baseDir, extends, data); err != nil {
		return nil, fmt.Errorf("failed to load extends: %w", err)
	}

	// Load remote template repos
	templateReposDir := filepath.Join(baseDir, "remotes")
	if err := m.loadTemplateRepos(templateReposDir, data); err != nil {
		return nil, fmt.Errorf("failed to load template repos: %w", err)
	}

	// Load local plugin templates
	projectRoot := filepath.Dir(baseDir)
	pluginsDir := filepath.Join(baseDir, "plugins")
	if err := m.loadPluginTemplates(pluginsDir, projectRoot, data); err != nil {
		return nil, fmt.Errorf("failed to load plugin templates: %w", err)
	}

	// Load base templates last to ensure local rules take precedence
	baseTemplatesDir := filepath.Join(baseDir, "templates")
	basePlugin := m.getOrCreatePlugin(data, "base")
	if err := m.loadTemplatesFromDir(baseTemplatesDir, basePlugin); err != nil {
		return nil, fmt.Errorf("failed to load base templates: %w", err)
	}

	agentsDir := filepath.Join(baseDir, "agents")
	if err := m.applyAgentOverrides(agentsDir, data); err != nil {
		return nil, fmt.Errorf("failed to load agent overrides: %w", err)
	}

	return data, nil
}

func (m *Manager) loadExtends(baseDir string, extends []string, data *TemplateData) error {
	for _, extend := range extends {
		parts := strings.Split(extend, "/")
		if len(parts) != 2 {
			continue
		}
		// TODO: Implement fetching from GitHub
	}
	return nil
}

// Stub implementations for missing methods - TODO: implement properly
func (m *Manager) loadTemplateRepos(templateReposDir string, data *TemplateData) error {
	// Stub implementation
	return nil
}

func (m *Manager) loadPluginTemplates(pluginsDir, projectRoot string, data *TemplateData) error {
	// Stub implementation
	return nil
}

func (m *Manager) getOrCreatePlugin(data *TemplateData, name string) *PluginData {
	// Stub implementation
	if data.Plugins == nil {
		data.Plugins = make(map[string]*PluginData)
	}
	if data.Plugins[name] == nil {
		data.Plugins[name] = &PluginData{
			Skills:   make(map[string]interface{}),
			Rules:    make(map[string]interface{}),
			Commands: make(map[string]interface{}),
		}
	}
	return data.Plugins[name]
}

func (m *Manager) loadTemplatesFromDir(dir string, plugin *PluginData) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) < 1 {
			return nil
		}
		templateType := parts[0]

		filename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var target map[string]interface{}
		switch templateType {
		case "skills":
			target = plugin.Skills
		case "rules":
			target = plugin.Rules
		case "commands":
			target = plugin.Commands
		default:
			return nil
		}

		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			var data interface{}
			if err := yaml.Unmarshal(content, &data); err != nil {
				return fmt.Errorf("failed to parse YAML %s: %w", path, err)
			}
			target[filename] = data
		} else if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".mdc") {
			frontmatter, mdContent := parseMarkdown(content)
			ruleData := make(map[string]interface{})
			for k, v := range frontmatter {
				ruleData[k] = v
			}
			ruleData["content"] = mdContent
			target[filename] = ruleData
		}

		return nil
	})
}

func parseMarkdown(content []byte) (map[string]interface{}, string) {
	contentStr := string(content)

	if !strings.HasPrefix(contentStr, "---\n") {
		return make(map[string]interface{}), contentStr
	}

	parts := strings.SplitN(contentStr, "\n---\n", 2)
	if len(parts) != 2 {
		return make(map[string]interface{}), contentStr
	}

	frontmatterStr := parts[0] + "\n"
	body := parts[1]

	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatterStr), &frontmatter); err != nil {
		return make(map[string]interface{}), contentStr
	}

	return frontmatter, body
}

func (m *Manager) applyAgentOverrides(agentsDir string, data *TemplateData) error {
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		agentName := entry.Name()
		agentDir := filepath.Join(agentsDir, agentName)

		overridePlugin := m.getOrCreatePlugin(data, "override")

		if err := m.applyOverrideForType(agentDir, "rules", overridePlugin.Rules); err != nil {
			return fmt.Errorf("failed to apply %s rules overrides: %w", agentName, err)
		}
		if err := m.applyOverrideForType(agentDir, "skills", overridePlugin.Skills); err != nil {
			return fmt.Errorf("failed to apply %s skills overrides: %w", agentName, err)
		}
		if err := m.applyOverrideForType(agentDir, "commands", overridePlugin.Commands); err != nil {
			return fmt.Errorf("failed to apply %s commands overrides: %w", agentName, err)
		}
	}

	return nil
}

func (m *Manager) applyOverrideForType(agentDir, templateType string, target map[string]interface{}) error {
	overrideDir := filepath.Join(agentDir, templateType)
	if _, err := os.Stat(overrideDir); os.IsNotExist(err) {
		return nil
	}

	return filepath.WalkDir(overrideDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		isMd := strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".mdc")
		if !isMd {
			return nil
		}

		filename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if len(strings.TrimSpace(string(content))) == 0 {
			delete(target, filename)
			return nil
		}

		var overrideData map[string]interface{}
		frontmatter, mdContent := parseMarkdown(content)
		ruleData := make(map[string]interface{})
		for k, v := range frontmatter {
			ruleData[k] = v
		}
		ruleData["content"] = mdContent
		overrideData = map[string]interface{}{
			filename: ruleData,
		}

		for key, value := range overrideData {
			target[key] = value
		}

		return nil
	})
}

// RenderTemplate renders a Go template string with the provided data
func (m *Manager) RenderTemplate(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("template").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
