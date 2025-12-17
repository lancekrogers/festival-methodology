package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewShellInitCommand creates the shell-init command
func NewShellInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell-init <shell>",
		Short: "Output shell integration code for directory navigation",
		Long: `Output shell code for directory navigation integration.

This command outputs shell-specific code that creates the 'fgo' function,
which wraps 'fest go' to actually change your working directory.

SETUP (one-time):
  # For zsh, add to ~/.zshrc:
  eval "$(fest shell-init zsh)"

  # For bash, add to ~/.bashrc:
  eval "$(fest shell-init bash)"

  # For fish, add to ~/.config/fish/config.fish:
  fest shell-init fish | source

After setup, reload your shell or run: source ~/.zshrc

USAGE:
  fgo              Navigate to festivals root
  fgo 002          Navigate to phase 002
  fgo 2/1          Navigate to phase 2, sequence 1
  fgo active       Navigate to active directory

The 'fgo' function calls 'fest go' internally and uses cd to change
directories. Without shell integration, use: cd $(fest go)`,
		Example: `  # Output zsh integration code
  fest shell-init zsh

  # Add to your shell config (zsh)
  eval "$(fest shell-init zsh)"

  # After setup, navigate with:
  fgo              # Go to festivals root
  fgo 2            # Go to phase 002`,
		Args: cobra.ExactArgs(1),
		RunE: runShellInit,
	}

	return cmd
}

func runShellInit(cmd *cobra.Command, args []string) error {
	shell := args[0]

	switch shell {
	case "zsh", "bash":
		fmt.Print(bashZshInit())
	case "fish":
		fmt.Print(fishInit())
	default:
		return fmt.Errorf("unsupported shell: %s\nSupported shells: zsh, bash, fish", shell)
	}

	return nil
}

func bashZshInit() string {
	return `# fest shell integration - directory navigation
# Add to ~/.zshrc or ~/.bashrc:
#   eval "$(fest shell-init zsh)"

fgo() {
    local dest
    dest=$(command fest go "$@" 2>&1)
    local exit_code=$?

    if [[ $exit_code -eq 0 && -n "$dest" && -d "$dest" ]]; then
        cd "$dest"
    else
        # If fest go failed or returned non-directory output, show the error
        if [[ -n "$dest" ]]; then
            echo "$dest" >&2
        fi
        return 1
    fi
}
`
}

func fishInit() string {
	return `# fest shell integration - directory navigation
# Add to ~/.config/fish/config.fish:
#   fest shell-init fish | source

function fgo
    set -l dest (command fest go $argv 2>&1)
    set -l exit_code $status

    if test $exit_code -eq 0 -a -n "$dest" -a -d "$dest"
        cd $dest
    else
        # If fest go failed or returned non-directory output, show the error
        if test -n "$dest"
            echo $dest >&2
        end
        return 1
    end
end
`
}
