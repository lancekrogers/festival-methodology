package config

import (
	"fmt"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
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
  fgo              Smart navigation (linked project ↔ festival, or festivals root)
  fgo 002          Navigate to phase 002
  fgo 2/1          Navigate to phase 2, sequence 1
  fgo active       Navigate to active directory
  fgo link         Link current festival to project (or vice versa)
  fgo --help       Show this help message

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
		return errors.Validation("unsupported shell - supported: zsh, bash, fish").WithField("shell", shell)
	}

	return nil
}

func bashZshInit() string {
	return `# fest shell integration - directory navigation
# Add to ~/.zshrc or ~/.bashrc:
#   eval "$(fest shell-init zsh)"

fgo() {
    case "$1" in
        --help|-h|help)
            # Show help for fgo/fest go
            command fest go --help
            ;;
        link)
            # Context-aware linking (no cd needed, shows TUI if needed)
            command fest go "$@"
            ;;
        map|unmap|list)
            # Pass through to fest go subcommands (no cd needed)
            command fest go "$@"
            ;;
        project)
            # Navigate to linked project
            local dest
            dest=$(command fest go project 2>&1)
            local exit_code=$?
            if [[ $exit_code -eq 0 && -n "$dest" && -d "$dest" ]]; then
                cd "$dest"
            else
                echo "fgo: no project linked (use 'fest link <path>' from a festival)" >&2
                return 1
            fi
            ;;
        fest)
            # Navigate back to festival from project
            local dest
            dest=$(command fest go fest 2>&1)
            local exit_code=$?
            if [[ $exit_code -eq 0 && -n "$dest" && -d "$dest" ]]; then
                cd "$dest"
            else
                echo "fgo: not in a linked project" >&2
                return 1
            fi
            ;;
        -*)
            # Shortcut navigation: strip leading dash and lookup
            local name="${1#-}"
            local dest
            dest=$(command fest go shortcut "$name" 2>&1)
            local exit_code=$?
            if [[ $exit_code -eq 0 && -n "$dest" && -d "$dest" ]]; then
                cd "$dest"
            else
                echo "fgo: shortcut not found: -$name" >&2
                return 1
            fi
            ;;
        *)
            # Normal navigation (festival/phase/status directories)
            local dest
            dest=$(command fest go "$@" 2>&1)
            local exit_code=$?
            if [[ $exit_code -eq 0 && -n "$dest" && -d "$dest" ]]; then
                cd "$dest"
            else
                if [[ -n "$dest" ]]; then
                    echo "$dest" >&2
                fi
                return 1
            fi
            ;;
    esac
}
`
}

func fishInit() string {
	return `# fest shell integration - directory navigation
# Add to ~/.config/fish/config.fish:
#   fest shell-init fish | source

function fgo
    switch $argv[1]
        case --help -h help
            # Show help for fgo/fest go
            command fest go --help
        case link
            # Context-aware linking (no cd needed, shows TUI if needed)
            command fest go $argv
        case map unmap list
            # Pass through to fest go subcommands (no cd needed)
            command fest go $argv
        case project
            # Navigate to linked project
            set -l dest (command fest go project 2>&1)
            set -l exit_code $status
            if test $exit_code -eq 0 -a -n "$dest" -a -d "$dest"
                cd $dest
            else
                echo "fgo: no project linked (use 'fest link <path>' from a festival)" >&2
                return 1
            end
        case fest
            # Navigate back to festival from project
            set -l dest (command fest go fest 2>&1)
            set -l exit_code $status
            if test $exit_code -eq 0 -a -n "$dest" -a -d "$dest"
                cd $dest
            else
                echo "fgo: not in a linked project" >&2
                return 1
            end
        case '-*'
            # Shortcut navigation: strip leading dash and lookup
            set -l name (string sub -s 2 $argv[1])
            set -l dest (command fest go shortcut $name 2>&1)
            set -l exit_code $status
            if test $exit_code -eq 0 -a -n "$dest" -a -d "$dest"
                cd $dest
            else
                echo "fgo: shortcut not found: -$name" >&2
                return 1
            end
        case '*'
            # Normal navigation (festival/phase/status directories)
            set -l dest (command fest go $argv 2>&1)
            set -l exit_code $status
            if test $exit_code -eq 0 -a -n "$dest" -a -d "$dest"
                cd $dest
            else
                if test -n "$dest"
                    echo $dest >&2
                end
                return 1
            end
    end
end
`
}
