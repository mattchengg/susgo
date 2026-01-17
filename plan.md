# susgo GUI Migration Implementation Plan

## Project Overview
Transform susgo from a CLI-based Samsung firmware downloader into a cross-platform GUI application using Fyne while maintaining all existing core functionality except the "list" command.

## Current Architecture Analysis

### Existing Components
1. **main.go**: CLI entry point with command parsing (checkupdate, list, download, decrypt)
2. **Core Logic Modules** (NO changes needed):
   - `auth.go`: Authentication and encryption utilities
   - `crypt.go`: Firmware decryption (V2/V4 keys)
   - `fusclient.go`: HTTP client for Samsung FUS API
   - `imei.go`: IMEI validation and generation
   - `progress.go`: Terminal progress bar (will be replaced)
   - `request.go`: XML request builders
   - `versionfetch.go`: Firmware version retrieval

### Dependencies
- Current: Go standard library only
- New: `fyne.io/fyne/v2` (cross-platform GUI framework)

## Changes Required

### Phase 1: Code Removal (List Command)
**Goal**: Remove the list command functionality completely

**Files to Modify**:
- `main.go`

**Specific Removals**:
1. **Function `parseListFlags()`** (lines 100-105)
   - Parses -l and -q flags for list command
   - Can be completely removed

2. **Function `listFirmware()`** (lines 142-173)
   - Displays firmware versions in formatted output
   - Uses `getVersionInfo()` from versionfetch.go
   - Can be completely removed

3. **Case statement in main()** (lines 45-47)
   ```go
   case "list":
       parseListFlags(args[1:])
       listFirmware()
   ```
   - Remove this entire case block

4. **Global variables** (lines 14-27)
   - Remove: `latest`, `quiet` (only used by list command)
   - Keep all others: needed by remaining commands

5. **printUsage() function** (lines 61-98)
   - Remove list-related documentation:
     - Line 66: `susgo -m <model> -r <region> list [-l] [-q]`
     - Lines 78: "list" command description
     - Lines 83-84: List Options section

**Note**: Keep `getVersionInfo()` in versionfetch.go - it may be useful for future GUI enhancements

### Phase 2: GUI Architecture Design

#### 2.1 Application Structure
```
susgo-gui/
├── main.go          # GUI entry point with Fyne setup
├── gui/
│   ├── app.go       # Main application window
│   ├── checkupdate.go  # Check update tab/view
│   ├── download.go     # Download tab/view
│   ├── decrypt.go      # Decrypt tab/view
│   └── progress.go     # GUI progress widget
├── core/            # Move existing logic here
│   ├── auth.go
│   ├── crypt.go
│   ├── fusclient.go
│   ├── imei.go
│   ├── request.go
│   └── versionfetch.go
└── go.mod
```

**Alternative Simpler Structure** (Recommended for Phase 1):
```
susgo/
├── main.go          # GUI version (replace CLI main)
├── gui.go           # All GUI code in one file initially
├── auth.go          # No changes
├── crypt.go         # No changes
├── fusclient.go     # No changes
├── imei.go          # No changes
├── progress.go      # Add GUI progress methods
├── request.go       # No changes
├── versionfetch.go  # No changes
└── go.mod           # Add fyne dependency
```

#### 2.2 GUI Components

**Main Window**:
- Title: "susgo - Samsung Firmware Downloader"
- Size: 700x500 pixels (default)
- Resizable: Yes
- Layout: Tabbed interface with 3 tabs

**Tab 1: Check Update**
- Input: Model (Entry)
- Input: Region (Entry)
- Button: Check Update
- Output: Latest version display (Label/Rich Text)
- Status: Success/Error messages

**Tab 2: Download**
- Input: Model (Entry)
- Input: Region (Entry)
- Input: IMEI/TAC (Entry with help text)
- Input: Serial Number (Entry - optional)
- Input: Version (Entry - optional, defaults to latest)
- Button: Select Output Directory (File Dialog)
- Label: Selected path display
- Checkbox: Show MD5 hash
- Button: Start Download
- Progress Bar: Download progress
- Label: Status (speed, ETA, size)
- Button: Cancel (to stop download)

**Tab 3: Decrypt**
- Input: Model (Entry)
- Input: Region (Entry)
- Input: IMEI/TAC (Entry with help text)
- Input: Serial Number (Entry - optional)
- Input: Version (Entry with help text)
- Button: Select Input File (File Dialog)
- Button: Select Output File (File Dialog)
- Radio Group: Encryption Version (V2 / V4 - default V4)
- Button: Start Decrypt
- Progress Bar: Decryption progress
- Label: Status messages

**Common UI Elements**:
- Help icons with tooltips
- Form validation with error highlights
- Consistent padding and spacing
- Responsive layout

### Phase 3: Core Logic Adaptation

#### 3.1 Progress Reporting
**Current**: Terminal-based with `progress.go`
- Prints to stdout with ANSI codes
- Updates every 100ms

**New**: GUI-compatible interface
```go
type ProgressReporter interface {
    SetTotal(total int64)
    SetCurrent(current int64)
    Add(delta int64)
    SetStatus(message string)
    Finish()
}

// Terminal implementation (keep for potential CLI mode)
type TerminalProgress struct { ... }

// GUI implementation
type GUIProgress struct {
    bar *widget.ProgressBar
    label *widget.Label
    // ... Fyne widgets
}
```

**Strategy**: 
- Add interface to `progress.go`
- Implement both Terminal and GUI versions
- Pass progress reporter to download/decrypt functions

#### 3.2 Function Signature Changes

**Minimal Changes Approach**:

1. **Download function** - Add progress callback:
```go
// Old (current)
func download()

// New - make it callable from GUI
func downloadFirmware(model, region, imei, serial, version, outputPath string, 
                      showMD5 bool, progress ProgressReporter) error
```

2. **Decrypt function** - Add progress callback:
```go
// Old (current)
func decrypt()

// New
func decryptFirmware(inFile, outFile string, key []byte, 
                     progress ProgressReporter) error
// Already exists but needs progress interface
```

3. **Check Update** - Already clean:
```go
func getLatestVersion(model, region string) (string, error)
// No changes needed - already returns result
```

#### 3.3 Error Handling
**Current**: Prints to stderr and calls os.Exit(1)

**New**: Return errors to GUI layer
- Remove all `os.Exit()` calls from core functions
- Return errors via `error` return values
- GUI layer displays errors in dialog boxes

### Phase 4: Fyne Integration Details

#### 4.1 Dependencies
Add to `go.mod`:
```go
require (
    fyne.io/fyne/v2 v2.4.5  // Latest stable version
)
```

Install command:
```bash
go get fyne.io/fyne/v2@latest
```

#### 4.2 Main Entry Point
```go
package main

import (
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

func main() {
    myApp := app.New()
    myWindow := myApp.NewWindow("susgo - Samsung Firmware Downloader")
    
    // Create tabs
    tabs := container.NewAppTabs(
        container.NewTabItem("Check Update", makeCheckUpdateTab()),
        container.NewTabItem("Download", makeDownloadTab()),
        container.NewTabItem("Decrypt", makeDecryptTab()),
    )
    
    myWindow.SetContent(tabs)
    myWindow.Resize(fyne.NewSize(700, 500))
    myWindow.ShowAndRun()
}
```

#### 4.3 Check Update Tab Implementation
```go
func makeCheckUpdateTab() *fyne.Container {
    modelEntry := widget.NewEntry()
    modelEntry.SetPlaceHolder("e.g., SM-S928B")
    
    regionEntry := widget.NewEntry()
    regionEntry.SetPlaceHolder("e.g., EUX")
    
    resultLabel := widget.NewLabel("")
    
    checkButton := widget.NewButton("Check Update", func() {
        model := strings.TrimSpace(modelEntry.Text)
        region := strings.TrimSpace(regionEntry.Text)
        
        if model == "" || region == "" {
            resultLabel.SetText("Error: Model and Region required")
            return
        }
        
        resultLabel.SetText("Checking...")
        
        // Run in goroutine to avoid blocking UI
        go func() {
            version, err := getLatestVersion(model, region)
            if err != nil {
                resultLabel.SetText("Error: " + err.Error())
            } else {
                resultLabel.SetText("Latest Version: " + version)
            }
        }()
    })
    
    form := container.NewVBox(
        widget.NewLabel("Model:"),
        modelEntry,
        widget.NewLabel("Region:"),
        regionEntry,
        checkButton,
        widget.NewSeparator(),
        resultLabel,
    )
    
    return container.NewPadded(form)
}
```

#### 4.4 Download Tab Implementation (Simplified)
```go
func makeDownloadTab() *fyne.Container {
    // Input fields
    modelEntry := widget.NewEntry()
    regionEntry := widget.NewEntry()
    imeiEntry := widget.NewEntry()
    serialEntry := widget.NewEntry()
    versionEntry := widget.NewEntry()
    
    // Output selection
    outputPath := ""
    outputLabel := widget.NewLabel("No directory selected")
    selectDirButton := widget.NewButton("Select Output Directory", func() {
        dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
            if err == nil && uri != nil {
                outputPath = uri.Path()
                outputLabel.SetText(outputPath)
            }
        }, myWindow)
    })
    
    // Progress
    progressBar := widget.NewProgressBar()
    statusLabel := widget.NewLabel("")
    
    // Download button
    downloadButton := widget.NewButton("Start Download", func() {
        // Validate inputs
        model := strings.TrimSpace(modelEntry.Text)
        region := strings.TrimSpace(regionEntry.Text)
        // ... validate others
        
        // Create progress reporter
        guiProgress := &GUIProgress{
            bar: progressBar,
            label: statusLabel,
        }
        
        // Run download in goroutine
        go func() {
            err := downloadFirmware(model, region, imei, serial, version, 
                                   outputPath, false, guiProgress)
            if err != nil {
                statusLabel.SetText("Error: " + err.Error())
            } else {
                statusLabel.SetText("Download complete!")
            }
        }()
    })
    
    // Layout
    form := container.NewVBox(
        widget.NewForm(
            widget.NewFormItem("Model", modelEntry),
            widget.NewFormItem("Region", regionEntry),
            widget.NewFormItem("IMEI/TAC", imeiEntry),
            widget.NewFormItem("Serial", serialEntry),
            widget.NewFormItem("Version", versionEntry),
        ),
        selectDirButton,
        outputLabel,
        downloadButton,
        progressBar,
        statusLabel,
    )
    
    return container.NewPadded(form)
}
```

#### 4.5 Decrypt Tab Implementation
Similar structure to Download tab with:
- Input file selector (dialog.ShowFileOpen)
- Output file selector (dialog.ShowFileSave)
- Encryption version radio buttons
- Progress bar for decryption

### Phase 5: Cross-Platform Considerations

#### 5.1 Build Process

**Windows**:
```bash
# Standard build
go build -o susgo.exe

# With Windows GUI (no console window)
go build -ldflags="-H windowsgui" -o susgo.exe
```

**macOS**:
```bash
# Standard build
go build -o susgo

# Create .app bundle (requires fyne CLI tool)
go install fyne.io/fyne/v2/cmd/fyne@latest
fyne package -os darwin -icon icon.png
```

**Linux**:
```bash
# Standard build
go build -o susgo

# May require additional system packages:
# Debian/Ubuntu: libgl1-mesa-dev xorg-dev
# Fedora: mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel
```

**Android**:
```bash
# Requires Android SDK and NDK
fyne package -os android -appID com.mattchengg.susgo -icon icon.png

# Additional considerations:
# - File picker requires storage permissions
# - Network requires INTERNET permission
# - May need to handle different screen sizes
```

#### 5.2 Platform-Specific Features

**File Dialogs**:
- Fyne handles platform differences automatically
- Use `dialog.ShowFolderOpen()`, `dialog.ShowFileOpen()`, `dialog.ShowFileSave()`

**File Paths**:
- Use `filepath.Join()` for path construction (already used)
- Fyne URIs work across platforms

**Permissions** (Android):
- Storage: Handled by Fyne file dialogs
- Network: Add to AndroidManifest.xml if packaging

#### 5.3 Icon and Assets
Create application icon:
- Size: 512x512 PNG recommended
- Location: `icon.png` in project root
- Fyne will generate platform-specific formats

### Phase 6: Testing Strategy

#### 6.1 Unit Testing
**Keep existing logic testable**:
```go
// Test core functions independently
func TestGetLatestVersion(t *testing.T) { ... }
func TestIMEIValidation(t *testing.T) { ... }
func TestDecryption(t *testing.T) { ... }
```

#### 6.2 Integration Testing
**Test GUI integration**:
- Use Fyne's test package: `fyne.io/fyne/v2/test`
- Test button clicks, input validation
```go
func TestCheckUpdateButton(t *testing.T) {
    app := test.NewApp()
    window := test.NewWindow(makeCheckUpdateTab())
    // ... simulate user input
}
```

#### 6.3 Manual Testing Checklist
- [ ] Check Update: Valid model/region
- [ ] Check Update: Invalid model/region (error handling)
- [ ] Download: Full download with progress
- [ ] Download: Resume interrupted download
- [ ] Download: Auto-decrypt after download
- [ ] Download: IMEI generation from TAC
- [ ] Download: Invalid IMEI handling
- [ ] Decrypt: V2 encryption
- [ ] Decrypt: V4 encryption
- [ ] Decrypt: Invalid file handling
- [ ] File dialogs work on each platform
- [ ] Window resize/layout
- [ ] Application icon displays correctly

#### 6.4 Platform Testing
Test on each target platform:
- [ ] Windows 10/11
- [ ] macOS 12+ (Intel and Apple Silicon)
- [ ] Linux (Ubuntu, Fedora)
- [ ] Android 8.0+ (optional, if targeting mobile)

### Phase 7: Migration Path

#### 7.1 Incremental Approach (Recommended)

**Step 1**: Remove list command (no dependencies)
- Remove functions and case statement
- Update usage text
- Test remaining CLI functionality

**Step 2**: Create basic GUI with Check Update only
- Add Fyne dependency
- Create simple single-tab window
- Test on one platform

**Step 3**: Add Download tab
- Implement file selection
- Add progress reporting
- Test download and auto-decrypt

**Step 4**: Add Decrypt tab
- Implement file selection
- Add progress reporting
- Test both encryption versions

**Step 5**: Polish and cross-platform testing
- Add icons and branding
- Test on all platforms
- Create build scripts

#### 7.2 Backward Compatibility Option

**Keep CLI mode** (optional):
```go
func main() {
    if len(os.Args) > 1 {
        // CLI mode - parse flags and run commands
        runCLI()
    } else {
        // GUI mode - no arguments provided
        runGUI()
    }
}
```

Benefits:
- Allows scripting/automation
- Useful for servers without GUI
- Gradual migration for users

Drawbacks:
- More code to maintain
- Larger binary size

**Decision**: Start GUI-only, add CLI mode later if needed

### Phase 8: Documentation Updates

#### 8.1 README.md Updates
- Update description: "GUI application for downloading Samsung firmware"
- Add screenshots of main window and tabs
- Update build instructions with Fyne requirements
- Add platform-specific installation guides
- Update usage section with GUI workflow

#### 8.2 New Documentation Files
- `BUILDING.md`: Detailed build instructions per platform
- `CONTRIBUTING.md`: How to contribute to GUI development
- `CHANGELOG.md`: Track changes from CLI to GUI

#### 8.3 Code Documentation
- Add godoc comments to new GUI functions
- Document the ProgressReporter interface
- Add examples in comments

## Technical Specifications

### Dependency Versions
- Go: 1.25.4 (current)
- Fyne: v2.4.5 or later
- Minimum Go version: 1.19 (Fyne requirement)

### Build Artifacts
- Windows: susgo.exe (~15-20 MB with Fyne)
- macOS: susgo.app bundle
- Linux: susgo binary
- Android: susgo.apk (optional)

### Performance Considerations
- Download: No performance impact (network-bound)
- Decryption: Same performance (CPU-bound)
- GUI overhead: Minimal (~5-10 MB RAM for Fyne)
- Startup time: ~1-2 seconds (Fyne initialization)

### Security Considerations
- No changes to encryption/auth logic
- File dialogs prevent path injection
- Input validation in GUI layer
- Error messages don't expose sensitive data

## Risk Assessment

### Low Risk
- ✅ Core logic remains unchanged
- ✅ Fyne is mature and stable
- ✅ No breaking changes to algorithms
- ✅ List command removal is isolated

### Medium Risk
- ⚠️ Cross-platform testing requires multiple environments
- ⚠️ Android packaging is complex
- ⚠️ File permission issues on mobile

### Mitigation Strategies
- Use virtual machines for platform testing
- Start with desktop platforms only
- Add Android support in later phase
- Comprehensive error handling in GUI

## Timeline Estimate

### Phase 1: List Command Removal
- Time: 1-2 hours
- Tasks: Code removal, testing, documentation update

### Phase 2: Basic GUI Setup
- Time: 4-6 hours
- Tasks: Add Fyne, create main window, Check Update tab

### Phase 3: Download Tab
- Time: 6-8 hours
- Tasks: Form creation, file dialogs, progress integration

### Phase 4: Decrypt Tab
- Time: 4-6 hours
- Tasks: Form creation, file dialogs, progress integration

### Phase 5: Polish and Testing
- Time: 6-8 hours
- Tasks: Layout refinement, error handling, cross-platform testing

### Phase 6: Documentation
- Time: 2-4 hours
- Tasks: README, screenshots, build guides

**Total Estimated Time**: 23-34 hours

## Success Criteria

### Functional Requirements
- ✅ All existing functionality works in GUI (except list)
- ✅ Downloads can be started, monitored, and completed
- ✅ Decryption works with both V2 and V4
- ✅ Check update displays latest version
- ✅ File dialogs work on all platforms
- ✅ Progress bars update correctly

### Non-Functional Requirements
- ✅ Application starts within 2 seconds
- ✅ UI remains responsive during operations
- ✅ Builds successfully on Windows, macOS, Linux
- ✅ Binary size under 30 MB
- ✅ Memory usage under 100 MB during downloads
- ✅ No regressions in download/decrypt performance

### User Experience
- ✅ Intuitive layout with clear labels
- ✅ Helpful error messages
- ✅ Progress feedback during long operations
- ✅ Tooltips and help text where needed
- ✅ Consistent behavior across platforms

## Future Enhancements (Post-MVP)

### Short Term
1. Add firmware version history view (using removed list functionality)
2. Save common device profiles (model/region/IMEI combinations)
3. Dark mode support
4. Download queue (multiple files)
5. Checksum verification UI

### Long Term
1. Android app with mobile-optimized UI
2. Automatic update checks for susgo itself
3. Firmware changelog display
4. Multi-language support
5. Cloud backup integration for downloaded files
6. Batch download for multiple regions

## Appendix

### A. Fyne Resources
- Documentation: https://developer.fyne.io/
- Examples: https://github.com/fyne-io/examples
- API Reference: https://pkg.go.dev/fyne.io/fyne/v2

### B. Go Cross-Compilation
```bash
# Build for different platforms
GOOS=windows GOARCH=amd64 go build -o susgo.exe
GOOS=darwin GOARCH=amd64 go build -o susgo-mac-intel
GOOS=darwin GOARCH=arm64 go build -o susgo-mac-arm
GOOS=linux GOARCH=amd64 go build -o susgo-linux
```

### C. Fyne Packaging Commands
```bash
# Install fyne command
go install fyne.io/fyne/v2/cmd/fyne@latest

# Package for current platform
fyne package -icon icon.png

# Package for specific platforms
fyne package -os windows -icon icon.png
fyne package -os darwin -icon icon.png
fyne package -os linux -icon icon.png
fyne package -os android -icon icon.png -appID com.mattchengg.susgo
fyne package -os ios -icon icon.png -appID com.mattchengg.susgo
```

### D. System Dependencies (Linux)
Ubuntu/Debian:
```bash
sudo apt-get install gcc libgl1-mesa-dev xorg-dev
```

Fedora/RHEL:
```bash
sudo dnf install gcc mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel
```

Arch:
```bash
sudo pacman -S go gcc libxcursor libxrandr libxinerama libxi
```

### E. Code Style Guidelines
- Follow standard Go formatting (`gofmt`, `goimports`)
- Use meaningful variable names in GUI code
- Add comments for non-obvious UI logic
- Keep GUI code separate from core logic
- Use goroutines for long-running operations
- Always handle errors from Fyne operations

---

**Document Version**: 1.0  
**Last Updated**: 2025-01-21  
**Author**: Implementation Planner Agent  
**Status**: Ready for Implementation
