package config

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

var lockTimeout = 10 * time.Second

func ConfigDir() string { return filepath.Dir(gitegoConfigPath) }

func NewCredentialID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate credential id: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func WithLock(fn func() error) error {
	if err := os.MkdirAll(ConfigDir(), dirPermissions); err != nil {
		return err
	}
	l := flock.New(filepath.Join(ConfigDir(), "config.lock"))
	ctx, cancel := context.WithTimeout(context.Background(), lockTimeout)
	defer cancel()
	locked, err := l.TryLockContext(ctx, 50*time.Millisecond)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("timed out after %s waiting for configuration lock", lockTimeout)
		}
		return fmt.Errorf("acquire configuration lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("timed out after %s waiting for configuration lock", lockTimeout)
	}
	defer func() { _ = l.Unlock() }()
	return fn()
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("configuration must not be nil")
	}
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}
	for name, profile := range c.Profiles {
		if err := ValidateProfileName(name); err != nil {
			return err
		}
		if profile == nil {
			return fmt.Errorf("profile %q must not be null", name)
		}
		if strings.TrimSpace(profile.Name) == "" || strings.TrimSpace(profile.Email) == "" {
			return fmt.Errorf("profile %q requires name and email", name)
		}
		if err := ValidateCredentialHosts(profile.Hosts); err != nil {
			return fmt.Errorf("profile %q: %w", name, err)
		}
	}
	if c.ActiveProfile != "" {
		if _, ok := c.Profiles[c.ActiveProfile]; !ok {
			return fmt.Errorf("active profile %q does not exist", c.ActiveProfile)
		}
	}
	seen := map[string]string{}
	for _, rule := range c.AutoRules {
		if rule == nil {
			return fmt.Errorf("auto-rule must not be null")
		}
		if _, ok := c.Profiles[rule.Profile]; !ok {
			return fmt.Errorf("auto-rule %q references unknown profile %q", rule.Path, rule.Profile)
		}
		normalized, err := NormalizeAutoRulePath(rule.Path)
		if err != nil {
			return fmt.Errorf("normalize auto-rule %q: %w", rule.Path, err)
		}
		key := normalized
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if prior, ok := seen[key]; ok {
			return fmt.Errorf("duplicate auto-rule path %q assigned to %q and %q", normalized, prior, rule.Profile)
		}
		seen[key] = rule.Profile
		rule.Path = normalized
	}
	return nil
}

func (c *Config) SortedAutoRules() []*AutoRule {
	rules := append([]*AutoRule(nil), c.AutoRules...)
	sort.SliceStable(rules, func(i, j int) bool { return len(rules[i].Path) < len(rules[j].Path) })
	return rules
}
