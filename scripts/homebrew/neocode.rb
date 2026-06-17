class Neocode < Formula
  desc "AI coding agent for Chinese and international models"
  homepage "https://neocode.dev"
  version "2.0.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/user/neocode/releases/download/v#{version}/neocode-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64"
    end
    on_intel do
      url "https://github.com/user/neocode/releases/download/v#{version}/neocode-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER_AMD64"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/user/neocode/releases/download/v#{version}/neocode-linux-arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64"
    end
    on_intel do
      url "https://github.com/user/neocode/releases/download/v#{version}/neocode-linux-amd64.tar.gz"
      sha256 "PLACEHOLDER_AMD64"
    end
  end

  def install
    bin.install "neocode"
  end

  test do
    assert_match "neocode", shell_output("#{bin}/neocode version")
  end
end
