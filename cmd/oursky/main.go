func extractConfigToPlugin(pluginName string, extractRules, extractSkills, extractCommands bool) error {
	pluginDir := filepath.Join(".vendatta", "plugins", pluginName)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	pluginConfig := map[string]interface{}{
		"name":        pluginName,
		"version":     "1.0.0",
		"description": fmt.Sprintf("Extracted plugin from %s project", pluginName),
		"author":      "Vendatta Config Extractor",
	}

	// Extract rules
	if extractRules {
		rulesDir := filepath.Join(".vendatta", "templates", "rules")
		if _, err := os.Stat(rulesDir); err == nil {
			if err := extractDirectory(rulesDir, filepath.Join(pluginDir, "rules"), "rules"); err != nil {
				return fmt.Errorf("failed to extract rules: %w", err)
			}
			pluginConfig["rules"] = []string{"*"} // Include all extracted rules
		}
	}

	// Extract skills
	if extractSkills {
		skillsDir := filepath.Join(".vendatta", "templates", "skills")
		if _, err := os.Stat(skillsDir); err == nil {
			if err := extractDirectory(skillsDir, filepath.Join(pluginDir, "skills"), "skills"); err != nil {
				return fmt.Errorf("failed to extract skills: %w", err)
			}
			pluginConfig["skills"] = []string{"*"} // Include all extracted skills
		}
	}

	// Extract commands
	if extractCommands {
		commandsDir := filepath.Join(".vendatta", "templates", "commands")
		if _, err := os.Stat(commandsDir); err == nil {
			if err := extractDirectory(commandsDir, filepath.Join(pluginDir, "commands"), "commands"); err != nil {
				return fmt.Errorf("failed to extract commands: %w", err)
			}
			pluginConfig["commands"] = []string{"*"} // Include all extracted commands
		}
	}

	// Write plugin manifest
	manifestPath := filepath.Join(pluginDir, "plugin.yaml")
	if err := writePluginManifest(manifestPath, pluginConfig); err != nil {
		return fmt.Errorf("failed to write plugin manifest: %w", err)
	}

	fmt.Printf("✅ Successfully extracted configuration to plugin: %s\n", pluginName)
	fmt.Printf("📁 Plugin location: %s\n", pluginDir)
	fmt.Printf("📄 Manifest: %s\n", manifestPath)

	return nil
}

func extractDirectory(srcDir, dstDir, configType string) error {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dstDir, relPath)

		// Read source file
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Write to destination
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return err
		}

		fmt.Printf("📋 Extracted %s: %s\n", configType, relPath)
		return nil
	})
}

func writePluginManifest(path string, config map[string]interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}