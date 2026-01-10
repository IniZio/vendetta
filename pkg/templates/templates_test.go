package templates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "templates-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	baseDir := tempDir
	os.MkdirAll(filepath.Join(baseDir, "templates/skills"), 0755)
	os.MkdirAll(filepath.Join(baseDir, "templates/rules"), 0755)
	os.MkdirAll(filepath.Join(baseDir, "template-repos/repo1/templates/skills"), 0755)

	os.WriteFile(filepath.Join(baseDir, "templates/skills/skill1.yaml"), []byte("name: skill1\nversion: \"1.0\""), 0644)
	os.WriteFile(filepath.Join(baseDir, "templates/rules/rule1.md"), []byte("---\ntitle: rule1\n---\nRule 1 Content"), 0644)
	os.WriteFile(filepath.Join(baseDir, "template-repos/repo1/templates/skills/skill1.yaml"), []byte("version: \"2.0\"\nnew_field: value"), 0644)

	m := NewManager(baseDir)
	data, err := m.Merge(baseDir)

	assert.NoError(t, err)
	assert.NotNil(t, data)

	skill1 := data.Skills["skill1"].(map[string]interface{})
	assert.Equal(t, "skill1", skill1["name"])
	assert.Equal(t, "2.0", skill1["version"])
	assert.Equal(t, "value", skill1["new_field"])

	rule1 := data.Rules["rule1"].(map[string]interface{})
	assert.Equal(t, "rule1", rule1["title"])
	assert.Equal(t, "Rule 1 Content", rule1["content"])
}

func TestRenderTemplate(t *testing.T) {
	m := NewManager("")
	tmpl := "Hello {{.Name}}!"
	data := map[string]string{"Name": "World"}

	result, err := m.RenderTemplate(tmpl, data)
	assert.NoError(t, err)
	assert.Equal(t, "Hello World!", result)
}
