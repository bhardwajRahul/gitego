// cmd/completion.go

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script for your shell",
	// Disables file completion for this command
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	//Args:                  cobra.ExactValidArgs(1),
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			if err := cmd.Root().GenBashCompletion(os.Stdout); err != nil {
				log.Fatalf("Failed to generate bash completion: %v", err)
			}
		case "zsh":
			if err := cmd.Root().GenZshCompletion(os.Stdout); err != nil {
				log.Fatalf("Failed to generate zsh completion: %v", err)
			}
		case "fish":
			if err := cmd.Root().GenFishCompletion(os.Stdout, true); err != nil {
				log.Fatalf("Failed to generate fish completion: %v", err)
			}
		case "powershell":
			if err := cmd.Root().GenPowerShellCompletion(os.Stdout); err != nil {
				log.Fatalf("Failed to generate powershell completion: %v", err)
			}
		}
	},
}

func init() {
	completionCmd.Long = fmt.Sprintf(`To load completions:

Bash:
  $ source <(%s completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ %s completion bash > /etc/bash_completion.d/%s
  # macOS:
  $ %s completion bash > /usr/local/etc/bash_completion.d/%s

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ %s completion zsh > "${fpath[1]}/_%s"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ %s completion fish | source

  # To load completions for each session, execute once:
  $ %s completion fish > ~/.config/fish/completions/%s.fish

PowerShell:
  PS> %s.exe completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> %s.exe completion powershell > %s.ps1
  # and source this file from your PowerShell profile.
`,
		binaryName, binaryName, binaryName, binaryName, binaryName,
		binaryName, binaryName, binaryName, binaryName, binaryName,
		binaryName, binaryName, binaryName,
	)
	rootCmd.AddCommand(completionCmd)
}
