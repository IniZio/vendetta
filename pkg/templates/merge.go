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
	Skills   map[string]interface{} `yaml:"skills"`
	Rules    map[string]interface{} `yaml:"rules"`
	Commands map[string]interface{} `yaml:"commands"`
}

// Merge merges templates from multiple sources with priority order:
// 1. Base templates (.vendatta/templates/)
// 2. Template repos (.vendatta/template-repos/*/templates/)
// 3. Agent overrides (.vendatta/agents/)
func (m *Manager) Merge(baseDir string) (*TemplateData, error) {
	data := &TemplateData{
		Skills:   make(map[string]interface{}),
		Rules:    make(map[string]interface{}),
		Commands: make(map[string]interface{}),
	}

	// 1. Load base templates
	baseTemplatesDir := filepath.Join(baseDir, "templates")
	if err := m.loadTemplatesFromDir(baseTemplatesDir, data); err != nil {
		return nil, fmt.Errorf("failed to load base templates: %w", err)
	}

	// 2. Load template repos
	templateReposDir := filepath.Join(baseDir, "template-repos")
	if err := m.loadTemplateRepos(templateReposDir, data); err != nil {
		return nil, fmt.Errorf("failed to load template repos: %w", err)
	}

	// 3. Load agent overrides
	agentsDir := filepath.Join(baseDir, "agents")
	if err := m.loadTemplatesFromDir(agentsDir, data); err != nil {
		return nil, fmt.Errorf("failed to load agent templates: %w", err)
	}

	return data, nil
}

func (m *Manager) loadTemplatesFromDir(dir string, data *TemplateData) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, skip
	}

	// Load skills
	skillsDir := filepath.Join(dir, "skills")
	if err := m.loadTemplateFiles(skillsDir, data.Skills); err != nil {
		return fmt.Errorf("failed to load skills from %s: %w", skillsDir, err)
	}

	// Load rules
	rulesDir := filepath.Join(dir, "rules")
	if err := m.loadTemplateFiles(rulesDir, data.Rules); err != nil {
		return fmt.Errorf("failed to load rules from %s: %w", rulesDir, err)
	}

	// Load commands
	commandsDir := filepath.Join(dir, "commands")
	if err := m.loadTemplateFiles(commandsDir, data.Commands); err != nil {
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

		repoDir := filepath.Join(reposDir, entry.Name())
		templatesDir := filepath.Join(repoDir, "templates")

		if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
			continue // This repo doesn't have templates
		}

		if err := m.loadTemplatesFromDir(templatesDir, data); err != nil {
			return fmt.Errorf("failed to load templates from repo %s: %w", entry.Name(), err)
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

		if d.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var templateData map[string]interface{}
		if err := yaml.Unmarshal(content, &templateData); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Recursively merge with existing data (key is the filename without extension)
		recursiveMerge(target, templateData)

		return nil
	})
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
