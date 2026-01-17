# GitHub Actions Release Workflow Fix

## Overview
Fixed the GitHub Actions release workflow to properly build sfgo with Fyne GUI support by replacing CGO_ENABLED=0 builds with fyne-cross tool.

## Problem
The previous workflow failed because:
- Fyne framework requires CGO (C bindings for OpenGL, system libraries)
- Workflow used `CGO_ENABLED=0` for "pure Go" cross-compilation
- CGO cross-compilation requires platform-specific toolchains (C compilers, system headers)
- Setting up CGO for multiple platforms manually is complex and error-prone

## Solution
Replaced the matrix-based build strategy with **fyne-cross**, a specialized tool for Fyne applications:

### What is fyne-cross?
- Official Fyne tool for cross-platform builds
- Uses Docker containers with pre-configured toolchains
- Handles all CGO requirements automatically
- Supports Linux, Windows, macOS, Android, iOS

### Benefits
1. **Automatic CGO Setup**: Each Docker container has the right compilers and libraries
2. **Consistent Builds**: Same environment every time
3. **No Manual Configuration**: No need to install platform-specific toolchains
4. **Proper Fyne Packaging**: Generates correct app structures (.app for macOS, etc.)

## Changes Made

### Before (CGO_ENABLED=0)
```yaml
- name: Build
  env:
    GOOS: ${{ matrix.goos }}
    GOARCH: ${{ matrix.goarch }}
    CGO_ENABLED: 0  # ❌ Breaks Fyne
  run: go build -o sfgo
```

### After (fyne-cross)
```yaml
- name: Install fyne-cross
  run: go install github.com/fyne-io/fyne-cross@latest

- name: Build for Linux
  run: fyne-cross linux -arch=amd64,arm64 -app-id=com.mattchengg.sfgo -release

- name: Build for Windows
  run: fyne-cross windows -arch=amd64 -app-id=com.mattchengg.sfgo -release

- name: Build for macOS
  run: fyne-cross darwin -arch=amd64,arm64 -app-id=com.mattchengg.sfgo -release
```

## Platforms Supported
| Platform | Architectures | Output |
|----------|--------------|--------|
| Linux | amd64, arm64 | `sfgo` binary |
| Windows | amd64 | `sfgo.exe` |
| macOS | amd64, arm64 | `sfgo.app` bundle + binary |

## Removed Platforms
These platforms are not supported by Fyne:
- ❌ Linux 386 (32-bit)
- ❌ Linux ARM v7
- ❌ Windows 386 (32-bit)
- ❌ Windows ARM64 (limited Fyne support)
- ❌ FreeBSD (no Fyne support)
- ❌ Android (requires special packaging, not suitable for this workflow)

## Artifact Structure
```
artifacts/
├── sfgo-linux-amd64
├── sfgo-linux-arm64
├── sfgo-windows-amd64.exe
├── sfgo-darwin-amd64
├── sfgo-darwin-arm64
├── sfgo-darwin-amd64.app.tar.gz  # Full .app bundle
├── sfgo-darwin-arm64.app.tar.gz  # Full .app bundle
└── checksums.txt
```

## Testing the Workflow

### Manual Trigger
```bash
# Via GitHub UI: Actions → Release → Run workflow
```

### Tag-based Release
```bash
git tag v0.0.5
git push origin v0.0.5
# Workflow runs automatically, creates GitHub Release
```

### Local Testing (requires Docker)
```bash
# Install fyne-cross
go install github.com/fyne-io/fyne-cross@latest

# Build for Linux
fyne-cross linux -arch=amd64

# Build for Windows
fyne-cross windows -arch=amd64

# Build for macOS
fyne-cross darwin -arch=amd64
```

## Expected Build Times
- **fyne-cross install**: ~30 seconds
- **Linux build**: 3-5 minutes (first run downloads Docker images)
- **Windows build**: 3-5 minutes
- **macOS build**: 3-5 minutes
- **Total**: ~10-15 minutes

## Troubleshooting

### Docker Permission Issues
If the workflow fails with Docker permission errors:
```yaml
- name: Setup Docker
  run: |
    sudo groupadd docker || true
    sudo usermod -aG docker $USER || true
```

### Out of Disk Space
fyne-cross Docker images are large (~2GB each). If space is limited:
```yaml
- name: Clean Docker
  run: docker system prune -af
```

### macOS Codesigning (Future)
For App Store distribution, add codesigning:
```yaml
- name: Build for macOS
  run: fyne-cross darwin -arch=amd64,arm64 -app-id=com.mattchengg.sfgo -release
  env:
    MACOS_CERTIFICATE: ${{ secrets.MACOS_CERTIFICATE }}
    MACOS_CERTIFICATE_PWD: ${{ secrets.MACOS_CERTIFICATE_PWD }}
```

## Related Documentation
- [fyne-cross GitHub](https://github.com/fyne-io/fyne-cross)
- [Fyne Packaging Guide](https://developer.fyne.io/started/packaging)
- [CGO and Fyne](https://developer.fyne.io/started/cross-compiling)

## Commit
```
commit cf0bec1b7fbeab8edce18f9bb8514eb511ff0753
Author: mattchengg <mattcheng20080205@gmail.com>
Date:   Sat Jan 17 17:06:23 2026 +0800

    fix: replace release workflow with fyne-cross for CGO support
```

## Status
✅ **Complete** - Ready for next release

---
*Last Updated: 2026-01-17*
