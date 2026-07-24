package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/flock"
)

func TestCredentialIDIsRFC4122Version4(t *testing.T) {
	id, err := NewCredentialID()
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 36 || id[14] != '4' || !strings.Contains("89ab", strings.ToLower(string(id[19]))) {
		t.Fatalf("invalid v4 UUID %q", id)
	}
}

func TestWithLockSerializesAndTimesOut(t *testing.T) {
	temp := t.TempDir()
	oldPath, oldTimeout := gitegoConfigPath, lockTimeout
	gitegoConfigPath = filepath.Join(temp, "config.yaml")
	lockTimeout = 100 * time.Millisecond
	t.Cleanup(func() { gitegoConfigPath, lockTimeout = oldPath, oldTimeout })
	held := flock.New(filepath.Join(temp, "config.lock"))
	if err := held.Lock(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = held.Unlock() }()
	start := time.Now()
	err := WithLock(func() error { t.Fatal("entered a lock held by another owner"); return nil })
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout, got %v", err)
	}
	if time.Since(start) < 75*time.Millisecond {
		t.Fatal("lock timeout returned too early")
	}
}

func TestValidateRejectsDuplicateNormalizedRules(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{Profiles: map[string]*Profile{"a": {Name: "A", Email: "a@example.com"}, "b": {Name: "B", Email: "b@example.com"}}, AutoRules: []*AutoRule{{Path: dir, Profile: "a"}, {Path: filepath.Join(dir, "."), Profile: "b"}}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected duplicate normalized path error")
	}
}

func TestReconcileManagedIncludesAndExactLegacyCleanup(t *testing.T) {
	temp := t.TempDir()
	oldConfig, oldGit, oldProfiles := gitegoConfigPath, gitConfigPath, profilesDir
	gitegoConfigPath = filepath.Join(temp, ".gitego", "config.yaml")
	gitConfigPath = filepath.Join(temp, ".gitconfig")
	profilesDir = filepath.Join(temp, ".gitego", "profiles")
	t.Cleanup(func() { gitegoConfigPath, gitConfigPath, profilesDir = oldConfig, oldGit, oldProfiles })
	legacy := filepath.ToSlash(filepath.Join(profilesDir, "work.gitconfig"))
	unrelated := legacy + ".backup"
	initial := "[alias]\n    x = status\n[includeIf \"gitdir:/old/\"]\n    path = \"" + legacy + "\"\n[include]\n    path = \"" + unrelated + "\"\n"
	if err := os.WriteFile(gitConfigPath, []byte(initial), 0600); err != nil {
		t.Fatal(err)
	}
	broad := filepath.ToSlash(filepath.Join(temp, "src")) + "/"
	nested := broad + "nested/"
	cfg := &Config{Profiles: map[string]*Profile{"work": {Name: "Work", Email: "w@example.com"}, "client": {Name: "Client", Email: "c@example.com"}}, ActiveProfile: "work", AutoRules: []*AutoRule{{Path: nested, Profile: "client"}, {Path: broad, Profile: "work"}}}
	if err := cfg.Reconcile(); err != nil {
		t.Fatal(err)
	}
	global, _ := os.ReadFile(gitConfigPath)
	text := string(global)
	if strings.Contains(text, "gitdir:/old/") || !strings.Contains(text, unrelated) || !strings.Contains(text, "[alias]") {
		t.Fatalf("unexpected global config:\n%s", text)
	}
	if strings.Count(text, filepath.ToSlash(IncludesGitconfigPath())) != 1 {
		t.Fatalf("managed include count:\n%s", text)
	}
	includes, _ := os.ReadFile(IncludesGitconfigPath())
	content := string(includes)
	if strings.Index(content, broad) > strings.Index(content, nested) {
		t.Fatalf("rules not broad-to-specific:\n%s", content)
	}
	profile, _ := os.ReadFile(filepath.Join(profilesDir, "work.gitconfig"))
	if !strings.Contains(string(profile), "profile = \"work\"") {
		t.Fatal("profile marker missing")
	}
	before := content
	if err := cfg.Reconcile(); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(IncludesGitconfigPath())
	if before != string(after) {
		t.Fatal("reconciliation is not idempotent")
	}
}
