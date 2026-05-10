class Mcpgo < Formula
  desc "MCP bridge for Ollama"
  homepage "https://github.com/manojshevate/mcpgo"

  on_macos do
    on_arm do
      url "https://github.com/manojshevate/mcpgo/releases/download/v0.1.0/mcpgo_v0.1.0_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/manojshevate/mcpgo/releases/download/v0.1.0/mcpgo_v0.1.0_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_AMD64_SHA256"
    end
  end

  version "0.1.0"
  license "MIT"

  def install
    bin.install "mcpgo"
  end

  def test
    assert_match "mcpgo version", shell_output("#{bin}/mcpgo version")
  end
end
