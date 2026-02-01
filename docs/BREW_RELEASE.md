# RTUI Homebrew Release Guide

Purpose: publish RTUI for macOS via Homebrew using prebuilt binaries.

## Preconditions
- GitHub repo is public (brew cannot download private release assets).
- gh CLI is logged in.
- Homebrew and Go are installed.

## 1) Tag the release
```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

## 2) Build macOS binaries
```bash
mkdir -p dist
GOOS=darwin GOARCH=arm64 go build -o dist/rtui-darwin-arm64 ./cmd/rtui
GOOS=darwin GOARCH=amd64 go build -o dist/rtui-darwin-amd64 ./cmd/rtui
```

## 3) Package and checksum
```bash
cd dist
tar -czf rtui_darwin_arm64.tar.gz rtui-darwin-arm64
tar -czf rtui_darwin_amd64.tar.gz rtui-darwin-amd64
shasum -a 256 rtui_darwin_arm64.tar.gz rtui_darwin_amd64.tar.gz
```

## 4) Create GitHub release with assets
```bash
gh release create v0.1.0 \
  dist/rtui_darwin_arm64.tar.gz \
  dist/rtui_darwin_amd64.tar.gz \
  -t "v0.1.0" -n "RTUI v0.1.0"
```

## 5) Create Homebrew tap
```bash
gh repo create mrekoj/homebrew-rtui --public --description "Homebrew tap for RTUI"
git clone https://github.com/mrekoj/homebrew-rtui
```

## 6) Add the formula
Create `Formula/rtui.rb`:
```ruby
class Rtui < Formula
  desc "Minimal TUI dashboard to monitor and manage multiple git repos"
  homepage "https://github.com/mrekoj/rtui"
  version "0.1.0"

  if Hardware::CPU.arm?
    url "https://github.com/mrekoj/rtui/releases/download/v0.1.0/rtui_darwin_arm64.tar.gz"
    sha256 "REPLACE_ARM_SHA"
  else
    url "https://github.com/mrekoj/rtui/releases/download/v0.1.0/rtui_darwin_amd64.tar.gz"
    sha256 "REPLACE_AMD_SHA"
  end

  def install
    bin.install "rtui-darwin-#{Hardware::CPU.arm? ? "arm64" : "amd64"}" => "rtui"
  end

  test do
    system "#{bin}/rtui", "-h"
  end
end
```

Commit and push:
```bash
cd homebrew-rtui
git add Formula/rtui.rb
git commit -m "Add rtui formula"
git push
```

## 7) Install and test
```bash
brew tap mrekoj/rtui
brew install rtui
rtui -h
```

## Update version (repeatable)

### 1) Pick version
```bash
VERSION=0.1.1
TAG="v${VERSION}"
```

### 2) Tag and push
```bash
git tag -a "$TAG" -m "$TAG"
git push origin "$TAG"
```

### 3) Build macOS binaries
```bash
rm -rf dist
mkdir -p dist
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/rtui-darwin-arm64 ./cmd/rtui
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/rtui-darwin-amd64 ./cmd/rtui
```

### 4) Package and checksum
```bash
cd dist
tar -czf rtui_darwin_arm64.tar.gz rtui-darwin-arm64
tar -czf rtui_darwin_amd64.tar.gz rtui-darwin-amd64
shasum -a 256 rtui_darwin_arm64.tar.gz rtui_darwin_amd64.tar.gz
```

### 5) Create GitHub release
```bash
gh release create "$TAG" \
  dist/rtui_darwin_arm64.tar.gz \
  dist/rtui_darwin_amd64.tar.gz \
  -t "$TAG" -n "RTUI $TAG"
```

### 6) Update Homebrew formula
Edit `Formula/rtui.rb`:
- Bump `version` to `VERSION`
- Update both `url` entries to the new tag
- Replace `sha256` values with the new checksums

Commit and push:
```bash
cd ../homebrew-rtui
git add Formula/rtui.rb
git commit -m "Bump rtui to $TAG"
git push
```

### 7) Verify install
```bash
brew update
brew upgrade rtui
rtui -h
```

*Last updated: February 1, 2026*
