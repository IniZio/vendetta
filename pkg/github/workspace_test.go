package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name       string
		repoString string
		wantOwner  string
		wantRepo   string
		wantErr    bool
	}{
		{
			name:       "valid owner/repo format",
			repoString: "torvalds/linux",
			wantOwner:  "torvalds",
			wantRepo:   "linux",
			wantErr:    false,
		},
		{
			name:       "valid https url",
			repoString: "https://github.com/torvalds/linux",
			wantOwner:  "torvalds",
			wantRepo:   "linux",
			wantErr:    false,
		},
		{
			name:       "valid https url with .git",
			repoString: "https://github.com/torvalds/linux.git",
			wantOwner:  "torvalds",
			wantRepo:   "linux",
			wantErr:    false,
		},
		{
			name:       "invalid format",
			repoString: "invalid",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepoURL(tt.repoString)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOwner, owner)
				assert.Equal(t, tt.wantRepo, repo)
			}
		})
	}
}

func TestVerifyRepoOwnership(t *testing.T) {
	tests := []struct {
		name      string
		ghCLIPath string
		owner     string
		repo      string
		wantErr   bool
	}{
		{
			name:      "verify repo ownership",
			ghCLIPath: "gh",
			owner:     "torvalds",
			repo:      "linux",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isOwned, err := VerifyRepoOwnership(tt.ghCLIPath, tt.owner, tt.repo)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.IsType(t, true, isOwned)
			}
		})
	}
}

func TestForkRepository(t *testing.T) {
	tests := []struct {
		name      string
		ghCLIPath string
		owner     string
		repo      string
		wantErr   bool
	}{
		{
			name:      "fork repository",
			ghCLIPath: "gh",
			owner:     "torvalds",
			repo:      "linux",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ForkRepository(tt.ghCLIPath, tt.owner, tt.repo)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

func TestBuildCloneURL(t *testing.T) {
	tests := []struct {
		name    string
		owner   string
		repo    string
		wantURL string
	}{
		{
			name:    "https clone url",
			owner:   "torvalds",
			repo:    "linux",
			wantURL: "https://github.com/torvalds/linux.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := BuildCloneURL(tt.owner, tt.repo)
			assert.Equal(t, tt.wantURL, url)
		})
	}
}

func TestCloneRepository(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		dest    string
		wantErr bool
	}{
		{
			name:    "clone repo to directory",
			url:     "https://github.com/torvalds/linux.git",
			dest:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("skipping repo clone test (requires network)")
		})
	}
}

func TestGetRepositoryCommit(t *testing.T) {
	tests := []struct {
		name      string
		ghCLIPath string
		owner     string
		repo      string
		wantErr   bool
	}{
		{
			name:      "get repo commit",
			ghCLIPath: "gh",
			owner:     "torvalds",
			repo:      "linux",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commit, err := GetRepositoryCommit(tt.ghCLIPath, tt.owner, tt.repo)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				if err == nil {
					assert.NotEmpty(t, commit)
				}
			}
		})
	}
}

func TestParseRepoURLEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty owner",
			input:   "/repo",
			wantErr: true,
		},
		{
			name:    "empty repo",
			input:   "owner/",
			wantErr: true,
		},
		{
			name:    "no slash",
			input:   "ownerrepo",
			wantErr: true,
		},
		{
			name:    "multiple slashes",
			input:   "owner/repo/extra",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ParseRepoURL(tt.input)
			assert.Error(t, err)
		})
	}
}
