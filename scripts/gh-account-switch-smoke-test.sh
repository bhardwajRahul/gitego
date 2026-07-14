#!/usr/bin/env bash
# Verify switching from a regular profile to a GitHub PAT profile and back.
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: scripts/gh-account-switch-smoke-test.sh --push

Required environment variables:
  GH_GIT_EGO                     PAT for the temporary GitHub test account
  GIT_EGO_SMOKE_RETURN_PROFILE   regular git-ego profile to restore afterwards

Optional environment variables:
  GIT_EGO_BIN                    git-ego executable (default: git-ego)
  GIT_EGO_SMOKE_TEST_PROFILE     test profile (default: gh-smoke)
  GIT_EGO_SMOKE_REPO             test repository HTTPS URL
                                 (default: https://github.com/w108bmg/git-ego-smoke-test.git)
  GIT_EGO_SMOKE_USERNAME         expected GitHub username (default: w108bmg)
  GIT_EGO_SMOKE_NAME             expected Git author name (default: Brandon Greenwell)
  GIT_EGO_SMOKE_EMAIL            expected Git email
                                 (default: greenwell.brandon+github@gmail.com)

The script temporarily activates the test profile by directory rule, appends a
timestamped entry to GIT-EGO-SMOKE-CHANGELOG.md, and pushes it with the test
account's PAT. It always removes the temporary rule and restores the regular
profile before exiting.
EOF
}

if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
  usage
  exit 0
fi

if [ "${1:-}" != "--push" ] || [ "$#" -ne 1 ]; then
  usage >&2
  exit 2
fi

: "${GH_GIT_EGO:?set GH_GIT_EGO to the GitHub PAT before running this script}"
: "${GIT_EGO_SMOKE_RETURN_PROFILE:?set GIT_EGO_SMOKE_RETURN_PROFILE to the profile to restore}"

git_ego="${GIT_EGO_BIN:-git-ego}"
test_profile="${GIT_EGO_SMOKE_TEST_PROFILE:-gh-smoke}"
return_profile="$GIT_EGO_SMOKE_RETURN_PROFILE"
repo_url="${GIT_EGO_SMOKE_REPO:-https://github.com/w108bmg/git-ego-smoke-test.git}"
username="${GIT_EGO_SMOKE_USERNAME:-w108bmg}"
name="${GIT_EGO_SMOKE_NAME:-Brandon Greenwell}"
email="${GIT_EGO_SMOKE_EMAIL:-greenwell.brandon+github@gmail.com}"
test_dir="$(mktemp -d "${TMPDIR:-/tmp}/git-ego-account-switch.XXXXXX")"

command -v "$git_ego" >/dev/null || {
  echo "git-ego executable not found: $git_ego" >&2
  exit 1
}

cleanup() {
  "$git_ego" auto rm "$test_dir" >/dev/null 2>&1 || true
  "$git_ego" use "$return_profile" >/dev/null ||
    echo "warning: could not restore profile '$return_profile'" >&2
}
trap cleanup EXIT

printf '%s' "$GH_GIT_EGO" | "$git_ego" pat set "$test_profile"
"$git_ego" auto "$test_dir" "$test_profile"

cd "$test_dir"

if ! printf 'protocol=https\nhost=github.com\n\n' |
  "$git_ego" credential get |
  awk -v username="$username" '$0 == "username=" username { found_username = 1 } /^password=.+/ { found_password = 1 } END { exit !(found_username && found_password) }'; then
  echo "git-ego did not return the expected GitHub test-account credential" >&2
  exit 1
fi

if printf 'protocol=https\nhost=example.com\n\n' | "$git_ego" credential get | grep -q .; then
  echo "git-ego returned credentials for an unconfigured host" >&2
  exit 1
fi

git -c credential.helper= -c credential.helper="!$git_ego credential" clone "$repo_url"
repo_dir="${repo_url##*/}"
cd "${repo_dir%.git}"

test "$(git config user.name)" = "$name" || {
  echo "unexpected effective Git name: $(git config user.name)" >&2
  exit 1
}

test "$(git config user.email)" = "$email" || {
  echo "unexpected effective Git email: $(git config user.email)" >&2
  exit 1
}

printf '%s | profile=%s | user=%s | email=%s\n' \
  "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$test_profile" "$username" "$email" \
  >> GIT-EGO-SMOKE-CHANGELOG.md
git add GIT-EGO-SMOKE-CHANGELOG.md
git commit -m "Test git-ego account switch"

GIT_TERMINAL_PROMPT=0 git -c credential.helper= -c credential.helper="!$git_ego credential" push

echo "Test-account commit pushed; restoring '$return_profile'."
