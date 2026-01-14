package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vibegear/vendetta/cmd/internal"
	"github.com/vibegear/vendetta/pkg/config"
	"github.com/vibegear/vendetta/pkg/ctrl"
	"github.com/vibegear/vendetta/pkg/metrics"
	"github.com/vibegear/vendetta/pkg/provider"
	dockerProvider "github.com/vibegear/vendetta/pkg/provider/docker"
	lxcProvider "github.com/vibegear/vendetta/pkg/provider/lxc"
	"github.com/vibegear/vendetta/pkg/templates"
	"github.com/vibegear/vendetta/pkg/worktree"
)

var rootCmd = &cobra.Command{
	Use:   "vendetta",
	Short: "Isolated development environments that work with AI agents",
	Long: `Vendetta provides isolated development environments that integrate
seamlessly with AI coding assistants like Cursor, OpenCode, Claude, and others.`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new vendetta project",
	Long:  `Initialize a new vendetta project by creating the .vendetta directory and default configuration files.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx := context.Background()
		controller := createController()
		return controller.Init(ctx)
	},
}

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage workspaces",
	Long:  `Create, start, stop, and manage isolated development workspaces.`,
}

var workspaceCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new workspace",
	Long:  `Create a new workspace with the specified name. This will set up a Git worktree and generate AI agent configurations.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()
		controller := createController()
		return controller.WorkspaceCreate(ctx, args[0])
	},
}

var workspaceUpCmd = &cobra.Command{
	Use:   "up [name]",
	Short: "Start a workspace",
	Long:  `Start the specified workspace or auto-detect if no name is provided. This will create and start the isolated environment.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()
		controller := createController()
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		return controller.WorkspaceUp(ctx, name)
	},
}

var workspaceDownCmd = &cobra.Command{
	Use:   "down [name]",
	Short: "Stop a workspace",
	Long:  `Stop the specified workspace or auto-detect if no name is provided.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()
		controller := createController()
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		return controller.WorkspaceDown(ctx, name)
	},
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces",
	Long:  `List all workspaces, showing their status and provider information.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx := context.Background()
		controller := createController()
		return controller.WorkspaceList(ctx)
	},
}

var workspaceRmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "Remove a workspace",
	Long:  `Remove the specified workspace, stopping it first if it's running.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()
		controller := createController()
		return controller.WorkspaceRm(ctx, args[0])
	},
}

var workspaceShellCmd = &cobra.Command{
	Use:   "shell [name]",
	Short: "Open shell in workspace",
	Long:  `Open an interactive shell in the specified workspace or auto-detect if no name is provided.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		ctx := context.Background()
		controller := createController()
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		return controller.WorkspaceShell(ctx, name)
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply latest configuration to agent configs",
	Long:  `Apply the latest vendetta configuration to all enabled AI agent configuration files (Cursor, OpenCode, Claude, etc.).`,
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx := context.Background()
		controller := createController()
		return controller.Apply(ctx)
	},
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
	Long:  `Manage plugins: add, remove, update, and list available plugins.`,
}

var pluginUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all plugins to latest versions",
	Long:  `Update all loaded plugins to their latest versions and refresh the lockfile.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx := context.Background()
		controller := createController()
		return controller.PluginUpdate(ctx)
	},
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all loaded plugins",
	Long:  `List all currently loaded plugins with their versions and status.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		ctx := context.Background()
		controller := createController()
		return controller.PluginList(ctx)
	},
}

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Analyze usage metrics and productivity",
	Long:  `Generate reports and insights from usage logs.`,
}

var usageSummaryCmd = &cobra.Command{
	Use:   "summary [date]",
	Short: "Generate daily summary report",
	Long:  `Generate a daily summary of usage metrics and insights. Date format: YYYY-MM-DD (defaults to today).`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runUsageSummary(args)
	},
}

var usageMetricsCmd = &cobra.Command{
	Use:   "metrics [days]",
	Short: "Calculate productivity metrics",
	Long:  `Calculate detailed productivity metrics for the specified number of days (defaults to 7).`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runUsageMetrics(args)
	},
}

var usagePatternsCmd = &cobra.Command{
	Use:   "patterns [days]",
	Short: "Analyze usage patterns",
	Long:  `Analyze usage patterns and trends for the specified number of days (defaults to 7).`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runUsagePatterns(args)
	},
}

var usageBenchmarkCmd = &cobra.Command{
	Use:   "benchmark <baseline-days> <current-days>",
	Short: "Compare baseline and current metrics",
	Long:  `Compare productivity metrics between baseline period and current period.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		return runUsageBenchmark(args)
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage vendetta configuration files.`,
}

var configExtractCmd = &cobra.Command{
	Use:   "extract <plugin-name>",
	Short: "Extract configuration to plugin",
	Long: `Extract local configuration (rules, skills, commands) into a reusable plugin.
This allows teams to share their coding standards and configurations as distributable plugins.`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		pluginName := args[0]

		// Default to extracting all types
		return internal.ExtractConfigToPlugin(pluginName, true, true, true)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all extends to latest versions",
	Long: `Fetch the latest versions of all configured extends and update the lockfile.
This ensures you have the most recent templates from remote repositories.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return runUpdate()
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(pluginCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(usageCmd)
	rootCmd.AddCommand(workspaceCmd)

	// Plugin subcommands
	pluginCmd.AddCommand(pluginUpdateCmd)
	pluginCmd.AddCommand(pluginListCmd)

	// Config subcommands
	configCmd.AddCommand(configExtractCmd)

	// Usage subcommands
	usageCmd.AddCommand(usageSummaryCmd)
	usageCmd.AddCommand(usageMetricsCmd)
	usageCmd.AddCommand(usagePatternsCmd)
	usageCmd.AddCommand(usageBenchmarkCmd)

	// Workspace subcommands
	workspaceCmd.AddCommand(workspaceCreateCmd)
	workspaceCmd.AddCommand(workspaceUpCmd)
	workspaceCmd.AddCommand(workspaceDownCmd)
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceRmCmd)
	workspaceCmd.AddCommand(workspaceShellCmd)
}

func createController() ctrl.Controller {
	// Create providers
	dockerProv, err := dockerProvider.NewDockerProvider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Docker provider: %v\n", err)
		os.Exit(1)
	}

	lxcProv, err := lxcProvider.NewLXCProvider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create LXC provider: %v\n", err)
		os.Exit(1)
	}

	providers := []provider.Provider{dockerProv, lxcProv}

	// Create worktree manager
	wtManager := worktree.NewManager(".", ".vendetta/worktrees")

	// Create controller
	return ctrl.NewBaseController(providers, wtManager)
}

// runUpdate updates all extends to their latest versions
func runUpdate() error {
	fmt.Println("📦 Updating extends to latest versions...")

	// Load config
	cfg, err := config.LoadConfig(".vendetta/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create templates manager
	templateManager := templates.NewManager(".")

	// Update each extend
	updated := 0
	for _, ext := range cfg.Extends {
		extStr, ok := ext.(string)
		if !ok {
			continue
		}
		parts := strings.Split(extStr, "/")
		if len(parts) != 2 {
			continue
		}

		owner, repo := parts[0], parts[1]

		// Parse optional branch
		branch := ""
		if strings.Contains(repo, "@") {
			repoParts := strings.SplitN(repo, "@", 2)
			repo = repoParts[0]
			branch = repoParts[1]
		}

		repoURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)
		repoTemplate := templates.TemplateRepo{
			URL:    repoURL,
			Branch: branch,
		}

		fmt.Printf("  Updating %s...\n", ext)
		if err := templateManager.PullWithUpdate(repoTemplate); err != nil {
			fmt.Printf("  ⚠️  Failed to update %s: %v\n", ext, err)
			continue
		}

		// Get new SHA
		sha, err := templateManager.GetRepoSHA(repoTemplate)
		if err != nil {
			fmt.Printf("  ⚠️  Failed to get SHA for %s: %v\n", ext, err)
			continue
		}

		fmt.Printf("  ✅ %s (SHA: %s)\n", ext, sha[:7])
		updated++
	}

	if updated == 0 {
		fmt.Println("No extends to update")
	} else {
		fmt.Printf("✅ Updated %d extends\n", updated)
	}

	return nil
}

func runUsageSummary(args []string) error {
	logger := metrics.NewLogger(".")
	reporter := metrics.NewReporter()

	var date time.Time
	if len(args) > 0 {
		var err error
		date, err = time.Parse("2006-01-02", args[0])
		if err != nil {
			return fmt.Errorf("invalid date format: %w (use YYYY-MM-DD)", err)
		}
	} else {
		date = time.Now()
	}

	summary, err := reporter.GenerateDailySummary(logger, date)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func runUsageMetrics(args []string) error {
	days := 7
	if len(args) > 0 {
		parsed, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid days value: %w", err)
		}
		days = parsed
	}

	logger := metrics.NewLogger(".")
	reporter := metrics.NewReporter()

	m, summary, patterns, err := reporter.GenerateReport(logger, days)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	result := map[string]interface{}{
		"summary":  summary,
		"metrics":  m,
		"patterns": patterns,
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func runUsagePatterns(args []string) error {
	days := 7
	if len(args) > 0 {
		parsed, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid days value: %w", err)
		}
		days = parsed
	}

	logger := metrics.NewLogger(".")
	analyzer := metrics.NewAnalyzer()

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	logs, err := logger.Query(metrics.Filter{
		StartTime: startDate,
		EndTime:   endDate,
	})

	if err != nil {
		return fmt.Errorf("failed to query logs: %w", err)
	}

	patterns, err := analyzer.AnalyzePatterns(logs)
	if err != nil {
		return fmt.Errorf("failed to analyze patterns: %w", err)
	}

	data, err := json.MarshalIndent(patterns, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal patterns: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func runUsageBenchmark(args []string) error {
	baselineDays, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid baseline days: %w", err)
	}

	currentDays, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid current days: %w", err)
	}

	logger := metrics.NewLogger(".")
	reporter := metrics.NewReporter()

	comparison, err := reporter.GenerateBenchmark(logger, baselineDays, currentDays)
	if err != nil {
		return fmt.Errorf("failed to generate benchmark: %w", err)
	}

	data, err := json.MarshalIndent(comparison, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal benchmark: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
