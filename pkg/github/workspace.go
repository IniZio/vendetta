package github

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

func ParseRepoURL(repoString string) (owner string, repo string, err error) {
	repoString = strings.TrimSpace(repoString)

	if strings.Contains(repoString, "://") {
		u, err := url.Parse(repoString)
		if err != nil {
			return "", "", fmt.Errorf("invalid repo URL: %w", err)
		}

		path := strings.TrimPrefix(u.Path, "/")
		path = strings.TrimSuffix(path, ".git")

		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid repo URL format: expected owner/repo")
		}

		return parts[0], parts[1], nil
	}

	parts := strings.Split(repoString, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo format: expected owner/repo")
	}

	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repo format: owner and repo must not be empty")
	}

	return parts[0], parts[1], nil
}

func BuildCloneURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

func CloneRepository(url, destDir string) error {
	args := []string{"clone", url}
	if destDir != "" {
		args = append(args, destDir)
	}

	cmd := exec.Command("git", args...)

	var stderr strings.Builder
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}
