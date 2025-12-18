package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// NewConfigCommand creates the config command group
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage fest configuration repositories",
		Long: `Manage fest configuration repositories.

Config repos contain custom templates, policies, plugins, and extensions
that override or extend the built-in fest methodology resources.`,
	}

	cmd.AddCommand(newConfigAddCommand())
	cmd.AddCommand(newConfigSyncCommand())
	cmd.AddCommand(newConfigUseCommand())
	cmd.AddCommand(newConfigShowCommand())
	cmd.AddCommand(newConfigListCommand())
	cmd.AddCommand(newConfigRemoveCommand())

	return cmd
}

func newConfigAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <source>",
		Short: "Add a configuration repository",
		Long: `Add a configuration repository from a git URL or local path.

For git repos, the repository will be cloned to ~/.config/fest/config-repos/<name>.
For local paths, a symlink will be created instead.`,
		Example: `  fest config add myconfig https://github.com/user/fest-config
  fest config add local /path/to/my/config
  fest config add work git@github.com:company/fest-templates.git`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigAdd(args[0], args[1])
		},
	}
	return cmd
}

func runConfigAdd(name, source string) error {
	display := ui.New(noColor, verbose)

	rm, err := config.NewRepoManager()
	if err != nil {
		return fmt.Errorf("failed to initialize repo manager: %w", err)
	}

	display.Info("Adding config repo '%s' from %s...", name, source)

	repo, err := rm.Add(name, source)
	if err != nil {
		return err
	}

	if repo.IsGitRepo {
		display.Success("Cloned git repository to %s", repo.LocalPath)
	} else {
		display.Success("Created symlink to %s", source)
	}

	display.Info("Use 'fest config use %s' to activate this config", name)
	return nil
}

func newConfigSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync [name]",
		Short: "Sync configuration repository",
		Long: `Sync a configuration repository (git pull for git repos).

If no name is provided, syncs all configured repos.`,
		Example: `  fest config sync myconfig
  fest config sync  # syncs all repos`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return runConfigSync(args[0])
			}
			return runConfigSyncAll()
		},
	}
	return cmd
}

func runConfigSync(name string) error {
	display := ui.New(noColor, verbose)

	rm, err := config.NewRepoManager()
	if err != nil {
		return fmt.Errorf("failed to initialize repo manager: %w", err)
	}

	display.Info("Syncing config repo '%s'...", name)

	if err := rm.Sync(name); err != nil {
		return err
	}

	display.Success("Synced successfully")
	return nil
}

func runConfigSyncAll() error {
	display := ui.New(noColor, verbose)

	rm, err := config.NewRepoManager()
	if err != nil {
		return fmt.Errorf("failed to initialize repo manager: %w", err)
	}

	repos := rm.List()
	if len(repos) == 0 {
		display.Info("No config repos configured")
		return nil
	}

	display.Info("Syncing %d config repo(s)...", len(repos))

	if err := rm.SyncAll(); err != nil {
		return err
	}

	display.Success("All repos synced")
	return nil
}

func newConfigUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Set active configuration repository",
		Long: `Set a configuration repository as the active one.

The active config repo is symlinked at ~/.config/fest/active and its
contents are used for templates, policies, plugins, and extensions.`,
		Example: `  fest config use myconfig
  fest config use work`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigUse(args[0])
		},
	}
	return cmd
}

func runConfigUse(name string) error {
	display := ui.New(noColor, verbose)

	rm, err := config.NewRepoManager()
	if err != nil {
		return fmt.Errorf("failed to initialize repo manager: %w", err)
	}

	if err := rm.Use(name); err != nil {
		return err
	}

	display.Success("Active config set to '%s'", name)
	display.Info("Active config path: %s", config.ActiveConfigPath())
	return nil
}

func newConfigShowCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show active configuration",
		Long:  `Show the currently active configuration repository and its details.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	return cmd
}

func runConfigShow(jsonOutput bool) error {
	display := ui.New(noColor, verbose)

	rm, err := config.NewRepoManager()
	if err != nil {
		return fmt.Errorf("failed to initialize repo manager: %w", err)
	}

	active := rm.GetActive()

	if jsonOutput {
		output := map[string]interface{}{
			"active":      rm.GetActiveName(),
			"active_path": config.ActiveConfigPath(),
		}
		if active != nil {
			output["repo"] = active
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if active == nil {
		display.Info("No active config repo")
		display.Info("Use 'fest config add' to add a config repo, then 'fest config use' to activate it")
		return nil
	}

	display.Info("Active config: %s", active.Name)
	display.Info("Source: %s", active.Source)
	display.Info("Local path: %s", active.LocalPath)
	display.Info("Type: %s", repoType(active.IsGitRepo))
	if !active.LastSync.IsZero() {
		display.Info("Last sync: %s", active.LastSync.Format(time.RFC3339))
	}

	// Show available resources
	userPath := config.ActiveUserPath()
	festPath := config.ActiveFestivalsPath()

	if userPath != "" {
		display.Info("User config: %s", userPath)
	}
	if festPath != "" {
		display.Info("Methodology overrides: %s", festPath)
	}

	return nil
}

func newConfigListCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration repositories",
		Long:  `List all configured configuration repositories.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigList(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	return cmd
}

func runConfigList(jsonOutput bool) error {
	display := ui.New(noColor, verbose)

	rm, err := config.NewRepoManager()
	if err != nil {
		return fmt.Errorf("failed to initialize repo manager: %w", err)
	}

	repos := rm.List()
	activeName := rm.GetActiveName()

	if jsonOutput {
		output := map[string]interface{}{
			"active": activeName,
			"repos":  repos,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(repos) == 0 {
		display.Info("No config repos configured")
		display.Info("Use 'fest config add <name> <source>' to add one")
		return nil
	}

	display.Info("Config repositories:")
	fmt.Println()

	for _, repo := range repos {
		marker := "  "
		if repo.Name == activeName {
			marker = "* "
		}
		fmt.Printf("%s%s\n", marker, repo.Name)
		fmt.Printf("    Source: %s\n", repo.Source)
		fmt.Printf("    Type: %s\n", repoType(repo.IsGitRepo))
		if !repo.LastSync.IsZero() {
			fmt.Printf("    Last sync: %s\n", repo.LastSync.Format(time.RFC3339))
		}
		fmt.Println()
	}

	if activeName != "" {
		display.Info("* = active")
	}

	return nil
}

func newConfigRemoveCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a configuration repository",
		Long: `Remove a configuration repository.

For git repos, this removes the cloned directory.
For local symlinks, this only removes the symlink (not the original directory).`,
		Example: `  fest config remove myconfig`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigRemove(args[0], force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation")
	return cmd
}

func runConfigRemove(name string, force bool) error {
	display := ui.New(noColor, verbose)

	rm, err := config.NewRepoManager()
	if err != nil {
		return fmt.Errorf("failed to initialize repo manager: %w", err)
	}

	repo := rm.Manifest().GetRepo(name)
	if repo == nil {
		return fmt.Errorf("config repo '%s' not found", name)
	}

	if !force {
		if repo.IsGitRepo {
			display.Warning("This will remove the cloned repository at %s", repo.LocalPath)
		}
		if !display.Confirm("Remove config repo '%s'?", name) {
			display.Info("Cancelled")
			return nil
		}
	}

	if err := rm.Remove(name); err != nil {
		return err
	}

	display.Success("Removed config repo '%s'", name)
	return nil
}

func repoType(isGit bool) string {
	if isGit {
		return "git"
	}
	return "local"
}
