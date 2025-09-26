# Homebrew formula for fest CLI
class Fest < Formula
  desc "Festival methodology CLI for managing software development festivals"
  homepage "https://github.com/anthropics/guild-framework"
  version "0.1.0"

  # URL to your release archive
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/anthropics/guild-framework/releases/download/fest-v0.1.0/fest-darwin-arm64.tar.gz"
    sha256 "YOUR_SHA256_HERE"
  elsif OS.mac?
    url "https://github.com/anthropics/guild-framework/releases/download/fest-v0.1.0/fest-darwin-amd64.tar.gz"
    sha256 "YOUR_SHA256_HERE"
  elsif OS.linux?
    url "https://github.com/anthropics/guild-framework/releases/download/fest-v0.1.0/fest-linux-amd64.tar.gz"
    sha256 "YOUR_SHA256_HERE"
  end

  def install
    bin.install "fest"
  end

  test do
    system "#{bin}/fest", "--version"
  end
end