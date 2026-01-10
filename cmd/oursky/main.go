package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vibegear/oursky/pkg/config"
	"github.com/vibegear/oursky/pkg/ctrl"
	"github.com/vibegear/oursky/pkg/provider"
	"github.com/vibegear/oursky/pkg/provider/docker"
	"github.com/vibegear/oursky/pkg/templates"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

func updateVendatta() error {
	const repo = "IniZio/vendatta"

	// Detect platform
	osName := runtime.GOOS
	arch := runtime.GOARCH

	var binaryName string
	switch osName {
	case "linux", "darwin":
		binaryName = fmt.Sprintf("oursky-%s-%s", osName, arch)
	case "windows":
		binaryName = fmt.Sprintf("oursky-%s-%s.exe", osName, arch)
	default:
		return fmt.Errorf("unsupported OS: %s", osName)
	}

	// Get latest release
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo))
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return fmt.Errorf("failed to parse release: %w", err)
	}

	fmt.Printf("Latest version: %s\n", release.TagName)

	// Download binary
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, release.TagName, binaryName)
	fmt.Printf("Downloading from %s\n", downloadURL)

	resp, err = http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Get current binary path
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Create temp file in /tmp
	tempPath := fmt.Sprintf("/tmp/oursky-update-%d", os.Getpid())
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()
	defer os.Remove(tempPath)

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	// Make executable on Unix
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempPath, 0755); err != nil {
			return fmt.Errorf("failed to make executable: %w", err)
		}
	}

	// Check if we can write to the directory
	dir := filepath.Dir(currentPath)
	testFile := filepath.Join(dir, ".vendatta_write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("cannot write to %s. Please run with sudo or reinstall to a user-writable directory like ~/.local/bin", dir)
	}
	f.Close()
	os.Remove(testFile)

	// Backup current binary
	backupPath := currentPath + ".backup"
	if err := os.Rename(currentPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Replace with new binary
	if err := os.Rename(tempPath, currentPath); err != nil {
		// Try to restore backup
		os.Rename(backupPath, currentPath)
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("Successfully updated to %s\n", release.TagName)
	fmt.Printf("Backup saved at %s\n", backupPath)

	return nil
}

func syncRemoteTarget(targetName string) error {
	cfg, err := config.LoadConfig(".vendatta/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var target *config.Remote
	for _, t := range cfg.SyncTargets {
		if t.Name == targetName {
			target = &t
			break
		}
	}
	if target == nil {
		return fmt.Errorf("sync target '%s' not found in config", targetName)
	}

	fmt.Printf("Syncing .vendatta to '%s' (%s)...\n", target.Name, target.URL)

	fmt.Println("Pulling latest changes from origin...")
	if err := runGit("pull", "origin", "main"); err != nil {
		return fmt.Errorf("failed to pull from origin: %w", err)
	}

	fmt.Printf("Adding/updating remote '%s'...\n", target.Name)
	if err := runGit("remote", "add", target.Name, target.URL); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to add remote: %w", err)
		}
		if err := runGit("remote", "set-url", target.Name, target.URL); err != nil {
			return fmt.Errorf("failed to update remote: %w", err)
		}
	}

	fmt.Println("Creating filtered branch with .vendatta content...")
	if err := runGit("checkout", "-b", "temp-sync"); err != nil {
		return fmt.Errorf("failed to create temp branch: %w", err)
	}
	if err := runGit("rm", "-rf", "--cached", "."); err != nil {
		return fmt.Errorf("failed to clear index: %w", err)
	}
	if err := runGit("add", ".vendatta"); err != nil {
		return fmt.Errorf("failed to add .vendatta: %w", err)
	}
	if err := runGit("commit", "-m", "Sync .vendatta directory"); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	if err := runGit("push", target.Name, "temp-sync:main"); err != nil {
		return fmt.Errorf("failed to push .vendatta: %w", err)
	}
	if err := runGit("checkout", "main"); err != nil {
		return fmt.Errorf("failed to checkout main: %w", err)
	}
	if err := runGit("branch", "-D", "temp-sync"); err != nil {
		return fmt.Errorf("failed to delete temp branch: %w", err)
	}

	fmt.Printf("Successfully synced .vendatta to '%s'!\n", target.Name)
	return nil
}

func syncAllRemotes() error {
	cfg, err := config.LoadConfig(".vendatta/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.SyncTargets) == 0 {
		fmt.Println("No sync targets configured in .vendatta/config.yaml")
		return nil
	}

	for _, target := range cfg.SyncTargets {
		fmt.Printf("Syncing to target '%s' (%s)...\n", target.Name, target.URL)
		if err := syncRemoteTarget(target.Name); err != nil {
			return fmt.Errorf("failed to sync target '%s': %w", target.Name, err)
		}
		fmt.Printf("Successfully synced target '%s'!\n", target.Name)
	}

	fmt.Println("All configured sync targets updated successfully!")
	return nil
}

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	var providers []provider.Provider
	dProvider, err := docker.NewDockerProvider()
	if err == nil {
		providers = append(providers, dProvider)
	}

	controller := ctrl.NewBaseController(providers, nil)

	rootCmd := &cobra.Command{
		Use:   "vendatta",
		Short: "Vendatta Dev Environment Manager",
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize .vendatta in the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Init(context.Background())
		},
	}

	workspaceCmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage isolated development workspaces",
	}

	workspaceCreateCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new workspace with agent configs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.WorkspaceCreate(context.Background(), args[0])
		},
	}

	workspaceUpCmd := &cobra.Command{
		Use:   "up [name]",
		Short: "Start workspace services (auto-detects workspace if in worktree)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return controller.WorkspaceUp(context.Background(), name)
		},
	}

	workspaceDownCmd := &cobra.Command{
		Use:   "down [name]",
		Short: "Stop workspace services (auto-detects workspace if in worktree)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return controller.WorkspaceDown(context.Background(), name)
		},
	}

	workspaceShellCmd := &cobra.Command{
		Use:   "shell [name]",
		Short: "Open interactive shell in workspace (auto-detects workspace if in worktree)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return controller.WorkspaceShell(context.Background(), name)
		},
	}

	workspaceListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all workspaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.WorkspaceList(context.Background())
		},
	}

	workspaceRmCmd := &cobra.Command{
		Use:   "rm <name>",
		Short: "Remove workspace entirely",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.WorkspaceRm(context.Background(), args[0])
		},
	}

	// Add subcommands to workspace command group
	workspaceCmd.AddCommand(workspaceCreateCmd, workspaceUpCmd, workspaceDownCmd, workspaceShellCmd, workspaceListCmd, workspaceRmCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List active sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			sessions, err := controller.List(context.Background())
			if err != nil {
				return err
			}
			for _, s := range sessions {
				fmt.Printf("%s\t%s\t%s\n", s.Labels["oursky.session.id"], s.Provider, s.Status)
			}
			return nil
		},
	}

	killCmd := &cobra.Command{
		Use:   "kill [session-id]",
		Short: "Stop and destroy a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Kill(context.Background(), args[0])
		},
	}

	templatesCmd := &cobra.Command{
		Use:   "templates",
		Short: "Manage AI agent templates",
	}

	templatesPullCmd := &cobra.Command{
		Use:   "pull [url]",
		Short: "Pull templates from a git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := templates.NewManager(".vendatta")
			repo := templates.TemplateRepo{
				URL: args[0],
			}

			// Check if branch flag is provided
			if branch, _ := cmd.Flags().GetString("branch"); branch != "" {
				repo.Branch = branch
			}

			return manager.PullRepo(repo)
		},
	}
	templatesPullCmd.Flags().String("branch", "", "Branch to pull from")

	templatesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List pulled template repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := templates.NewManager(".vendatta")
			repos, err := manager.ListRepos()
			if err != nil {
				return err
			}

			if len(repos) == 0 {
				fmt.Println("No template repositories pulled")
				return nil
			}

			fmt.Println("Pulled template repositories:")
			for _, repo := range repos {
				fmt.Printf("  - %s\n", repo)
			}
			return nil
		},
	}

	templatesMergeCmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge templates from all sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := templates.NewManager(".vendatta")
			data, err := manager.Merge(".vendatta")
			if err != nil {
				return err
			}

			fmt.Printf("Merged %d skills, %d rules, %d commands\n",
				len(data.Skills), len(data.Rules), len(data.Commands))
			return nil
		},
	}

	templatesCmd.AddCommand(templatesPullCmd, templatesListCmd, templatesMergeCmd)

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Vendatta configuration and templates",
	}

	configPullCmd := &cobra.Command{
		Use:   "pull <url>",
		Short: "Pull templates from a remote Git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := templates.NewManager(".vendatta")
			repo := templates.TemplateRepo{
				URL: args[0],
			}

			// Check if branch flag is provided
			if branch, _ := cmd.Flags().GetString("branch"); branch != "" {
				repo.Branch = branch
			}

			return manager.PullRepo(repo)
		},
	}
	configPullCmd.Flags().String("branch", "", "Branch to pull from")

	configListCmd := &cobra.Command{
		Use:   "list",
		Short: "List pulled template repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := templates.NewManager(".vendatta")
			repos, err := manager.ListRepos()
			if err != nil {
				return err
			}

			if len(repos) == 0 {
				fmt.Println("No template repositories pulled")
				return nil
			}

			fmt.Println("Pulled template repositories:")
			for _, repo := range repos {
				fmt.Printf("  - %s\n", repo)
			}
			return nil
		},
	}

	configSyncCmd := &cobra.Command{
		Use:   "sync [target-name]",
		Short: "Sync .vendatta directory to a configured remote target",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncRemoteTarget(args[0])
		},
	}

	configSyncAllCmd := &cobra.Command{
		Use:   "sync-all",
		Short: "Sync all configured remotes from .vendatta/config.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncAllRemotes()
		},
	}

	configCmd.AddCommand(configPullCmd, configListCmd, configSyncCmd, configSyncAllCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update vendatta to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateVendatta()
		},
	}

	rootCmd.AddCommand(initCmd, workspaceCmd, listCmd, killCmd, templatesCmd, configCmd, updateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
