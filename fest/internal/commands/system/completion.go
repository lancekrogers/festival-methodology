package system

import (
	"os"

	"github.com/spf13/cobra"
)

// NewCompletionCommand creates the completion command for generating shell scripts.
func NewCompletionCommand(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for fest.

This command generates shell-specific completion scripts that enable
tab completion for commands, flags, and arguments.

SETUP:

  Bash:
    # Add to ~/.bashrc:
    source <(fest completion bash)

    # Or save to a file:
    fest completion bash > /usr/local/etc/bash_completion.d/fest

  Zsh:
    # Add to ~/.zshrc:
    source <(fest completion zsh)

    # Or save to completions directory:
    fest completion zsh > "${fpath[1]}/_fest"

  Fish:
    fest completion fish | source

    # Or save to completions directory:
    fest completion fish > ~/.config/fish/completions/fest.fish

  PowerShell:
    fest completion powershell | Out-String | Invoke-Expression

CUSTOM COMPLETIONS:

After setup, you can tab-complete:
  fest status act<TAB>     → fest status active/
  fest show pla<TAB>       → fest show plan
  fest create <TAB>        → festival, phase, sequence, task`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return cmd
}
