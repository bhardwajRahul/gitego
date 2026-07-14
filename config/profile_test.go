package config

import "testing"

func TestProfileCredentialHosts(t *testing.T) {
	legacy := &Profile{}
	if got := legacy.CredentialHosts(); len(got) != 1 || got[0] != "github.com" {
		t.Fatalf("legacy hosts = %v, want [github.com]", got)
	}
	if !legacy.SupportsCredentialHost("GITHUB.COM") {
		t.Fatal("legacy profile should support github.com case-insensitively")
	}

	profile := &Profile{Hosts: []string{"gitlab.example.com", "github.example.com:8443"}}
	for _, host := range []string{"gitlab.example.com", "GITLAB.EXAMPLE.COM", "github.example.com:8443"} {
		if !profile.SupportsCredentialHost(host) {
			t.Fatalf("configured profile should support %q", host)
		}
	}
	if profile.SupportsCredentialHost("github.com") {
		t.Fatal("configured hosts must not fall back to github.com")
	}
}

func TestValidateCredentialHosts(t *testing.T) {
	for name, tc := range map[string]struct {
		hosts []string
		want  bool
	}{
		"host":       {hosts: []string{"github.com"}, want: true},
		"host port":  {hosts: []string{"gitlab.example.com:8443"}, want: true},
		"empty":      {hosts: []string{""}, want: false},
		"URL":        {hosts: []string{"https://github.com"}, want: false},
		"path":       {hosts: []string{"github.com/path"}, want: false},
		"credential": {hosts: []string{"user@github.com"}, want: false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := ValidateCredentialHosts(tc.hosts) == nil; got != tc.want {
				t.Fatalf("ValidateCredentialHosts(%v) success = %t, want %t", tc.hosts, got, tc.want)
			}
		})
	}
}
