# Quick Start Guide - sfgo GUI Migration

This is a quick reference for implementing the GUI migration. See `plan.md` for detailed technical specifications and `task.md` for complete task breakdown.

## What We're Doing

1. **Removing** the "list" command from CLI
2. **Adding** a cross-platform GUI using Fyne framework
3. **Keeping** all core logic unchanged (auth, crypto, download, decrypt)

## File Structure After Migration

```
sfgo/
├── main.go              # GUI entry point (replaces CLI)
├── gui.go              # Optional: separate GUI code
├── progress.go         # Updated: ProgressReporter interface + implementations
├── auth.go             # No changes
├── crypt.go            # No changes
├── fusclient.go        # No changes
├── imei.go             # No changes
├── request.go          # No changes
├── versionfetch.go     # No changes
├── icon.png            # New: application icon
├── go.mod              # Updated: add Fyne dependency
├── plan.md             # This plan document
├── task.md             # Task breakdown
└── README.md           # Updated: GUI documentation
```

## Prerequisites

- Go 1.19+ (you have 1.25.4 ✅)
- GCC compiler (for Fyne)
- **Linux**: `sudo apt install gcc libgl1-mesa-dev xorg-dev`
- **Windows**: Install TDM-GCC or MinGW
- **macOS**: Xcode Command Line Tools

## Step-by-Step Implementation

### Step 1: Remove List Command (30 minutes)
```bash
# Edit main.go:
# 1. Delete parseListFlags() function
# 2. Delete listFirmware() function
# 3. Remove case "list": from switch statement
# 4. Remove global vars: latest, quiet
# 5. Update printUsage() to remove list documentation

# Test:
go build && ./sfgo -m SM-S928B -r EUX checkupdate
```

### Step 2: Install Fyne (10 minutes)
```bash
go get fyne.io/fyne/v2@latest
go install fyne.io/fyne/v2/cmd/fyne@latest
go mod tidy
```

### Step 3: Create Basic GUI (1 hour)
```bash
# Backup current main.go
cp main.go main_cli_backup.go

# Replace main.go with GUI version
# See task.md Task 3.2 for starter code
```

### Step 4: Implement Check Update Tab (1 hour)
```go
// Add makeCheckUpdateTab() function
// See task.md Task 3.3 for complete code
```

### Step 5: Add Progress Interface (30 minutes)
```go
// Edit progress.go
// Add ProgressReporter interface
// Add GUIProgressReporter implementation
// See task.md Task 4.1 for code
```

### Step 6: Implement Download Tab (2 hours)
```go
// Refactor download() to downloadFirmwareGUI()
// Create makeDownloadTab() function
// See task.md Tasks 4.2 and 4.3 for code
```

### Step 7: Implement Decrypt Tab (1.5 hours)
```go
// Refactor decrypt() to decryptFirmwareGUI()
// Create makeDecryptTab() function
// See task.md Tasks 5.1 and 5.2 for code
```

### Step 8: Test and Polish (2 hours)
```bash
# Test all functionality
# Add error handling
# Improve UI layout
# Add tooltips and help text
```

### Step 9: Build for Distribution (30 minutes)
```bash
# Linux
go build -o sfgo

# Windows (cross-compile or on Windows)
GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o sfgo.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o sfgo-mac

# Or use fyne package
fyne package -os windows -icon icon.png
fyne package -os darwin -icon icon.png
fyne package -os linux -icon icon.png
```

### Step 10: Update Documentation (1 hour)
```bash
# Update README.md with GUI instructions
# Add screenshots
# Update build instructions
# Create CHANGELOG.md
```

## GUI Design Overview

### Main Window
- **Size**: 700x500 pixels
- **Layout**: Tabbed interface
- **Tabs**: Check Update, Download, Decrypt

### Check Update Tab
```
┌─────────────────────────────────────┐
│ Check Latest Firmware Version       │
├─────────────────────────────────────┤
│ Model:  [SM-S928B            ]      │
│ Region: [EUX                 ]      │
│                                     │
│        [Check Update]               │
├─────────────────────────────────────┤
│ ✅ Latest Version: S928B...         │
└─────────────────────────────────────┘
```

### Download Tab
```
┌─────────────────────────────────────┐
│ Download Firmware                   │
├─────────────────────────────────────┤
│ Model:    [              ] *        │
│ Region:   [              ] *        │
│ IMEI/TAC: [              ] *        │
│ Serial:   [              ]          │
│ Version:  [              ]          │
│                                     │
│ [Select Output Directory]           │
│ 📁 /home/user/downloads             │
│                                     │
│        [Start Download]             │
├─────────────────────────────────────┤
│ [████████░░░░] 75.3%                │
│ 2.5 GB / 3.3 GB - 45 MB/s - ETA 1m │
└─────────────────────────────────────┘
```

### Decrypt Tab
```
┌─────────────────────────────────────┐
│ Decrypt Firmware                    │
├─────────────────────────────────────┤
│ Model:    [              ] *        │
│ Region:   [              ] *        │
│ IMEI:     [              ] *        │
│ Version:  [              ] *        │
│                                     │
│ Encryption: ○ V2  ● V4              │
│                                     │
│ [Select Input File]                 │
│ 📄 firmware.enc4                    │
│                                     │
│ [Select Output File]                │
│ 📄 firmware.zip                     │
│                                     │
│        [Start Decrypt]              │
├─────────────────────────────────────┤
│ 🔓 Decrypting... 45%                │
└─────────────────────────────────────┘
```

## Key Code Patterns

### Running Operations in Goroutines
```go
button := widget.NewButton("Action", func() {
    // Validate inputs first
    if model == "" {
        statusLabel.SetText("❌ Error: Model required")
        return
    }
    
    // Disable button
    button.Disable()
    statusLabel.SetText("⏳ Working...")
    
    // Run operation in goroutine
    go func() {
        defer button.Enable() // Re-enable when done
        
        err := doSomethingLong(model, region)
        if err != nil {
            statusLabel.SetText("❌ Error: " + err.Error())
        } else {
            statusLabel.SetText("✅ Success!")
        }
    }()
})
```

### Progress Reporting
```go
progress := NewGUIProgressReporter(progressBar, statusLabel)
progress.SetTotal(totalSize)

// In your operation:
progress.Add(bytesRead)
progress.SetStatus("Downloading...")

// When done:
progress.Finish()
```

### File Dialogs
```go
// Folder selection
dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
    if err == nil && uri != nil {
        outputDir = uri.Path()
        label.SetText("📁 " + outputDir)
    }
}, window)

// File open
dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
    if err == nil && uri != nil {
        inputFile = uri.URI().Path()
        uri.Close()
    }
}, window)

// File save
dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
    if err == nil && uri != nil {
        outputFile = uri.URI().Path()
        uri.Close()
    }
}, window)
```

## Common Issues and Solutions

### Issue: "cannot find package fyne.io/fyne/v2"
**Solution**: 
```bash
go get fyne.io/fyne/v2@latest
go mod tidy
```

### Issue: "gcc: command not found" (Linux)
**Solution**:
```bash
sudo apt install gcc libgl1-mesa-dev xorg-dev
```

### Issue: UI freezes during download
**Solution**: Always run long operations in goroutines
```go
go func() {
    // Your long operation here
}()
```

### Issue: Progress bar doesn't update
**Solution**: Make sure you're calling `Add()` or `SetCurrent()` from the goroutine, Fyne will handle UI updates automatically

### Issue: Window too small/big
**Solution**: Adjust in main():
```go
myWindow.Resize(fyne.NewSize(800, 600))
```

## Testing Checklist

Quick tests to run after implementation:

- [ ] Application launches without errors
- [ ] Check Update returns version for valid model/region
- [ ] Check Update shows error for invalid model/region
- [ ] Download: File dialog opens and allows selection
- [ ] Download: Progress bar updates during download
- [ ] Download: Error shown if inputs invalid
- [ ] Decrypt: Input and output file dialogs work
- [ ] Decrypt: Encryption version selection works
- [ ] All tabs accessible via tab bar
- [ ] Window can be resized
- [ ] Application can be closed normally

## Resources

- **Fyne Documentation**: https://developer.fyne.io/
- **Fyne Examples**: https://github.com/fyne-io/examples
- **Fyne API Reference**: https://pkg.go.dev/fyne.io/fyne/v2
- **This Project Plan**: See `plan.md` for detailed architecture
- **Task List**: See `task.md` for step-by-step tasks

## Time Estimates

- **Experienced Go + GUI dev**: 20-25 hours
- **Experienced Go, new to Fyne**: 30-35 hours  
- **Learning as you go**: 40-50 hours

## Getting Help

1. Check `plan.md` for technical details
2. Check `task.md` for specific implementation steps
3. Review Fyne examples: https://github.com/fyne-io/examples
4. Fyne community Discord: https://discord.gg/fyne

## Next Steps

1. ✅ Read this document
2. ⏭️ Read detailed `plan.md`
3. ⏭️ Follow tasks in `task.md` in order
4. ⏭️ Start with Phase 1 (Remove list command)
5. ⏭️ Then Phase 2 (Setup Fyne)
6. ⏭️ Continue through phases sequentially

Good luck! 🚀
