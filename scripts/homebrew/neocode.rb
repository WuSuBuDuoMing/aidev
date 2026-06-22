class Neocode < Formula
  desc "AI coding agent for Chinese and international models"
  homepage "https://neocode.dev"
  version "2.9.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/WuSuBuDuoMing/aidev/releases/download/v#{version}/neocode_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64"
    end
    on_intel do
      url "https://github.com/WuSuBuDuoMing/aidev/releases/download/v#{version}/neocode_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_AMD64"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/WuSuBuDuoMing/aidev/releases/download/v#{version}/neocode_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64"
    end
    on_intel do
      url "https://github.com/WuSuBuDuoMing/aidev/releases/download/v#{version}/neocode_linux_amd64.tar.gz"
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
