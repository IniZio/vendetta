package templates

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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
	os.MkdirAll(filepath.Join(baseDir, "remotes/repo1/templates/skills"), 0755)

	os.WriteFile(filepath.Join(baseDir, "templates/skills/skill1.yaml"), []byte("name: skill1\nversion: \"1.0\""), 0644)
	os.WriteFile(filepath.Join(baseDir, "templates/rules/rule1.md"), []byte("---\ntitle: rule1\n---\nRule 1 Content"), 0644)
	os.WriteFile(filepath.Join(baseDir, "remotes/repo1/templates/skills/skill1.yaml"), []byte("version: \"2.0\"\nnew_field: value"), 0644)

	m := NewManager(baseDir)
	data, err := m.Merge(baseDir, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, data)

	basePlugin := data.Plugins["base"]
	assert.NotNil(t, basePlugin)

	skill1 := basePlugin.Skills["skill1"].(map[string]interface{})
	assert.Equal(t, "skill1", skill1["name"])
	assert.Equal(t, "1.0", skill1["version"])
	// new_field from remote should not be present since base takes precedence
	assert.NotContains(t, skill1, "new_field")

	rule1 := basePlugin.Rules["rule1"].(map[string]interface{})
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

func TestGetRepoDir(t *testing.T) {
	m := NewManager("/base")
	repoDir := m.GetRepoDir("my-repo")
	assert.Equal(t, "/base/remotes/my-repo", repoDir)
}

func TestPullWithoutUpdate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "templates-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	m := NewManager(tempDir)

	// Test that GetRepoDir works correctly
	repoDir := m.GetRepoDir("my-repo")
	assert.Equal(t, filepath.Join(tempDir, "remotes", "my-repo"), repoDir)

	// Test PullWithoutUpdate with non-existent repo (will fail on clone, which is expected)
	// This verifies the method structure is correct
	repo := TemplateRepo{
		URL:    "https://github.com/owner/repo",
		Branch: "main",
	}

	// Verify the repo directory path is computed correctly before any clone attempt
	expectedRepoDir := filepath.Join(tempDir, "remotes", "repo")
	assert.Equal(t, expectedRepoDir, m.GetRepoDir("repo"))

	// Test that existing directories are detected (simulate cached repo)
	os.MkdirAll(expectedRepoDir, 0755)
	err = m.PullWithoutUpdate(repo)
	assert.NoError(t, err) // Should succeed because repo already exists
}

func TestGetRepoSHA(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "templates-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	m := NewManager(tempDir)

	// Test with non-existent repo
	repo := TemplateRepo{URL: "https://github.com/owner/nonexistent"}
	_, err = m.GetRepoSHA(repo)
	assert.Error(t, err)

	// Create a fake git repo for testing
	repoDir := m.GetRepoDir("test-repo")
	os.MkdirAll(repoDir, 0755)

	// Initialize a bare git repo to get a valid SHA
	r, err := git.PlainInit(repoDir, false)
	assert.NoError(t, err)

	// Create a commit
	wt, err := r.Worktree()
	assert.NoError(t, err)

	os.WriteFile(filepath.Join(repoDir, "test.txt"), []byte("test"), 0644)
	_, err = wt.Add("test.txt")
	assert.NoError(t, err)

	_, err = wt.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Now test GetRepoSHA
	sha, err := m.GetRepoSHA(TemplateRepo{URL: "https://github.com/owner/test-repo"})
	assert.NoError(t, err)
	assert.Len(t, sha, 40) // SHA length
}
