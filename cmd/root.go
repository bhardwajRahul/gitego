// cmd/root.go

// Package cmd provides the root command for the gitego CLI application.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// The version of the application.
var version = "0.2.4"

// binaryName is the name of the running executable, supporting invocation as
// both "gitego" and "git-ego" (git subcommand form).
var binaryName = filepath.Base(os.Args[0])

var (
	// versionFlag is a flag to print the version and exit.
	versionFlag bool
)

// rootCmd represents the base command when called without any subcommands.
// It's the main entry point for the CLI application.
var rootCmd = &cobra.Command{
	Short:         "A clever, context-aware identity manager for Git.",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		if versionFlag {
			fmt.Printf("%s version %s\n", binaryName, version)
			return nil
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.Use = binaryName
	rootCmd.Long = fmt.Sprintf(`%s is a command-line tool to seamlessly manage your Git "alter egos".

It allows you to define, switch between, and automatically apply different
user profiles (user.name, user.email), SSH keys, and Personal Access Tokens
depending on your current working directory or other contexts.`, binaryName)

	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, fmt.Sprintf("Print %s's version number", binaryName))
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}
