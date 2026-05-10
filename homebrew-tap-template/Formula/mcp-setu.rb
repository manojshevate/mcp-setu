class McpSetu < Formula
  desc "MCP bridge for Ollama"
  homepage "https://github.com/manojshevate/mcp-setu"

  on_macos do
    on_arm do
      url "https://github.com/manojshevate/mcp-setu/releases/download/v0.1.0/mcp-setu_v0.1.0_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/manojshevate/mcp-setu/releases/download/v0.1.0/mcp-setu_v0.1.0_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_AMD64_SHA256"
    end
  end

  version "0.1.0"
  license "MIT"

  def install
    bin.install "mcp-setu"
  end

  def test
    assert_match "mcp-setu version", shell_output("#{bin}/mcp-setu version")
  end
end
