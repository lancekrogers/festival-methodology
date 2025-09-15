# Distribution Guide for fest CLI

This guide explains how to build, sign, and distribute the fest CLI tool across different platforms.

## Code Signing

### Why Code Signing?

Modern operating systems (especially macOS) require binaries to be signed to avoid security warnings. Without signing:
- macOS Gatekeeper will block the binary by default
- Users will see "unidentified developer" warnings
- Windows SmartScreen may flag the binary

### macOS Signing

#### Ad-hoc Signing (Development)

For local development and testing:

```bash
# Build and sign with ad-hoc signature
make sign
```

This creates a locally-signed binary that:
- Works on your machine without warnings
- Can be shared with technical users who know how to bypass Gatekeeper
- Does NOT work seamlessly for general distribution

#### Developer ID Signing (Production)

For wider distribution, you need an Apple Developer ID certificate:

1. Enroll in Apple Developer Program ($99/year)
2. Create a Developer ID Application certificate
3. Sign with your certificate:

```bash
# Set your signing identity
export CODESIGN_IDENTITY="Developer ID Application: Your Name (TEAMID)"

# Build and sign
make sign-release
```

#### Notarization (Recommended)

For the best user experience on macOS:

1. Sign with Developer ID certificate
2. Submit to Apple for notarization
3. Staple the notarization ticket

```bash
# After signing, create a ZIP for notarization
zip -r fest.zip fest

# Submit for notarization
xcrun notarytool submit fest.zip \
  --apple-id "your-apple-id@example.com" \
  --team-id "YOURTEAMID" \
  --password "app-specific-password" \
  --wait

# Staple the ticket
xcrun stapler staple fest
```

### Linux Distribution

Linux doesn't require signing, but you can:
- Sign with GPG for authenticity
- Provide checksums for verification
- Package as .deb/.rpm for package managers

```bash
# Build Linux binary
make build-linux

# Create checksum
sha256sum fest-linux-amd64 > fest-linux-amd64.sha256
```

### Windows Distribution

For Windows distribution:

1. Get a code signing certificate from a CA
2. Use signtool.exe to sign:

```batch
signtool sign /a /t http://timestamp.digicert.com fest-windows-amd64.exe
```

## Building Release Packages

### Local Release Build

Build all platform releases locally:

```bash
# Build and package all platforms
make release

# This creates:
# - dist/fest-darwin-amd64.tar.gz (Intel Mac)
# - dist/fest-darwin-arm64.tar.gz (Apple Silicon)
# - dist/fest-linux-amd64.tar.gz
# - dist/fest-windows-amd64.zip
```

### GitHub Actions Release

The project includes a GitHub Actions workflow for automated releases:

1. Push a version tag:
```bash
git tag v1.0.0
git push origin v1.0.0
```

2. The workflow will:
   - Build binaries for all platforms
   - Ad-hoc sign macOS binaries
   - Create archives
   - Generate checksums
   - Create a GitHub release with all artifacts

## Distribution Channels

### Direct Download

Provide signed binaries on:
- GitHub Releases (automatic with workflow)
- Your website
- CDN

Include:
- Clear installation instructions
- SHA256 checksums
- GPG signatures (optional)

### Package Managers

#### Homebrew (macOS/Linux)

Create a formula:

```ruby
class Fest < Formula
  desc "Festival Methodology CLI tool"
  homepage "https://github.com/festival-methodology/fest"
  version "1.0.0"
  
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/festival-methodology/fest/releases/download/v1.0.0/fest-darwin-arm64.tar.gz"
    sha256 "ACTUAL_SHA256_HERE"
  elsif OS.mac?
    url "https://github.com/festival-methodology/fest/releases/download/v1.0.0/fest-darwin-amd64.tar.gz"
    sha256 "ACTUAL_SHA256_HERE"
  else
    url "https://github.com/festival-methodology/fest/releases/download/v1.0.0/fest-linux-amd64.tar.gz"
    sha256 "ACTUAL_SHA256_HERE"
  end

  def install
    bin.install "fest"
  end
end
```

#### Go Install

For Go developers:

```bash
go install github.com/festival-methodology/fest/cmd/fest@latest
```

## Security Best Practices

1. **Always sign binaries** for production distribution
2. **Provide checksums** for all downloads
3. **Use HTTPS** for all download links
4. **Document verification steps** clearly
5. **Keep signing certificates secure**
6. **Use timestamp servers** when signing
7. **Test on clean systems** before release

## Troubleshooting

### macOS Gatekeeper Issues

If users see "cannot be opened because the developer cannot be verified":

1. Right-click the binary and select "Open"
2. Or remove quarantine attribute:
```bash
xattr -d com.apple.quarantine fest
```

### Windows SmartScreen

If Windows blocks the executable:

1. Click "More info"
2. Click "Run anyway"

### Linux Permission Issues

Ensure executable permission:
```bash
chmod +x fest
```

## Release Checklist

- [ ] Update version in code
- [ ] Update CHANGELOG.md
- [ ] Run all tests: `make test`
- [ ] Build all platforms: `make build-all`
- [ ] Sign macOS binaries: `make sign-release`
- [ ] Create archives: `make release`
- [ ] Generate checksums
- [ ] Test on clean systems
- [ ] Create git tag
- [ ] Push tag to trigger release workflow
- [ ] Verify GitHub release artifacts
- [ ] Update documentation
- [ ] Announce release

## Support Matrix

| Platform | Architecture | Signing | Distribution |
|----------|-------------|---------|--------------|
| macOS | Intel (amd64) | ✅ Ad-hoc / Developer ID | Direct, Homebrew |
| macOS | Apple Silicon (arm64) | ✅ Ad-hoc / Developer ID | Direct, Homebrew |
| Linux | amd64 | Optional (GPG) | Direct, Package managers |
| Windows | amd64 | Optional (Certificate) | Direct, Chocolatey |

## Version Management

The version is set during build:

```bash
# Manual build with version
go build -ldflags "-X main.version=v1.0.0" -o fest cmd/fest/main.go

# Via Makefile (future enhancement)
make build VERSION=v1.0.0
```

## Monitoring Distribution

Track:
- Download counts on GitHub Releases
- Issue reports for platform-specific problems
- Security scanner results
- User feedback on installation process

## Resources

- [Apple Developer - Notarizing macOS Software](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [Windows - Introduction to Code Signing](https://docs.microsoft.com/en-us/windows/win32/seccrypto/cryptography-tools)
- [GitHub Actions - Creating Releases](https://docs.github.com/en/actions/guides/creating-releases)