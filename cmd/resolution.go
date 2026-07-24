package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bgreenwell/git-ego/config"
	"github.com/bgreenwell/git-ego/utils"
)

type profileResolution struct {
	Effective         string
	Origin            string
	Expected          string
	ExpectationSource string
	Name              string
	Email             string
	Legacy            bool
	Consistent        bool
	Problem           string
}

func resolveProfiles(cfg *config.Config) profileResolution {
	r := profileResolution{Consistent: true}
	r.Name, _ = utils.GetEffectiveGitConfig("user.name")
	r.Email, _ = utils.GetEffectiveGitConfig("user.email")
	r.Effective, r.Origin, _ = utils.GetEffectiveGitConfigWithOrigin("gitego.profile")
	r.Expected, r.ExpectationSource, r.Problem = expectedProfile(cfg)
	if r.Problem != "" {
		r.Consistent = false
		return r
	}
	if r.Effective == "" {
		for name, p := range cfg.Profiles {
			if p != nil && p.Name == r.Name && p.Email == r.Email {
				if r.Effective != "" {
					r.Effective = ""
					r.Problem = "effective identity matches multiple legacy profiles"
					r.Consistent = false
					return r
				}
				r.Effective = name
				r.Legacy = true
				r.Origin = "legacy identity fallback"
			}
		}
	}
	if r.Effective == "" {
		r.Consistent = false
		r.Problem = "no effective gitego.profile marker"
		return r
	}
	p, ok := cfg.Profiles[r.Effective]
	if !ok || p == nil {
		r.Consistent = false
		r.Problem = fmt.Sprintf("effective profile %q is unknown", r.Effective)
		return r
	}
	if p.Name != r.Name || p.Email != r.Email {
		r.Consistent = false
		r.Problem = "effective Git name/email do not match the effective profile"
		return r
	}
	if r.Expected != "" && r.Expected != r.Effective {
		r.Consistent = false
		r.Problem = fmt.Sprintf("expected profile %q but effective profile is %q", r.Expected, r.Effective)
	}
	return r
}

func expectedProfile(cfg *config.Config) (string, string, string) {
	root, err := utils.RepositoryRoot()
	if err == nil {
		assertion := filepath.Join(root, ".gitego")
		data, readErr := os.ReadFile(assertion)
		if readErr == nil {
			name := strings.TrimSpace(string(data))
			if name == "" {
				return "", assertion, "repository .gitego assertion is empty"
			}
			if _, ok := cfg.Profiles[name]; !ok {
				return name, assertion, fmt.Sprintf("repository .gitego references unknown profile %q", name)
			}
			return name, assertion, ""
		}
		if !os.IsNotExist(readErr) {
			return "", assertion, fmt.Sprintf("read repository .gitego: %v", readErr)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", "", ""
	}
	normalized, err := config.NormalizeAutoRulePath(wd)
	if err != nil {
		return "", "", ""
	}
	var best *config.AutoRule
	for _, rule := range cfg.AutoRules {
		if strings.HasPrefix(normalized, rule.Path) && (best == nil || len(rule.Path) > len(best.Path)) {
			best = rule
		}
	}
	if best != nil {
		return best.Profile, "auto-rule " + best.Path, ""
	}
	return "", "", ""
}
