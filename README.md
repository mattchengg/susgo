# susgo

Samsung firmware downloader with intuitive GUI - built with Go and Fyne

## Overview

susgo is a cross-platform desktop application for downloading and managing Samsung firmware files. Built with Go and the Fyne UI framework, it provides an easy-to-use graphical interface for checking firmware updates, downloading firmware files, and decrypting encrypted firmware packages.

## Features

### GUI Interface
- **Cross-platform desktop application** - Works on Windows, Linux, and macOS
- **Modern, intuitive interface** - Built with Fyne for a native look and feel
- **Three main tabs** for different operations:
  - **Check Update**: Query latest firmware versions for any Samsung device
  - **Download**: Download firmware with real-time progress tracking
  - **Decrypt**: Decrypt encrypted firmware files (V2 and V4 encryption)

### Core Functionality
- ✅ Check latest firmware versions for any Samsung model and region
- ✅ Download firmware files with progress tracking (speed, ETA, percentage)
- ✅ Resume interrupted downloads automatically
- ✅ Auto-decrypt firmware after download
- ✅ Manual firmware decryption with file selection dialogs
- ✅ Support for both V2 and V4 encryption formats
- ✅ IMEI/TAC validation and generation
- ✅ Input validation with helpful error messages
- ✅ Success notifications and error dialogs

### Technical Features
- Pure Go implementation with no external dependencies
- Supports Standard CSCs and EUX/EUY regions
- Single binary distribution
- Real-time progress reporting
- Efficient chunk-based downloading (32 KB chunks)
- File existence checks to avoid re-downloading

## Installation

### Option 1: Download Pre-built Binary (Recommended)

Download the latest release for your platform from [Releases](https://github.com/mattchengg/susgo/releases):
- **Windows**: `susgo.exe`
- **macOS**: `susgo.app` (Intel and Apple Silicon)
- **Linux**: `susgo` (AppImage or binary)

### Option 2: Build from Source

#### Prerequisites
- Go 1.25 or later (Go 1.25.4 recommended)
- Platform-specific dependencies for Fyne:

**Ubuntu/Debian**:
```bash
sudo apt-get install gcc libgl1-mesa-dev xorg-dev
```

**Fedora/RHEL**:
```bash
sudo dnf install gcc mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel
```

**macOS**:
```bash
# Xcode Command Line Tools required
xcode-select --install
```

**Windows**:
- Install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or [MSYS2](https://www.msys2.org/)

#### Build Instructions

```bash
# Clone the repository
git clone https://github.com/mattchengg/susgo.git
cd susgo

# Install dependencies
go mod download

# Build the application
go build -o susgo

# (Windows) Build without console window
go build -ldflags="-H windowsgui" -o susgo.exe
```

## Usage

### Launch the Application

Simply run the executable:
- **Windows**: Double-click `susgo.exe` or run from command line
- **Linux**: `./susgo` or double-click if executable permissions are set
- **macOS**: Open `susgo.app` or run from terminal

### Using the GUI

#### Tab 1: Check Update
1. Enter the device **Model** (e.g., `SM-S928B`)
2. Enter the **Region** code (e.g., `EUX`, `XAR`, `DBT`)
3. Click **Check Update**
4. View the latest firmware version in the success dialog

#### Tab 2: Download
1. Enter the device **Model** (e.g., `SM-S928B`)
2. Enter the **Region** code (e.g., `EUX`)
3. Enter **IMEI/TAC** (8 digits for TAC or 15 digits for full IMEI)
4. (Optional) Enter specific firmware **Version** (leave empty for latest)
5. Enter or select **Output Directory** where firmware will be saved
6. Click **Start Download**
7. Monitor progress with real-time speed, ETA, and percentage
8. Firmware will be automatically decrypted after download

#### Tab 3: Decrypt
1. Enter the device **Model** (e.g., `SM-S928B`)
2. Enter the **Region** code (e.g., `EUX`)
3. Enter **IMEI/TAC** (8 digits for TAC or 15 digits for full IMEI)
4. Enter firmware **Version** (e.g., `S928BXXS4CYK8/S928BOXM4CYK8/...`)
5. Click **Browse** to select the encrypted input file (`.enc2` or `.enc4`)
6. Click **Browse** to choose output location for decrypted firmware
7. Select **Encryption Version** (V2 or V4, defaults to V4)
8. Click **Start Decrypt**
9. Monitor decryption progress in real-time

### Input Guidelines

| Field | Format | Example |
|-------|--------|---------|
| Model | Must start with `SM-` | `SM-S928B`, `SM-G998B` |
| Region | 2-4 letter code | `EUX`, `XAR`, `DBT`, `BTU` |
| IMEI/TAC | 8 digits (TAC) or 15 digits (IMEI) | `35123456` or `351234567890123` |
| Version | Full firmware string | `S928BXXS4CYK8/S928BOXM4CYK8/S928BXXS4CYK8/S928BXXS4CYK8` |

## Examples

### Example 1: Check Latest Firmware
- Model: `SM-S928B`
- Region: `EUX`
- Click "Check Update"
- Result: `S928BXXS4CYK8/S928BOXM4CYK8/S928BXXS4CYK8/S928BXXS4CYK8`

### Example 2: Download Latest Firmware
- Model: `SM-S928B`
- Region: `EUX`
- IMEI/TAC: `35123456` (TAC)
- Version: (leave empty for latest)
- Output Directory: `/home/user/firmware`
- The firmware will be downloaded and automatically decrypted

### Example 3: Decrypt Firmware Manually
- Model: `SM-S928B`
- Region: `EUX`
- IMEI/TAC: `35123456`
- Version: `S928BXXS4CYK8/S928BOXM4CYK8/S928BXXS4CYK8/S928BXXS4CYK8`
- Input File: Browse to select `firmware.enc4`
- Output File: Browse to choose save location
- Encryption Version: V4
- Click "Start Decrypt"

## Building Application Packages

### Using Fyne Package Tool

Install the Fyne command-line tool:
```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
```

Package for different platforms:

```bash
# Current platform (auto-detect)
fyne package -icon icon.png

# Windows
fyne package -os windows -icon icon.png

# macOS (creates .app bundle)
fyne package -os darwin -icon icon.png

# Linux (creates AppImage)
fyne package -os linux -icon icon.png
```

### Cross-Compilation

Build for different platforms using Go's cross-compilation:

```bash
# Windows (from Linux/macOS)
GOOS=windows GOARCH=amd64 go build -o susgo.exe

# macOS Intel (from Linux/Windows)
GOOS=darwin GOARCH=amd64 go build -o susgo-mac-intel

# macOS Apple Silicon (from Linux/Windows)
GOOS=darwin GOARCH=arm64 go build -o susgo-mac-arm

# Linux (from Windows/macOS)
GOOS=linux GOARCH=amd64 go build -o susgo-linux
```

## Troubleshooting

### Linux: Application won't start or shows OpenGL errors

Make sure you have the required graphics libraries installed:

```bash
# Ubuntu/Debian
sudo apt-get install libgl1-mesa-dev libgl1-mesa-glx

# Fedora/RHEL
sudo dnf install mesa-libGL mesa-libGL-devel

# Arch
sudo pacman -S mesa
```

### Windows: Console window appears

Build with the `-H windowsgui` flag to hide the console window:
```bash
go build -ldflags="-H windowsgui" -o susgo.exe
```

### macOS: "App is damaged and can't be opened"

This occurs with unsigned apps. Remove the quarantine attribute:
```bash
xattr -cr susgo.app
```

Or right-click the app, select "Open", and click "Open" in the security dialog.

### Download fails or shows connection errors

- Check your internet connection
- Verify the model and region codes are correct
- Samsung servers may be temporarily unavailable - try again later
- Some regions may not have firmware available

### Decryption fails

- Ensure you're using the correct encryption version (V2 or V4)
- Verify the firmware version string matches the downloaded file
- Check that IMEI/TAC is correct (must match what was used for download)
- Make sure the input file is a valid encrypted firmware file

### IMEI validation error

- TAC: Must be exactly 8 numeric digits
- Full IMEI: Must be exactly 15 numeric digits
- No spaces, hyphens, or other characters allowed

## FAQ

**Q: Do I need an internet connection?**  
A: Yes, the application needs internet to communicate with Samsung's FUS servers for checking updates and downloading firmware.

**Q: Is my IMEI sent to Samsung?**  
A: Yes, the IMEI or TAC is required by Samsung's FUS API for authentication and firmware access. It's a standard part of their firmware distribution system.

**Q: Can I download firmware for any Samsung device?**  
A: You can download firmware for any device that has firmware available on Samsung's FUS servers. Most recent Samsung smartphones and tablets are supported.

**Q: What's the difference between TAC and full IMEI?**  
A: TAC (Type Allocation Code) is the first 8 digits of an IMEI. Both work for firmware downloads - TAC is simpler and doesn't identify a specific device.

**Q: Why does the download auto-decrypt?**  
A: Most firmware files from Samsung are encrypted. Auto-decryption saves you a manual step and ensures you have a usable firmware file immediately.

**Q: Can I pause a download?**  
A: No, but if a download is interrupted, it will automatically resume from where it left off when you start the download again.

**Q: Where can I find my device's model and region code?**  
A: 
- Model: Settings → About phone → Model number (e.g., SM-S928B)
- Region: Check your firmware version in Settings, or use common codes (EUX for Europe, XAR for Arabic, DBT for Germany, etc.)

**Q: What are .enc2 and .enc4 files?**  
A: These are encrypted firmware files. .enc2 uses V2 encryption (older), .enc4 uses V4 encryption (newer). The decrypt function handles both.

**Q: Is this application official?**  
A: No, this is a third-party tool. It uses Samsung's official firmware distribution servers but is not affiliated with or endorsed by Samsung.

## Technical Details

### Architecture
- **Language**: Go 1.25.4
- **GUI Framework**: Fyne v2.7.2
- **Binary Size**: ~20-30 MB (includes GUI framework)
- **Memory Usage**: ~50-100 MB during operation
- **Download Chunk Size**: 32 KB
- **Decryption Chunk Size**: 4 KB

### Supported Platforms
- Windows 10/11 (64-bit)
- macOS 10.15+ (Intel and Apple Silicon)
- Linux (major distributions with X11 or Wayland)

### Protocol
- Uses Samsung's FUS (Firmware Update Server) API
- HTTPS communication with Samsung servers
- Authentication via generated IMEI/TAC
- AES encryption for firmware files (V2 and V4 formats)

## Development

### Project Structure
```
susgo/
├── main.go              # GUI entry point and UI logic
├── helpers.go           # Helper functions (IMEI, download, decrypt)
├── progress.go          # Progress reporting (CLI and GUI)
├── fusclient.go         # Samsung FUS API client
├── versionfetch.go      # Firmware version queries
├── crypt.go             # Encryption/decryption logic
├── auth.go              # Authentication utilities
├── request.go           # XML request builders
├── imei.go              # IMEI validation and generation
├── go.mod               # Go module dependencies
└── go.sum               # Dependency checksums
```

### Contributing
Contributions are welcome! Please feel free to submit issues or pull requests.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Running Tests
```bash
go test ./...
```

### Code Style
This project follows standard Go conventions:
- Run `gofmt` or `goimports` before committing
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and modular

## License

This project is licensed under the terms specified in the LICENSE file.

## Credits

- [samloader](https://github.com/ananjaser1211/samloader/) - Original Python implementation
- [Fyne](https://fyne.io/) - Cross-platform GUI framework for Go
- Samsung - Firmware distribution infrastructure

## Disclaimer

This tool is for educational and backup purposes only. Users are responsible for ensuring they have the right to download and use Samsung firmware. Always back up your data before flashing firmware to your device.

## Changelog

### Version 2.0 (GUI Release)
- ✨ Complete GUI rewrite using Fyne framework
- ✨ Three-tab interface (Check Update, Download, Decrypt)
- ✨ Real-time progress tracking with speed and ETA
- ✨ File selection dialogs for decrypt functionality
- ✨ Input validation with helpful error messages
- ✨ Success notifications and error dialogs
- ✨ Auto-resume interrupted downloads
- ✨ Auto-decrypt after download
- ✨ Cross-platform support (Windows, macOS, Linux)
- 🗑️ Removed list command (no longer needed in GUI)

### Version 1.0 (CLI)
- Initial CLI implementation
- Check updates, list versions, download, and decrypt functionality