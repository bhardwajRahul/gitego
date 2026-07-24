package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/bgreenwell/git-ego/config"
	"github.com/bgreenwell/git-ego/utils"
	"github.com/spf13/cobra"
)

type checkCommitRunner struct {
	getGitConfig func(string) (string, error)
	loadConfig   func() (*config.Config, error)
	resolve      func(*config.Config) profileResolution
	stdin        io.Reader
	stderr       io.Writer
	exit         func(int)
}

func (r *checkCommitRunner) run(_ *cobra.Command, _ []string) {
	cfg, err := r.loadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(r.stderr, "%s: cannot load safety configuration: %v\n", binaryName, err)
		r.exit(1)
		return
	}
	if err := cfg.Validate(); err != nil {
		_, _ = fmt.Fprintf(r.stderr, "%s: commit aborted: invalid safety configuration: %v\n", binaryName, err)
		r.exit(1)
		return
	}
	resolver := r.resolve
	if resolver == nil {
		resolver = resolveProfiles
	}
	resolution := resolver(cfg)
	if resolution.Expected == "" && resolution.ExpectationSource == "" {
		r.exit(0)
		return
	}
	if !resolution.Consistent {
		_, _ = fmt.Fprintf(r.stderr, "%s: commit aborted: %s\n", binaryName, resolution.Problem)
		r.exit(1)
		return
	}
	r.exit(0)
}

var checkCommitCmd = &cobra.Command{Use: "check-commit", Short: "Internal: enforce repository profile assertions.", Hidden: true, Run: func(cmd *cobra.Command, args []string) {
	(&checkCommitRunner{getGitConfig: utils.GetEffectiveGitConfig, loadConfig: config.Load, resolve: resolveProfiles, stdin: os.Stdin, stderr: os.Stderr, exit: os.Exit}).run(cmd, args)
}}

func init() { internalCmd.AddCommand(checkCommitCmd) }
