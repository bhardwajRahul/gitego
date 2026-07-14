// config/config.go

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"

	"al.essio.dev/pkg/shellescape"
	"gopkg.in/yaml.v3"
)

// Profile represents a single user profile with a name and email.
type Profile struct {
	Name       string   `yaml:"name"`
	Email      string   `yaml:"email"`
	Username   string   `yaml:"username,omitempty"`
	SSHKey     string   `yaml:"ssh_key,omitempty"`
	SigningKey string   `yaml:"signing_key,omitempty"`
	Hosts      []string `yaml:"hosts,omitempty"`
	PAT        string   `yaml:"-"`
}

// CredentialHosts returns the hosts a profile may authenticate to. Hostless
// legacy profiles retain their historical GitHub-only behavior.
func (p *Profile) CredentialHosts() []string {
	if len(p.Hosts) == 0 {
		return []string{"github.com"}
	}
	return p.Hosts
}

// SupportsCredentialHost reports whether a profile is explicitly scoped to a host.
func (p *Profile) SupportsCredentialHost(host string) bool {
	for _, configured := range p.CredentialHosts() {
		if strings.EqualFold(strings.TrimSpace(configured), host) {
			return true
		}
	}
	return false
}

// ValidateCredentialHosts accepts DNS names (optionally with a port) and
// rejects URL/path forms that would never match Git's credential protocol.
func ValidateCredentialHosts(hosts []string) error {
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if host == "" || strings.ContainsAny(host, "/\\@ \t\r\n") {
			return fmt.Errorf("invalid credential host %q", host)
		}
	}
	return nil
}

// AutoRule represents a single directory-to-profile mapping.
type AutoRule struct {
	Path    string `yaml:"path"`
	Profile string `yaml:"profile"`
}

// Config represents the entire structure of our config file.
type Config struct {
	Profiles      map[string]*Profile `yaml:"profiles"`
	AutoRules     []*AutoRule         `yaml:"auto_rules,omitempty"`
	ActiveProfile string              `yaml:"active_profile,omitempty"`
}

const (
	// dirPermissions are the default permissions for directories created by gitego.
	dirPermissions = 0755
	// filePermissions are the default permissions for files created by gitego.
	filePermissions = 0600
)

var (
	gitegoConfigPath string
	gitConfigPath    string
	profilesDir      string
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("could not get user home directory: %v", err))
	}

	gitegoConfigPath = filepath.Join(home, ".gitego", "config.yaml")
	profilesDir = filepath.Join(home, ".gitego", "profiles")
	gitConfigPath = filepath.Join(home, ".gitconfig")
}

// Load reads and decodes the gitego config.yaml file and validates it.
func Load() (*Config, error) {
	return load(true)
}

// LoadQuiet reads configuration without emitting consistency warnings. It is
// intended for Git's credential-helper protocol, where unsolicited stderr
// output is treated as a transport error by some clients.
func LoadQuiet() (*Config, error) {
	return load(false)
}

func load(validate bool) (*Config, error) {
	cfg := &Config{
		Profiles: make(map[string]*Profile),
	}

	data, err := os.ReadFile(gitegoConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}

		return nil, fmt.Errorf("could not read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("could not parse config file: %w", err)
	}

	if validate {
		validateConfig(cfg)
	}

	return cfg, nil
}

func validateConfig(cfg *Config) {
	if cfg.ActiveProfile != "" {
		if _, exists := cfg.Profiles[cfg.ActiveProfile]; !exists {
			fmt.Fprintf(os.Stderr, "Warning: Active profile '%s' not found. It may have been deleted.\n", cfg.ActiveProfile)
		}
	}

	for _, rule := range cfg.AutoRules {
		if _, exists := cfg.Profiles[rule.Profile]; !exists {
			fmt.Fprintf(os.Stderr,
				"Warning: Auto-switch rule for path '%s' points to a non-existent profile '%s'.\n",
				rule.Path, rule.Profile)
		}
	}
}

func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("could not serialize config to yaml: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(gitegoConfigPath), dirPermissions); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	if err := os.WriteFile(gitegoConfigPath, data, filePermissions); err != nil {
		return fmt.Errorf("could not write config file: %w", err)
	}

	return nil
}

func (c *Config) GetActiveProfileForCurrentDir() (profileName, source string) {
	profileName = c.ActiveProfile
	source = getDefaultSource(c.ActiveProfile)

	if len(c.AutoRules) == 0 {
		return profileName, source
	}

	currentAbsDir, err := getCurrentAbsDir()
	if err != nil {
		return profileName, source
	}
	if localProfile, found := c.profileFromRepositoryFile(currentAbsDir); found {
		if _, exists := c.Profiles[localProfile]; exists {
			return localProfile, "repository .gitego profile"
		}
		return "", "repository .gitego references an unknown profile"
	}

	bestMatch := c.findBestMatchingRule(currentAbsDir)
	if bestMatch != nil {
		profileName = bestMatch.Profile
		source = fmt.Sprintf("gitego auto-rule for profile '%s'", bestMatch.Profile)
	}

	return profileName, source
}

func (c *Config) profileFromRepositoryFile(currentAbsDir string) (string, bool) {
	dir := filepath.FromSlash(strings.TrimSuffix(currentAbsDir, "/"))
	for {
		path := filepath.Join(dir, ".gitego")
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return "", true
			}
			name := strings.TrimSpace(string(data))
			return name, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

func getDefaultSource(activeProfile string) string {
	if activeProfile == "" {
		return "No active gitego profile"
	}

	return "Global gitego default"
}

func getCurrentAbsDir() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	evalDir, err := filepath.EvalSymlinks(currentDir)
	if err != nil {
		evalDir = currentDir
	}

	currentAbsDir, err := filepath.Abs(evalDir)
	if err != nil {
		return "", err
	}

	currentAbsDir = filepath.ToSlash(currentAbsDir)
	if !strings.HasSuffix(currentAbsDir, "/") {
		currentAbsDir += "/"
	}

	return currentAbsDir, nil
}

func (c *Config) findBestMatchingRule(currentAbsDir string) *AutoRule {
	var bestMatch *AutoRule

	bestMatchPath := ""

	for _, rule := range c.AutoRules {
		ruleAbsPath, err := NormalizeAutoRulePath(rule.Path)
		if err != nil {
			continue
		}

		if c.isPathMatch(currentAbsDir, ruleAbsPath) && len(ruleAbsPath) > len(bestMatchPath) {
			bestMatchPath = ruleAbsPath
			bestMatch = rule
		}
	}

	return bestMatch
}

func (c *Config) isPathMatch(currentAbsDir, ruleAbsPath string) bool {
	compareDir := currentAbsDir
	compareRulePath := ruleAbsPath

	if runtime.GOOS == "windows" {
		compareDir = strings.ToLower(compareDir)
		compareRulePath = strings.ToLower(compareRulePath)
	}

	return strings.HasPrefix(compareDir, compareRulePath)
}

// NormalizeAutoRulePath returns an absolute, symlink-aware, slash-normalized
// directory path suitable for both storage and matching.
func NormalizeAutoRulePath(path string) (string, error) {
	ruleEvalPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		ruleEvalPath = path
	}

	ruleAbsPath, err := filepath.Abs(ruleEvalPath)
	if err != nil {
		return "", err
	}

	ruleAbsPath = filepath.ToSlash(ruleAbsPath)

	if !strings.HasSuffix(ruleAbsPath, "/") {
		ruleAbsPath += "/"
	}

	return ruleAbsPath, nil
}

func EnsureProfileGitconfig(profileName string, profile *Profile) error {
	if err := ValidateProfileName(profileName); err != nil {
		return err
	}
	if err := os.MkdirAll(profilesDir, dirPermissions); err != nil {
		return fmt.Errorf("could not create profiles directory: %w", err)
	}

	name, err := quoteGitConfigValue(profile.Name)
	if err != nil {
		return fmt.Errorf("invalid profile name value: %w", err)
	}
	email, err := quoteGitConfigValue(profile.Email)
	if err != nil {
		return fmt.Errorf("invalid profile email value: %w", err)
	}

	content := fmt.Sprintf("[user]\n    name = %s\n    email = %s\n", name, email)

	if profile.SigningKey != "" {
		signingKey, err := quoteGitConfigValue(profile.SigningKey)
		if err != nil {
			return fmt.Errorf("invalid signing key: %w", err)
		}
		content += fmt.Sprintf("    signingkey = %s\n", signingKey)
		if IsSSHSigningKey(profile.SigningKey) {
			content += "\n[gpg]\n    format = ssh\n"
		}
	}

	if profile.SSHKey != "" {
		sshCommand, err := quoteGitConfigValue(SSHCommand(profile.SSHKey))
		if err != nil {
			return fmt.Errorf("invalid SSH key: %w", err)
		}
		coreBlock := fmt.Sprintf("\n[core]\n    sshCommand = %s\n", sshCommand)
		content += coreBlock
	}

	filePath, err := ProfileGitconfigPath(profileName)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(content), filePermissions)
}

func AddIncludeIf(profileName string, dirPath string) error {
	if err := ValidateProfileName(profileName); err != nil {
		return err
	}
	if err := validateGitConfigValue(dirPath); err != nil {
		return fmt.Errorf("invalid auto-rule path: %w", err)
	}
	profileConfigPath, err := ProfileGitconfigPath(profileName)
	if err != nil {
		return err
	}
	profileConfigPath = filepath.ToSlash(profileConfigPath)
	condition, _ := quoteGitConfigValue("gitdir:" + dirPath)
	quotedPath, _ := quoteGitConfigValue(profileConfigPath)
	includeLine := fmt.Sprintf("[includeIf %s]\n    path = %s", condition, quotedPath)

	displayConfigPath := fmt.Sprintf("~/.gitego/profiles/%s.gitconfig", profileName)
	displayLine := fmt.Sprintf("[includeIf \"gitdir:%s\"]\n    path = %s", dirPath, displayConfigPath)

	input, err := os.ReadFile(gitConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not read global .gitconfig: %w", err)
	}
	if hasIncludeIfRule(strings.Split(string(input), "\n"), condition, profileConfigPath) {
		fmt.Printf("✓ Auto-switch rule for profile '%s' on path '%s' already exists.\n", profileName, dirPath)

		return nil
	}

	f, err := os.OpenFile(gitConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		return fmt.Errorf("could not open .gitconfig for writing: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("Warning: Failed to close gitconfig file: %v\n", err)
		}
	}()

	if _, err := f.WriteString("\n# gitego auto-switch rule\n" + includeLine + "\n"); err != nil {
		return fmt.Errorf("could not write to .gitconfig: %w", err)
	}

	fmt.Printf("✓ Added auto-switch rule to ~/.gitconfig:\n%s\n", displayLine)

	return nil
}

// RemoveIncludeIf finds and removes the includeIf directive associated with a profile.
func RemoveIncludeIf(profileName string) error {
	profileConfigPath, err := ProfileGitconfigPath(profileName)
	if err != nil {
		return err
	}
	profileConfigPath = filepath.ToSlash(profileConfigPath)

	input, err := os.ReadFile(gitConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	lines := strings.Split(string(input), "\n")
	newLines := removeGitegoRules(lines, profileConfigPath)
	output := formatOutput(newLines)

	return os.WriteFile(gitConfigPath, []byte(output), filePermissions)
}

// RemoveIncludeIfAt removes one directory-specific includeIf rule while
// retaining other rules that use the same profile.
func RemoveIncludeIfAt(profileName, dirPath string) error {
	profileConfigPath, err := ProfileGitconfigPath(profileName)
	if err != nil {
		return err
	}
	condition, err := quoteGitConfigValue("gitdir:" + dirPath)
	if err != nil {
		return err
	}
	input, err := os.ReadFile(gitConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	lines := strings.Split(string(input), "\n")
	var kept []string
	for i := 0; i < len(lines); {
		if strings.TrimSpace(lines[i]) == fmt.Sprintf("[includeIf %s]", condition) && isGitegoRule(lines, i, filepath.ToSlash(profileConfigPath)) {
			kept = removeCommentIfPresent(kept, lines, i)
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "[") {
				i++
			}
			continue
		}
		kept = append(kept, lines[i])
		i++
	}
	return os.WriteFile(gitConfigPath, []byte(formatOutput(kept)), filePermissions)
}

func removeGitegoRules(lines []string, profileConfigPath string) []string {
	var newLines []string

	for i := 0; i < len(lines); {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "[includeIf") && isGitegoRule(lines, i, profileConfigPath) {
			newLines = removeCommentIfPresent(newLines, lines, i)
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "[") {
				i++
			}
			continue
		}

		newLines = append(newLines, line)
		i++
	}

	return newLines
}

func removeCommentIfPresent(newLines []string, lines []string, i int) []string {
	if i > 0 && strings.TrimSpace(lines[i-1]) == "# gitego auto-switch rule" && len(newLines) > 0 {
		return newLines[:len(newLines)-1]
	}

	return newLines
}

func formatOutput(lines []string) string {
	output := strings.Join(lines, "\n")
	output = strings.TrimSpace(output)

	if output != "" {
		output += "\n"
	}

	return output
}

func isGitegoRule(lines []string, index int, profileConfigPath string) bool {
	for j := index + 1; j < len(lines); j++ {
		nextLineTrimmed := strings.TrimSpace(lines[j])
		if strings.HasPrefix(nextLineTrimmed, "[") {
			return false
		}

		if strings.HasPrefix(nextLineTrimmed, "path") {
			return strings.Contains(filepath.ToSlash(nextLineTrimmed), profileConfigPath)
		}
	}

	return false
}

// ProfileGitconfigPath returns the only permitted location for a generated profile config.
func ProfileGitconfigPath(profileName string) (string, error) {
	if err := ValidateProfileName(profileName); err != nil {
		return "", err
	}

	return filepath.Join(profilesDir, profileName+".gitconfig"), nil
}

// ValidateProfileName rejects path components and control characters before a profile
// name is used in a file path or Git config.
func ValidateProfileName(profileName string) error {
	if profileName == "" {
		return fmt.Errorf("profile name must not be empty")
	}
	for _, r := range profileName {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' || r == '-') {
			return fmt.Errorf("invalid profile name %q: use only letters, numbers, '.', '_' and '-'", profileName)
		}
	}
	if profileName == "." || profileName == ".." {
		return fmt.Errorf("invalid profile name %q", profileName)
	}

	return nil
}

// IsSSHSigningKey identifies the documented SSH-key-path form of signing keys.
func IsSSHSigningKey(key string) bool {
	return filepath.IsAbs(key) || strings.HasPrefix(key, "./") || strings.HasPrefix(key, "../") || strings.HasPrefix(key, "~/") || strings.Contains(key, "\\")
}

// SSHCommand returns a shell-safe SSH invocation for Git's core.sshCommand.
func SSHCommand(keyPath string) string {
	return "ssh -i " + shellescape.Quote(keyPath)
}

func quoteGitConfigValue(value string) (string, error) {
	if err := validateGitConfigValue(value); err != nil {
		return "", err
	}

	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return "\"" + value + "\"", nil
}

func validateGitConfigValue(value string) error {
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Errorf("control characters are not allowed")
		}
	}
	return nil
}

func hasIncludeIfRule(lines []string, condition, profileConfigPath string) bool {
	for i, line := range lines {
		if strings.TrimSpace(line) != fmt.Sprintf("[includeIf %s]", condition) {
			continue
		}
		for j := i + 1; j < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[j]), "["); j++ {
			parts := strings.SplitN(strings.TrimSpace(lines[j]), "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == "path" {
				value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
				return filepath.ToSlash(value) == profileConfigPath
			}
		}
	}

	return false
}

// VerifyAutoRules reports drift between config.yaml, generated profile
// includes, and ~/.gitconfig. It does not modify user configuration.
func (c *Config) VerifyAutoRules() []error {
	var problems []error
	if c.ActiveProfile != "" {
		if _, ok := c.Profiles[c.ActiveProfile]; !ok {
			problems = append(problems, fmt.Errorf("active profile %q does not exist", c.ActiveProfile))
		}
	}
	input, err := os.ReadFile(gitConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return []error{fmt.Errorf("read global gitconfig: %w", err)}
	}
	lines := strings.Split(string(input), "\n")
	for _, rule := range c.AutoRules {
		profile, ok := c.Profiles[rule.Profile]
		if !ok || profile == nil {
			problems = append(problems, fmt.Errorf("%s: profile %q does not exist", rule.Path, rule.Profile))
			continue
		}
		profilePath, err := ProfileGitconfigPath(rule.Profile)
		if err != nil {
			problems = append(problems, err)
			continue
		}
		if _, err := os.Stat(profilePath); err != nil {
			problems = append(problems, fmt.Errorf("%s: profile include %q is missing", rule.Path, profilePath))
		}
		condition, _ := quoteGitConfigValue("gitdir:" + rule.Path)
		if !hasIncludeIfRule(lines, condition, filepath.ToSlash(profilePath)) {
			problems = append(problems, fmt.Errorf("%s: global includeIf rule for profile %q is missing", rule.Path, rule.Profile))
		}
	}
	return problems
}
