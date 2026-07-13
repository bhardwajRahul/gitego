class GitEgo < Formula
  desc "Context-aware Git identity manager"
  homepage "https://github.com/bgreenwell/git-ego"
  url "https://github.com/bgreenwell/git-ego/archive/refs/tags/v0.2.0.tar.gz"
  sha256 "REPLACE_WITH_RELEASE_SHA256"
  license "MIT"

  depends_on "go" => :build
  def install
    system "go", "build", *std_go_args(ldflags: "-s -w -X github.com/bgreenwell/git-ego/cmd.version=0.2.0"), "."
  end
  test do
    assert_match "0.2.0", shell_output("#{bin}/git-ego --version")
  end
end
