package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bgreenwell/git-ego/config"
	"github.com/spf13/cobra"
)

func TestAutoRejectsConflictingNormalizedPathUnlessReplace(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{Profiles: map[string]*config.Profile{
		"work":     {Name: "Work", Email: "work@example.com"},
		"personal": {Name: "Personal", Email: "personal@example.com"},
	}}
	normalized, err := config.NormalizeAutoRulePath(dir)
	if err != nil {
		t.Fatal(err)
	}
	cfg.AutoRules = []*config.AutoRule{{Path: normalized, Profile: "work"}}
	runner := &autoRunner{load: func() (*config.Config, error) { return cfg, nil }, save: func(*config.Config) error { return nil }}
	autoReplace = false
	if err := runner.run(&cobra.Command{}, []string{filepath.Join(dir, "."), "personal"}); err == nil || !strings.Contains(err.Error(), "--replace") {
		t.Fatalf("expected conflict error, got %v", err)
	}
	autoReplace = true
	t.Cleanup(func() { autoReplace = false })
	if err := runner.run(&cobra.Command{}, []string{dir, "personal"}); err != nil {
		t.Fatal(err)
	}
	if cfg.AutoRules[0].Profile != "personal" {
		t.Fatal("rule was not replaced")
	}
}

func TestCredentialRefusesExpectationMismatchAndDotFileSelection(t *testing.T) {
	cfg := &config.Config{Profiles: map[string]*config.Profile{"work": {Name: "Work", Email: "w@example.com", Username: "worker", CredentialID: "id-work"}}}
	called := false
	runner := &credentialRunner{loadConfig: func() (*config.Config, error) { return cfg, nil }, getToken: func(string) (string, error) { called = true; return "secret", nil }, stdin: strings.NewReader("protocol=https\nhost=github.com\n\n"), stdout: &bytes.Buffer{}, resolve: func(*config.Config) profileResolution {
		return profileResolution{Effective: "", Expected: "work", ExpectationSource: ".gitego", Consistent: false, Problem: "no marker"}
	}}
	runner.run(&cobra.Command{}, []string{"get"})
	if called {
		t.Fatal(".gitego selected a credential without an effective marker")
	}
}

func TestUseLocalWritesIdentityAndMarker(t *testing.T) {
	cfg := &config.Config{Profiles: map[string]*config.Profile{"work": {Name: "Work", Email: "w@example.com"}}}
	writes := map[string]string{}
	runner := &useRunner{load: func() (*config.Config, error) { return cfg, nil }, setLocalGit: func(k, v string) error { writes[k] = v; return nil }, unsetLocalGit: func(string) error { return nil }}
	useLocalFlag = true
	t.Cleanup(func() { useLocalFlag = false })
	if err := runner.run(&cobra.Command{}, []string{"work"}); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{"user.name": "Work", "user.email": "w@example.com", "gitego.profile": "work"} {
		if writes[key] != want {
			t.Fatalf("%s=%q, want %q", key, writes[key], want)
		}
	}
}

func TestResolutionUsesRepoRootAssertionAndMostSpecificRule(t *testing.T) {
	repo := t.TempDir()
	nested := filepath.Join(repo, "clients", "project")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init"}, {"config", "user.name", "Client"}, {"config", "user.email", "client@example.com"}, {"config", "gitego.profile", "client"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	broad, _ := config.NormalizeAutoRulePath(repo)
	specific, _ := config.NormalizeAutoRulePath(filepath.Join(repo, "clients"))
	cfg := &config.Config{Profiles: map[string]*config.Profile{"work": {Name: "Work", Email: "work@example.com"}, "client": {Name: "Client", Email: "client@example.com"}}, AutoRules: []*config.AutoRule{{Path: broad, Profile: "work"}, {Path: specific, Profile: "client"}}}
	old, _ := os.Getwd()
	if err := os.Chdir(nested); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	r := resolveProfiles(cfg)
	if !r.Consistent || r.Expected != "client" {
		t.Fatalf("nested resolution: %+v", r)
	}
	if err := os.WriteFile(filepath.Join(repo, ".gitego"), []byte("work\n"), 0600); err != nil {
		t.Fatal(err)
	}
	r = resolveProfiles(cfg)
	if r.Consistent || r.Expected != "work" {
		t.Fatalf("root assertion should detect mismatch: %+v", r)
	}
}
