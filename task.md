# susgo GUI Migration - Task List

## Overview
This document contains specific, actionable tasks for migrating susgo from CLI to GUI. Tasks are organized by phase and dependency order.

---

## Phase 1: Remove List Command ⚠️ PREREQUISITE

### Task 1.1: Remove list command functions from main.go
**File**: `main.go`  
**Priority**: High  
**Estimated Time**: 15 minutes  
**Dependencies**: None

**Actions**:
1. Delete function `parseListFlags()` (lines 100-105)
2. Delete function `listFirmware()` (lines 142-173)
3. Verify no other code references these functions

**Verification**:
```bash
go build
# Should compile successfully
```

---

### Task 1.2: Remove list case from main switch statement
**File**: `main.go`  
**Priority**: High  
**Estimated Time**: 5 minutes  
**Dependencies**: Task 1.1

**Actions**:
1. In `main()` function, locate switch statement (line 42)
2. Remove entire `case "list":` block (lines 45-47)
3. Keep checkupdate, download, decrypt cases

**Verification**:
```bash
go run . -m SM-S928B -r EUX list
# Should output: "Unknown command: list"
```

---

### Task 1.3: Remove unused global variables
**File**: `main.go`  
**Priority**: Medium  
**Estimated Time**: 5 minutes  
**Dependencies**: Task 1.1, 1.2

**Actions**:
1. Remove global variables (lines 14-27):
   - `latest bool`
   - `quiet bool`
2. Keep all other global variables (needed by remaining commands)

**Verification**:
```bash
go build
# Should compile without unused variable warnings
```

---

### Task 1.4: Update printUsage() documentation
**File**: `main.go`  
**Priority**: Medium  
**Estimated Time**: 10 minutes  
**Dependencies**: Task 1.1, 1.2, 1.3

**Actions**:
1. In `printUsage()` function (lines 61-98):
   - Remove line: `susgo -m <model> -r <region> list [-l] [-q]`
   - Remove from Commands section: "list         List all available firmware versions"
   - Remove entire section: "List Options:"
2. Clean up spacing for consistency

**Verification**:
```bash
go run .
# Should display help without list command
```

---

### Task 1.5: Test remaining CLI functionality
**File**: N/A (Testing)  
**Priority**: High  
**Estimated Time**: 15 minutes  
**Dependencies**: All Phase 1 tasks

**Actions**:
1. Test checkupdate command:
   ```bash
   go run . -m SM-S928B -r EUX checkupdate
   ```
2. Test download command (with fake IMEI):
   ```bash
   go run . -m SM-S928B -r EUX -i 35000000 download -O /tmp
   ```
3. Verify error handling still works
4. Verify help text displays correctly

**Success Criteria**:
- All remaining commands work as before
- No references to list command remain
- Code compiles without warnings

---

## Phase 2: Setup Fyne and Project Structure

### Task 2.1: Add Fyne dependency
**File**: `go.mod`  
**Priority**: High  
**Estimated Time**: 10 minutes  
**Dependencies**: Phase 1 complete

**Actions**:
1. Run: `go get fyne.io/fyne/v2@latest`
2. Run: `go mod tidy`
3. Verify go.mod includes Fyne v2.4.5+

**Verification**:
```bash
cat go.mod | grep fyne
# Should show: fyne.io/fyne/v2 v2.x.x
```

---

### Task 2.2: Install Fyne CLI tool (optional but recommended)
**File**: N/A (System installation)  
**Priority**: Medium  
**Estimated Time**: 5 minutes  
**Dependencies**: Task 2.1

**Actions**:
1. Run: `go install fyne.io/fyne/v2/cmd/fyne@latest`
2. Verify installation: `fyne version`

**Verification**:
```bash
which fyne
fyne version
```

---

### Task 2.3: Create application icon
**File**: `icon.png`  
**Priority**: Low  
**Estimated Time**: 30 minutes (or use placeholder)  
**Dependencies**: None

**Actions**:
1. Create or obtain 512x512 PNG icon
2. Save as `icon.png` in project root
3. Alternatively, use Fyne default icon temporarily

**Verification**:
```bash
file icon.png
# Should show: PNG image data, 512 x 512
```

---

## Phase 3: Create Basic GUI Structure

### Task 3.1: Backup current main.go
**File**: N/A (File management)  
**Priority**: High  
**Estimated Time**: 2 minutes  
**Dependencies**: Phase 1 complete

**Actions**:
1. Copy main.go to main_cli_backup.go
2. Keep for reference or future CLI mode

**Verification**:
```bash
ls -la main*.go
# Should show both main.go and main_cli_backup.go
```

---

### Task 3.2: Create new GUI main.go skeleton
**File**: `main.go` (replace content)  
**Priority**: High  
**Estimated Time**: 20 minutes  
**Dependencies**: Task 2.1, 3.1

**Actions**:
1. Replace main.go content with GUI entry point
2. Import Fyne packages:
   ```go
   import (
       "fyne.io/fyne/v2/app"
       "fyne.io/fyne/v2/container"
       "fyne.io/fyne/v2/widget"
   )
   ```
3. Create basic main() function:
   ```go
   func main() {
       myApp := app.New()
       myWindow := myApp.NewWindow("susgo - Samsung Firmware Downloader")
       
       // Placeholder content
       content := widget.NewLabel("susgo GUI - Coming Soon")
       
       myWindow.SetContent(content)
       myWindow.Resize(fyne.NewSize(700, 500))
       myWindow.ShowAndRun()
   }
   ```

**Verification**:
```bash
go run .
# Should open a window with "susgo GUI - Coming Soon"
```

---

### Task 3.3: Create makeCheckUpdateTab() function
**File**: `main.go`  
**Priority**: High  
**Estimated Time**: 45 minutes  
**Dependencies**: Task 3.2

**Actions**:
1. Add function above main():
   ```go
   func makeCheckUpdateTab() *fyne.Container {
       // Create input fields
       modelEntry := widget.NewEntry()
       modelEntry.SetPlaceHolder("e.g., SM-S928B")
       
       regionEntry := widget.NewEntry()
       regionEntry.SetPlaceHolder("e.g., EUX")
       
       resultLabel := widget.NewLabel("")
       resultLabel.Wrapping = fyne.TextWrapWord
       
       // Create button
       checkButton := widget.NewButton("Check Update", func() {
           model := strings.TrimSpace(modelEntry.Text)
           region := strings.TrimSpace(regionEntry.Text)
           
           // Validation
           if model == "" || region == "" {
               resultLabel.SetText("❌ Error: Model and Region are required")
               return
           }
           
           resultLabel.SetText("⏳ Checking...")
           
           // Run in goroutine to keep UI responsive
           go func() {
               version, err := getLatestVersion(model, region)
               if err != nil {
                   resultLabel.SetText("❌ Error: " + err.Error())
               } else {
                   resultLabel.SetText("✅ Latest Version: " + version)
               }
           }()
       })
       
       // Layout
       form := widget.NewForm(
           widget.NewFormItem("Model", modelEntry),
           widget.NewFormItem("Region", regionEntry),
       )
       
       return container.NewVBox(
           widget.NewLabelWithStyle("Check Latest Firmware Version",
               fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
           widget.NewSeparator(),
           form,
           checkButton,
           widget.NewSeparator(),
           resultLabel,
       )
   }
   ```

2. Update main() to use tabs:
   ```go
   tabs := container.NewAppTabs(
       container.NewTabItem("Check Update", makeCheckUpdateTab()),
   )
   myWindow.SetContent(tabs)
   ```

**Verification**:
```bash
go run .
# Should show tab with form, test with real model/region
```

---

## Phase 4: Implement Download Tab

### Task 4.1: Create ProgressReporter interface
**File**: `progress.go`  
**Priority**: High  
**Estimated Time**: 30 minutes  
**Dependencies**: Task 2.1

**Actions**:
1. Add to top of progress.go:
   ```go
   type ProgressReporter interface {
       SetTotal(total int64)
       SetCurrent(current int64)
       Add(delta int64)
       SetStatus(message string)
       Finish()
   }
   ```

2. Make existing ProgressBar implement interface (already has most methods)

3. Create GUIProgressReporter:
   ```go
   type GUIProgressReporter struct {
       bar       *widget.ProgressBar
       label     *widget.Label
       total     int64
       current   int64
       startTime time.Time
   }
   
   func NewGUIProgressReporter(bar *widget.ProgressBar, label *widget.Label) *GUIProgressReporter {
       return &GUIProgressReporter{
           bar:       bar,
           label:     label,
           startTime: time.Now(),
       }
   }
   
   func (g *GUIProgressReporter) SetTotal(total int64) {
       g.total = total
       g.bar.Max = float64(total)
   }
   
   func (g *GUIProgressReporter) SetCurrent(current int64) {
       g.current = current
       g.bar.SetValue(float64(current))
       g.updateLabel()
   }
   
   func (g *GUIProgressReporter) Add(delta int64) {
       g.current += delta
       g.bar.SetValue(float64(g.current))
       g.updateLabel()
   }
   
   func (g *GUIProgressReporter) SetStatus(message string) {
       g.label.SetText(message)
   }
   
   func (g *GUIProgressReporter) Finish() {
       g.bar.SetValue(float64(g.total))
       g.label.SetText("✅ Complete!")
   }
   
   func (g *GUIProgressReporter) updateLabel() {
       if g.total <= 0 {
           return
       }
       pct := float64(g.current) / float64(g.total) * 100
       elapsed := time.Since(g.startTime).Seconds()
       speed := float64(g.current) / elapsed
       
       var eta string
       if speed > 0 {
           remaining := float64(g.total-g.current) / speed
           if remaining < 60 {
               eta = fmt.Sprintf("%.0fs", remaining)
           } else {
               eta = fmt.Sprintf("%.0fm", remaining/60)
           }
       } else {
           eta = "calculating..."
       }
       
       g.label.SetText(fmt.Sprintf("%.1f%% - %s/%s - %s/s - ETA: %s",
           pct,
           formatSize(g.current),
           formatSize(g.total),
           formatSize(int64(speed)),
           eta))
   }
   ```

**Verification**:
```bash
go build
# Should compile successfully
```

---

### Task 4.2: Refactor download() to be GUI-callable
**File**: `main.go` (or create download.go)  
**Priority**: High  
**Estimated Time**: 60 minutes  
**Dependencies**: Task 4.1

**Actions**:
1. Create new function that doesn't use global variables:
   ```go
   func downloadFirmwareGUI(model, region, imei, serial, version, outputDir string,
                            showMD5 bool, progress ProgressReporter) error {
       // Move logic from existing download() function
       // Replace global variable access with parameters
       // Replace fmt.Printf with progress.SetStatus()
       // Replace os.Exit() with return err
       // Replace ProgressBar with ProgressReporter interface
       
       // Key changes:
       // - Don't call os.Exit on errors
       // - Use progress.SetStatus() for messages
       // - Return errors instead of printing
       
       effectiveIMEI, err := parseIMEIFromValues(imei, serial)
       if err != nil {
           return err
       }
       
       client := NewFUSClient()
       
       if version == "" {
           ver, err := getLatestVersion(model, region)
           if err != nil {
               return fmt.Errorf("getting version: %w", err)
           }
           version = ver
       }
       
       path, filename, size, err := getBinaryFile(client, version, model, region, effectiveIMEI)
       if err != nil {
           return fmt.Errorf("getting binary info: %w", err)
       }
       
       progress.SetTotal(size)
       progress.SetStatus(fmt.Sprintf("Downloading %s (%.2f GB)",
           filename, float64(size)/(1024*1024*1024)))
       
       // ... continue with download logic
       // Use progress.Add() in read loop
       
       return nil
   }
   
   // Helper function
   func parseIMEIFromValues(imei, serial string) (string, error) {
       if imei != "" {
           switch len(imei) {
           case 8:
               return validateAndGenerateIMEI(imei, "", "")
           case 15:
               return imei, nil
           default:
               return "", fmt.Errorf("IMEI must be 8 or 15 digits")
           }
       }
       if serial != "" {
           return serial, nil
       }
       return "", fmt.Errorf("IMEI or Serial required")
   }
   ```

2. Keep old download() function for now (marked as deprecated)

**Verification**:
```bash
go build
# Should compile successfully
```

---

### Task 4.3: Create makeDownloadTab() function
**File**: `main.go`  
**Priority**: High  
**Estimated Time**: 90 minutes  
**Dependencies**: Task 4.2

**Actions**:
1. Create comprehensive download tab:
   ```go
   func makeDownloadTab() *fyne.Container {
       // Input fields
       modelEntry := widget.NewEntry()
       modelEntry.SetPlaceHolder("e.g., SM-S928B")
       
       regionEntry := widget.NewEntry()
       regionEntry.SetPlaceHolder("e.g., EUX")
       
       imeiEntry := widget.NewEntry()
       imeiEntry.SetPlaceHolder("15 digits or 8 digit TAC")
       
       serialEntry := widget.NewEntry()
       serialEntry.SetPlaceHolder("Optional, if no IMEI")
       
       versionEntry := widget.NewEntry()
       versionEntry.SetPlaceHolder("Leave empty for latest")
       
       // Output directory selection
       var outputDir string
       outputLabel := widget.NewLabel("No directory selected")
       selectDirButton := widget.NewButton("Select Output Directory", func() {
           dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
               if err == nil && uri != nil {
                   outputDir = uri.Path()
                   outputLabel.SetText("📁 " + outputDir)
               }
           }, myWindow)
       })
       
       // Progress widgets
       progressBar := widget.NewProgressBar()
       progressBar.Hide()
       
       statusLabel := widget.NewLabel("")
       statusLabel.Wrapping = fyne.TextWrapWord
       
       // Download button
       var downloadInProgress bool
       downloadButton := widget.NewButton("Start Download", func() {
           if downloadInProgress {
               statusLabel.SetText("⚠️ Download already in progress")
               return
           }
           
           // Validation
           model := strings.TrimSpace(modelEntry.Text)
           region := strings.TrimSpace(regionEntry.Text)
           imei := strings.TrimSpace(imeiEntry.Text)
           serial := strings.TrimSpace(serialEntry.Text)
           version := strings.TrimSpace(versionEntry.Text)
           
           if model == "" || region == "" {
               statusLabel.SetText("❌ Model and Region are required")
               return
           }
           
           if imei == "" && serial == "" {
               statusLabel.SetText("❌ IMEI or Serial is required")
               return
           }
           
           if outputDir == "" {
               statusLabel.SetText("❌ Please select an output directory")
               return
           }
           
           // Start download
           downloadInProgress = true
           downloadButton.Disable()
           progressBar.Show()
           statusLabel.SetText("⏳ Initializing...")
           
           progress := NewGUIProgressReporter(progressBar, statusLabel)
           
           go func() {
               defer func() {
                   downloadInProgress = false
                   downloadButton.Enable()
               }()
               
               err := downloadFirmwareGUI(model, region, imei, serial, version,
                   outputDir, false, progress)
               
               if err != nil {
                   statusLabel.SetText("❌ Error: " + err.Error())
               } else {
                   progress.Finish()
               }
           }()
       })
       
       // Layout
       form := widget.NewForm(
           widget.NewFormItem("Model *", modelEntry),
           widget.NewFormItem("Region *", regionEntry),
           widget.NewFormItem("IMEI/TAC *", imeiEntry),
           widget.NewFormItem("Serial", serialEntry),
           widget.NewFormItem("Version", versionEntry),
       )
       
       return container.NewVBox(
           widget.NewLabelWithStyle("Download Firmware",
               fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
           widget.NewSeparator(),
           form,
           selectDirButton,
           outputLabel,
           downloadButton,
           widget.NewSeparator(),
           progressBar,
           statusLabel,
       )
   }
   ```

2. Add to main():
   ```go
   tabs := container.NewAppTabs(
       container.NewTabItem("Check Update", makeCheckUpdateTab()),
       container.NewTabItem("Download", makeDownloadTab()),
   )
   ```

**Verification**:
```bash
go run .
# Test download with real firmware (or at least initialization)
```

---

## Phase 5: Implement Decrypt Tab

### Task 5.1: Refactor decrypt() to be GUI-callable
**File**: `main.go`  
**Priority**: High  
**Estimated Time**: 45 minutes  
**Dependencies**: Task 4.1

**Actions**:
1. Create new function:
   ```go
   func decryptFirmwareGUI(model, region, imei, serial, version string,
                           inFile, outFile string, encVer int,
                           progress ProgressReporter) error {
       // Validate inputs
       if version == "" || inFile == "" || outFile == "" {
           return fmt.Errorf("version, input file, and output file are required")
       }
       
       effectiveIMEI, err := parseIMEIFromValues(imei, serial)
       if err != nil {
           return fmt.Errorf("IMEI/Serial error: %w", err)
       }
       
       progress.SetStatus("⏳ Generating decryption key...")
       
       // Get key
       var key []byte
       if encVer == 2 {
           key = getV2Key(version, model, region)
       } else {
           var err error
           key, err = getV4Key(version, model, region, effectiveIMEI)
           if err != nil {
               return fmt.Errorf("key generation: %w", err)
           }
       }
       
       progress.SetStatus("🔓 Decrypting firmware...")
       
       // Decrypt with progress reporting
       if err := decryptFirmwareWithProgress(inFile, outFile, key, progress); err != nil {
           return fmt.Errorf("decryption: %w", err)
       }
       
       progress.Finish()
       progress.SetStatus("✅ Decryption complete!")
       
       return nil
   }
   ```

2. Modify decryptFirmware() in crypt.go to accept progress reporter:
   ```go
   func decryptFirmwareWithProgress(inFile, outFile string, key []byte,
                                    progress ProgressReporter) error {
       // Add progress reporting similar to current implementation
       // Call progress.SetCurrent() periodically
   }
   ```

**Verification**:
```bash
go build
# Should compile successfully
```

---

### Task 5.2: Create makeDecryptTab() function
**File**: `main.go`  
**Priority**: High  
**Estimated Time**: 60 minutes  
**Dependencies**: Task 5.1

**Actions**:
1. Create decrypt tab:
   ```go
   func makeDecryptTab() *fyne.Container {
       // Input fields
       modelEntry := widget.NewEntry()
       modelEntry.SetPlaceHolder("e.g., SM-S928B")
       
       regionEntry := widget.NewEntry()
       regionEntry.SetPlaceHolder("e.g., EUX")
       
       imeiEntry := widget.NewEntry()
       imeiEntry.SetPlaceHolder("15 digits or 8 digit TAC")
       
       serialEntry := widget.NewEntry()
       serialEntry.SetPlaceHolder("Optional, if no IMEI")
       
       versionEntry := widget.NewEntry()
       versionEntry.SetPlaceHolder("e.g., S928BXXU1AXK1/S928BOXM1AXK1/...")
       
       // File selection
       var inputFile, outputFile string
       inputLabel := widget.NewLabel("No input file selected")
       outputLabel := widget.NewLabel("No output file selected")
       
       selectInputButton := widget.NewButton("Select Input File", func() {
           dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
               if err == nil && uri != nil {
                   inputFile = uri.URI().Path()
                   inputLabel.SetText("📄 " + filepath.Base(inputFile))
                   uri.Close()
               }
           }, myWindow)
       })
       
       selectOutputButton := widget.NewButton("Select Output File", func() {
           dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
               if err == nil && uri != nil {
                   outputFile = uri.URI().Path()
                   outputLabel.SetText("📄 " + filepath.Base(outputFile))
                   uri.Close()
               }
           }, myWindow)
       })
       
       // Encryption version selection
       encVerRadio := widget.NewRadioGroup([]string{"V2", "V4 (default)"}, nil)
       encVerRadio.SetSelected("V4 (default)")
       
       // Progress widgets
       progressBar := widget.NewProgressBar()
       progressBar.Hide()
       
       statusLabel := widget.NewLabel("")
       statusLabel.Wrapping = fyne.TextWrapWord
       
       // Decrypt button
       var decryptInProgress bool
       decryptButton := widget.NewButton("Start Decrypt", func() {
           if decryptInProgress {
               statusLabel.SetText("⚠️ Decryption already in progress")
               return
           }
           
           // Validation
           model := strings.TrimSpace(modelEntry.Text)
           region := strings.TrimSpace(regionEntry.Text)
           imei := strings.TrimSpace(imeiEntry.Text)
           serial := strings.TrimSpace(serialEntry.Text)
           version := strings.TrimSpace(versionEntry.Text)
           
           if model == "" || region == "" {
               statusLabel.SetText("❌ Model and Region are required")
               return
           }
           
           if imei == "" && serial == "" {
               statusLabel.SetText("❌ IMEI or Serial is required")
               return
           }
           
           if version == "" {
               statusLabel.SetText("❌ Version is required")
               return
           }
           
           if inputFile == "" || outputFile == "" {
               statusLabel.SetText("❌ Please select input and output files")
               return
           }
           
           encVer := 4
           if encVerRadio.Selected == "V2" {
               encVer = 2
           }
           
           // Start decryption
           decryptInProgress = true
           decryptButton.Disable()
           progressBar.Show()
           statusLabel.SetText("⏳ Starting decryption...")
           
           progress := NewGUIProgressReporter(progressBar, statusLabel)
           
           go func() {
               defer func() {
                   decryptInProgress = false
                   decryptButton.Enable()
               }()
               
               err := decryptFirmwareGUI(model, region, imei, serial, version,
                   inputFile, outputFile, encVer, progress)
               
               if err != nil {
                   statusLabel.SetText("❌ Error: " + err.Error())
               }
           }()
       })
       
       // Layout
       form := widget.NewForm(
           widget.NewFormItem("Model *", modelEntry),
           widget.NewFormItem("Region *", regionEntry),
           widget.NewFormItem("IMEI/TAC *", imeiEntry),
           widget.NewFormItem("Serial", serialEntry),
           widget.NewFormItem("Version *", versionEntry),
       )
       
       return container.NewVBox(
           widget.NewLabelWithStyle("Decrypt Firmware",
               fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
           widget.NewSeparator(),
           form,
           widget.NewLabel("Encryption Version:"),
           encVerRadio,
           widget.NewSeparator(),
           selectInputButton,
           inputLabel,
           selectOutputButton,
           outputLabel,
           decryptButton,
           widget.NewSeparator(),
           progressBar,
           statusLabel,
       )
   }
   ```

2. Add to main():
   ```go
   tabs := container.NewAppTabs(
       container.NewTabItem("Check Update", makeCheckUpdateTab()),
       container.NewTabItem("Download", makeDownloadTab()),
       container.NewTabItem("Decrypt", makeDecryptTab()),
   )
   ```

**Verification**:
```bash
go run .
# Test all three tabs are visible and functional
```

---

## Phase 6: Polish and Error Handling

### Task 6.1: Add missing imports
**File**: `main.go`  
**Priority**: High  
**Estimated Time**: 10 minutes  
**Dependencies**: Phase 3, 4, 5

**Actions**:
1. Ensure all imports are present:
   ```go
   import (
       "fmt"
       "path/filepath"
       "strings"
       "time"
       
       "fyne.io/fyne/v2"
       "fyne.io/fyne/v2/app"
       "fyne.io/fyne/v2/container"
       "fyne.io/fyne/v2/dialog"
       "fyne.io/fyne/v2/widget"
   )
   ```

2. Run `goimports` or `gofmt` to organize

**Verification**:
```bash
go build
# Should compile without import errors
```

---

### Task 6.2: Add input validation helpers
**File**: `main.go`  
**Priority**: Medium  
**Estimated Time**: 20 minutes  
**Dependencies**: Phase 5 complete

**Actions**:
1. Create validation helper functions:
   ```go
   func validateModel(model string) error {
       if model == "" {
           return fmt.Errorf("model is required")
       }
       if len(model) < 5 {
           return fmt.Errorf("model seems too short")
       }
       return nil
   }
   
   func validateRegion(region string) error {
       if region == "" {
           return fmt.Errorf("region is required")
       }
       if len(region) != 3 {
           return fmt.Errorf("region should be 3 characters (e.g., EUX)")
       }
       return nil
   }
   
   func validateIMEIInput(imei string) error {
       if imei == "" {
           return nil // Optional
       }
       if len(imei) != 8 && len(imei) != 15 {
           return fmt.Errorf("IMEI must be 8 or 15 digits")
       }
       return nil
   }
   ```

2. Use these in button click handlers

**Verification**:
```bash
# Test with invalid inputs in GUI
```

---

### Task 6.3: Add dialog confirmations
**File**: `main.go`  
**Priority**: Low  
**Estimated Time**: 15 minutes  
**Dependencies**: Task 6.2

**Actions**:
1. Add confirmation before large downloads:
   ```go
   // In download button handler, after getting size:
   if size > 5*1024*1024*1024 { // > 5 GB
       dialog.ShowConfirm("Large Download",
           fmt.Sprintf("This firmware is %.2f GB. Continue?",
               float64(size)/(1024*1024*1024)),
           func(confirm bool) {
               if confirm {
                   // Proceed with download
               }
           }, myWindow)
   }
   ```

**Verification**:
```bash
# Test with large firmware download
```

---

### Task 6.4: Improve error messages
**File**: `main.go`  
**Priority**: Medium  
**Estimated Time**: 20 minutes  
**Dependencies**: Task 6.2

**Actions**:
1. Create user-friendly error message function:
   ```go
   func formatErrorForUser(err error) string {
       errStr := err.Error()
       
       // Map common errors to user-friendly messages
       if strings.Contains(errStr, "model or region not found") {
           return "The model or region code was not found. Please verify:\n" +
                  "- Model format (e.g., SM-S928B)\n" +
                  "- Region code is valid (e.g., EUX, XAR)"
       }
       
       if strings.Contains(errStr, "no firmware available") {
           return "No firmware is available for this model/region combination."
       }
       
       if strings.Contains(errStr, "unable to find a valid IMEI") {
           return "Could not generate a valid IMEI from the TAC code provided.\n" +
                  "Try using a full 15-digit IMEI instead."
       }
       
       // Default: return original error
       return errStr
   }
   ```

2. Use in all error displays:
   ```go
   statusLabel.SetText("❌ " + formatErrorForUser(err))
   ```

**Verification**:
```bash
# Test with various error conditions
```

---

### Task 6.5: Add tooltips and help text
**File**: `main.go`  
**Priority**: Low  
**Estimated Time**: 30 minutes  
**Dependencies**: Phase 5 complete

**Actions**:
1. Add tooltips using NewLabelWithStyle or help icons
2. Add explanatory text for complex fields:
   ```go
   // Example for IMEI field:
   imeiHelp := widget.NewLabel("💡 Enter 15-digit IMEI or 8-digit TAC")
   imeiHelp.TextStyle = fyne.TextStyle{Italic: true}
   ```

3. Add tooltips for buttons (requires custom widget or icon)

**Verification**:
```bash
# Visually inspect all tabs
```

---

## Phase 7: Cross-Platform Testing and Building

### Task 7.1: Test on Linux
**File**: N/A (Testing)  
**Priority**: High  
**Estimated Time**: 30 minutes  
**Dependencies**: Phase 6 complete

**Actions**:
1. Build: `go build -o susgo`
2. Test all tabs
3. Test file dialogs
4. Check window rendering
5. Test with actual firmware operations

**Success Criteria**:
- Application runs without crashes
- All functionality works
- File dialogs work correctly
- No visual glitches

---

### Task 7.2: Test on Windows (if available)
**File**: N/A (Testing)  
**Priority**: Medium  
**Estimated Time**: 30 minutes  
**Dependencies**: Phase 6 complete

**Actions**:
1. Build: `go build -o susgo.exe`
2. Or with hidden console: `go build -ldflags="-H windowsgui" -o susgo.exe`
3. Test all functionality
4. Verify file paths work correctly (backslashes)

**Success Criteria**:
- Same as Task 7.1

---

### Task 7.3: Test on macOS (if available)
**File**: N/A (Testing)  
**Priority**: Medium  
**Estimated Time**: 30 minutes  
**Dependencies**: Phase 6 complete

**Actions**:
1. Build: `go build -o susgo`
2. Or create .app: `fyne package -os darwin -icon icon.png`
3. Test all functionality
4. Test on both Intel and Apple Silicon if possible

**Success Criteria**:
- Same as Task 7.1

---

### Task 7.4: Create build script
**File**: `build.sh` or `build.bat`  
**Priority**: Low  
**Estimated Time**: 20 minutes  
**Dependencies**: Task 7.1, 7.2, 7.3

**Actions**:
1. Create `build.sh` for Linux/Mac:
   ```bash
   #!/bin/bash
   echo "Building susgo..."
   go build -o susgo
   echo "Build complete: susgo"
   
   if command -v fyne &> /dev/null; then
       echo "Creating packaged version..."
       fyne package -os darwin -icon icon.png
       fyne package -os linux -icon icon.png
   fi
   ```

2. Create `build.bat` for Windows:
   ```batch
   @echo off
   echo Building susgo...
   go build -ldflags="-H windowsgui" -o susgo.exe
   echo Build complete: susgo.exe
   ```

3. Make executable: `chmod +x build.sh`

**Verification**:
```bash
./build.sh
# Should produce binary
```

---

### Task 7.5: Package for distribution
**File**: N/A (Build process)  
**Priority**: Low  
**Estimated Time**: 30 minutes  
**Dependencies**: Task 7.4

**Actions**:
1. Package for Windows:
   ```bash
   fyne package -os windows -icon icon.png
   ```

2. Package for macOS:
   ```bash
   fyne package -os darwin -icon icon.png
   ```

3. Package for Linux:
   ```bash
   fyne package -os linux -icon icon.png
   ```

4. Test packaged versions

**Deliverables**:
- susgo.exe (Windows)
- susgo.app (macOS)
- susgo.tar.xz (Linux)

---

## Phase 8: Documentation

### Task 8.1: Update README.md
**File**: `README.md`  
**Priority**: High  
**Estimated Time**: 45 minutes  
**Dependencies**: Phase 7 complete

**Actions**:
1. Update description to mention GUI
2. Add screenshots (take screenshots of each tab)
3. Update installation instructions:
   - Add system dependencies for building (Linux)
   - Add Fyne installation instructions
4. Replace usage section with GUI workflow
5. Add troubleshooting section
6. Update build instructions

**Template**:
```markdown
# susgo - Samsung Firmware Downloader

GUI application for downloading Samsung firmware directly from Samsung's servers.

## Features
- ✅ Check latest firmware version
- ⬇️ Download firmware with resume support
- 🔓 Decrypt firmware files (V2 and V4 encryption)
- 🖥️ Cross-platform GUI (Windows, macOS, Linux)
- 📊 Real-time download progress

## Screenshots
[Add screenshots here]

## Installation

### Pre-built Binaries
Download from [Releases](link)

### Building from Source

#### Prerequisites
- Go 1.19 or later
- GCC (for Fyne)
- Platform-specific requirements:
  - **Linux**: `sudo apt install gcc libgl1-mesa-dev xorg-dev`
  - **Windows**: Install [TDM-GCC](link)
  - **macOS**: Install Xcode Command Line Tools

#### Build
```bash
git clone https://github.com/mattchengg/susgo.git
cd susgo
go get
go build
```

## Usage

1. **Check Update**: Enter model and region, click "Check Update"
2. **Download**: Fill in firmware details, select output folder, click "Start Download"
3. **Decrypt**: Select encrypted file, provide details, click "Start Decrypt"

## Common Model/Region Examples
- Galaxy S24 Ultra (Europe): SM-S928B / EUX
- Galaxy S24 Ultra (USA): SM-S928U / XAA
...

## Troubleshooting
...
```

**Verification**:
```bash
# Review README for completeness
```

---

### Task 8.2: Create BUILDING.md
**File**: `BUILDING.md`  
**Priority**: Medium  
**Estimated Time**: 30 minutes  
**Dependencies**: Task 8.1

**Actions**:
1. Create detailed build guide for each platform
2. Include troubleshooting for common build errors
3. Document cross-compilation

**Content**:
- System dependencies per platform
- Step-by-step build instructions
- Packaging instructions
- Cross-compilation examples

---

### Task 8.3: Add code comments
**File**: `main.go`, `progress.go`  
**Priority**: Medium  
**Estimated Time**: 30 minutes  
**Dependencies**: Phase 6 complete

**Actions**:
1. Add godoc comments to all exported functions
2. Add inline comments for complex logic
3. Document the ProgressReporter interface

**Example**:
```go
// ProgressReporter provides a uniform interface for reporting progress
// during long-running operations like downloads and decryption.
// Implementations exist for both terminal and GUI display.
type ProgressReporter interface {
    // SetTotal sets the total size/count for the operation
    SetTotal(total int64)
    
    // SetCurrent updates the current progress value
    SetCurrent(current int64)
    
    // Add increments the progress by the given delta
    Add(delta int64)
    
    // SetStatus updates the status message
    SetStatus(message string)
    
    // Finish marks the operation as complete
    Finish()
}
```

---

### Task 8.4: Take screenshots
**File**: Screenshots (add to repo or README)  
**Priority**: Medium  
**Estimated Time**: 15 minutes  
**Dependencies**: Phase 7 complete

**Actions**:
1. Take screenshots of each tab:
   - Check Update tab (with result shown)
   - Download tab (showing progress)
   - Decrypt tab (showing file selection)
2. Save as PNG files
3. Add to README or docs folder
4. Update README to reference screenshots

**Verification**:
```bash
# Screenshots look professional and clear
```

---

### Task 8.5: Create CHANGELOG.md
**File**: `CHANGELOG.md`  
**Priority**: Low  
**Estimated Time**: 15 minutes  
**Dependencies**: Phase 8 complete

**Actions**:
1. Create changelog following Keep a Changelog format
2. Document major changes from CLI to GUI
3. Note removed features (list command)

**Template**:
```markdown
# Changelog

## [2.0.0] - 2025-01-XX

### Added
- 🎨 Cross-platform GUI using Fyne framework
- 📊 Visual progress bars for downloads and decryption
- 📁 File/folder picker dialogs
- ✅ Input validation with user-friendly error messages
- 🖼️ Application icon

### Changed
- Complete rewrite from CLI to GUI
- Improved error handling and reporting
- Better progress visualization

### Removed
- `list` command (may be re-added as GUI feature in future)
- CLI interface (may be re-added as alternative mode)

## [1.0.0] - Previous Release
- Original CLI version
```

---

## Phase 9: Final Testing and Release

### Task 9.1: Create test checklist
**File**: `TESTING.md` or checklist document  
**Priority**: High  
**Estimated Time**: 20 minutes  
**Dependencies**: Phase 8 complete

**Actions**:
1. Create comprehensive test checklist
2. Include all functional tests
3. Include edge cases and error conditions

**Checklist** (partial):
- [ ] Check Update with valid model/region
- [ ] Check Update with invalid model/region
- [ ] Download with auto-detect version
- [ ] Download with specific version
- [ ] Download with 8-digit TAC
- [ ] Download with 15-digit IMEI
- [ ] Download resume after cancellation
- [ ] Decrypt with V2 encryption
- [ ] Decrypt with V4 encryption
- [ ] All file dialogs work correctly
- [ ] Progress bars update smoothly
- [ ] Error messages are clear
- [ ] Window resizing works
- [ ] Tab switching works

---

### Task 9.2: Perform full integration testing
**File**: N/A (Testing)  
**Priority**: High  
**Estimated Time**: 60-90 minutes  
**Dependencies**: Task 9.1

**Actions**:
1. Go through entire test checklist
2. Document any issues found
3. Fix critical issues
4. Retest fixed issues

**Success Criteria**:
- All critical tests pass
- No crashes or hangs
- All features work as documented

---

### Task 9.3: Performance testing
**File**: N/A (Testing)  
**Priority**: Medium  
**Estimated Time**: 30 minutes  
**Dependencies**: Task 9.2

**Actions**:
1. Monitor memory usage during large download
2. Check CPU usage during decryption
3. Test with multiple tabs open
4. Measure application startup time

**Acceptance Criteria**:
- Startup time < 3 seconds
- Memory usage < 150 MB during operations
- CPU usage reasonable during decryption
- No memory leaks after operations

---

### Task 9.4: Create release build
**File**: N/A (Build process)  
**Priority**: High  
**Estimated Time**: 30 minutes  
**Dependencies**: Task 9.2

**Actions**:
1. Update version number in code (if applicable)
2. Build release binaries for all platforms
3. Test release binaries (not just debug builds)
4. Create checksums for binaries

**Commands**:
```bash
# Build release versions
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H windowsgui" -o release/susgo-windows-amd64.exe
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o release/susgo-linux-amd64
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o release/susgo-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o release/susgo-darwin-arm64

# Generate checksums
cd release
sha256sum * > checksums.txt
```

---

### Task 9.5: Create GitHub release
**File**: N/A (GitHub)  
**Priority**: Medium  
**Estimated Time**: 20 minutes  
**Dependencies**: Task 9.4

**Actions**:
1. Tag the release: `git tag v2.0.0`
2. Push tag: `git push origin v2.0.0`
3. Create GitHub release with:
   - Release notes from CHANGELOG
   - All binary files
   - checksums.txt
   - Installation instructions

**Release Notes Template**:
```markdown
## susgo v2.0.0 - GUI Edition

Major rewrite with cross-platform GUI!

### Downloads
- [Windows (64-bit)](link)
- [macOS (Intel)](link)
- [macOS (Apple Silicon)](link)
- [Linux (64-bit)](link)

### What's New
- Beautiful cross-platform GUI
- Visual progress tracking
- Improved error messages
- File picker dialogs

### Installation
See [README](link) for detailed instructions.

### Changes
See [CHANGELOG](link) for full list of changes.
```

---

## Summary Checklist

### Core Development
- [ ] Phase 1: Remove list command (4 tasks)
- [ ] Phase 2: Setup Fyne (3 tasks)
- [ ] Phase 3: Basic GUI (3 tasks)
- [ ] Phase 4: Download tab (3 tasks)
- [ ] Phase 5: Decrypt tab (2 tasks)
- [ ] Phase 6: Polish (5 tasks)

### Testing & Release
- [ ] Phase 7: Cross-platform testing (5 tasks)
- [ ] Phase 8: Documentation (5 tasks)
- [ ] Phase 9: Final testing and release (5 tasks)

### Total Tasks: 35

---

## Quick Start Guide (For Developers)

To start implementing:

1. **First**: Complete Phase 1 (remove list command) - can be done immediately
2. **Second**: Install Fyne and test basic window (Task 2.1, 2.2, 3.2)
3. **Third**: Implement Check Update tab (Task 3.3) - simplest tab, good learning
4. **Fourth**: Implement Download and Decrypt tabs (Phases 4-5)
5. **Fifth**: Polish and test (Phases 6-7)
6. **Finally**: Document and release (Phases 8-9)

## Time Estimates

- **Minimum (experienced developer)**: ~20-25 hours
- **Average**: ~30-35 hours
- **With learning Fyne**: ~40-50 hours

---

**Document Version**: 1.0  
**Last Updated**: 2025-01-21  
**Total Tasks**: 35  
**Estimated Total Time**: 23-34 hours
