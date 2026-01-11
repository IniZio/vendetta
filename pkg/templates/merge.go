package templates

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
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

func (m *Manager) Merge(baseDir string) (*TemplateData, error) {
	data := &TemplateData{
		Plugins: make(map[string]*PluginData),
	}

	baseTemplatesDir := filepath.Join(baseDir, "templates")
	basePlugin := m.getOrCreatePlugin(data, "base")
	if err := m.loadTemplatesFromDir(baseTemplatesDir, basePlugin); err != nil {
		return nil, fmt.Errorf("failed to load base templates: %w", err)
	}

	templateReposDir := filepath.Join(baseDir, "remotes")
	if err := m.loadTemplateRepos(templateReposDir, data); err != nil {
		return nil, fmt.Errorf("failed to load template repos: %w", err)
	}

	pluginsDir := filepath.Join(baseDir, "plugins")
	if err := m.loadPluginTemplates(pluginsDir, data); err != nil {
		return nil, fmt.Errorf("failed to load plugin templates: %w", err)
	}

	agentsDir := filepath.Join(baseDir, "agents")
	if err := m.applyAgentOverrides(agentsDir, data); err != nil {
		return nil, fmt.Errorf("failed to apply agent overrides: %w", err)
	}

	return data, nil
}

func (m *Manager) getOrCreatePlugin(data *TemplateData, name string) *PluginData {
	if plugin, ok := data.Plugins[name]; ok {
		return plugin
	}
	plugin := &PluginData{
		Skills:   make(map[string]interface{}),
		Rules:    make(map[string]interface{}),
		Commands: make(map[string]interface{}),
	}
	data.Plugins[name] = plugin
	return plugin
}

func (m *Manager) loadTemplatesFromDir(dir string, plugin *PluginData) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, skip
	}

	// Load skills
	skillsDir := filepath.Join(dir, "skills")
	if err := m.loadTemplateFiles(skillsDir, plugin.Skills); err != nil {
		return fmt.Errorf("failed to load skills from %s: %w", skillsDir, err)
	}

	// Load rules
	rulesDir := filepath.Join(dir, "rules")
	if err := m.loadTemplateFiles(rulesDir, plugin.Rules); err != nil {
		return fmt.Errorf("failed to load rules from %s: %w", rulesDir, err)
	}

	// Load commands
	commandsDir := filepath.Join(dir, "commands")
	if err := m.loadTemplateFiles(commandsDir, plugin.Commands); err != nil {
		return fmt.Errorf("failed to load commands from %s: %w", commandsDir, err)
	}

	return nil
}

func (m *Manager) loadTemplateRepos(reposDir string, data *TemplateData) error {
	if _, err := os.Stat(reposDir); os.IsNotExist(err) {
		return nil // No repos directory
	}

	entries, err := os.ReadDir(reposDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		repoName := entry.Name()
		repoDir := filepath.Join(reposDir, repoName)

		pluginsDir := filepath.Join(repoDir, "plugins")
		if _, err := os.Stat(pluginsDir); err == nil {
			if err := m.loadPluginTemplates(pluginsDir, data); err != nil {
				return fmt.Errorf("failed to load plugins from repo %s: %w", repoName, err)
			}
		}

		templatesDir := filepath.Join(repoDir, "templates")
		if _, err := os.Stat(templatesDir); err == nil {
			plugin := m.getOrCreatePlugin(data, repoName)
			if err := m.loadTemplatesFromDir(templatesDir, plugin); err != nil {
				return fmt.Errorf("failed to load templates from repo %s: %w", repoName, err)
			}
		}
	}

	return nil
}

func (m *Manager) loadPluginTemplates(pluginsDir string, data *TemplateData) error {
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		return nil // No plugins directory
	}

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginName := entry.Name()
		pluginTemplatesDir := filepath.Join(pluginsDir, pluginName, "templates")

		if _, err := os.Stat(pluginTemplatesDir); os.IsNotExist(err) {
			continue
		}

		plugin := m.getOrCreatePlugin(data, pluginName)
		if err := m.loadTemplatesFromDir(pluginTemplatesDir, plugin); err != nil {
			return fmt.Errorf("failed to load templates from plugin %s: %w", pluginName, err)
		}
	}

	return nil
}

func (m *Manager) loadTemplateFiles(dir string, target map[string]interface{}) error {
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

		isMd := strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".mdc")

		if !isMd {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var templateData map[string]interface{}
		data, mdContent := parseMarkdown(content)
		filename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		ruleData := make(map[string]interface{})
		for k, v := range data {
			ruleData[k] = v
		}
		ruleData["content"] = mdContent

		templateData = map[string]interface{}{
			filename: ruleData,
		}

		recursiveMerge(target, templateData)

		return nil
	})
}

func parseMarkdown(content []byte) (map[string]interface{}, string) {
	str := string(content)
	if !strings.HasPrefix(str, "---\n") {
		return nil, str
	}
	parts := strings.SplitN(str[4:], "\n---\n", 2)
	if len(parts) < 2 {
		return nil, str
	}
	var data map[string]interface{}
	_ = yaml.Unmarshal([]byte(parts[0]), &data)
	return data, parts[1]
}

// recursiveMerge merges source into dest, following chezmoi's pattern:
// - Maps are merged recursively
// - Other types replace
func recursiveMerge(dest, source map[string]interface{}) {
	for key, sourceValue := range source {
		destValue, exists := dest[key]
		if !exists {
			dest[key] = sourceValue
			continue
		}

		// Try to merge maps recursively
		destMap, destIsMap := destValue.(map[string]interface{})
		sourceMap, sourceIsMap := sourceValue.(map[string]interface{})

		if destIsMap && sourceIsMap {
			recursiveMerge(destMap, sourceMap)
		} else {
			// Replace with source value
			dest[key] = sourceValue
		}
	}
}

// RenderTemplate renders a template with the given data
func (m *Manager) RenderTemplate(templateContent string, data interface{}) (string, error) {
	tmpl, err := template.New("template").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
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
