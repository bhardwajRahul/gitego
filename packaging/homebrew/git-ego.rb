class GitEgo < Formula
  desc "Context-aware Git identity manager"
  homepage "https://github.com/bgreenwell/git-ego"
  url "https://github.com/bgreenwell/git-ego/releases/download/v0.2.1/git-ego-v0.2.1-macos-x86_64.tar.gz"
  sha256 "7104303773640bbfd4abd55ab2f84d06563da840ea3ceb43fd3624e495bac7e4"
  license "MIT"

  def install
    bin.install "git-ego"
  end
  test do
    assert_match "0.2.1", shell_output("#{bin}/git-ego --version")
  end
end
