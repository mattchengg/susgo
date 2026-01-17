package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

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

func makeDownloadTab() *fyne.Container {
	// Create input fields
	modelEntry := widget.NewEntry()
	modelEntry.SetPlaceHolder("e.g., SM-S928B")

	regionEntry := widget.NewEntry()
	regionEntry.SetPlaceHolder("e.g., EUX")

	imeiEntry := widget.NewEntry()
	imeiEntry.SetPlaceHolder("8 digits (TAC) or 15 digits (full IMEI)")

	versionEntry := widget.NewEntry()
	versionEntry.SetPlaceHolder("Leave empty for latest")

	// Output directory entry field
	outputDirEntry := widget.NewEntry()
	outputDirEntry.SetPlaceHolder("/path/to/output/directory")

	// Progress widgets
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	// Track download state
	var downloadInProgress bool

	// Download button
	downloadButton := widget.NewButton("Start Download", func() {
		if downloadInProgress {
			statusLabel.SetText("⚠️ Download already in progress")
			return
		}

		// Get and validate inputs
		model := strings.TrimSpace(modelEntry.Text)
		region := strings.TrimSpace(regionEntry.Text)
		imei := strings.TrimSpace(imeiEntry.Text)
		version := strings.TrimSpace(versionEntry.Text)
		outputDir := strings.TrimSpace(outputDirEntry.Text)

		// Validation
		if model == "" {
			statusLabel.SetText("❌ Error: Model is required")
			return
		}

		if region == "" {
			statusLabel.SetText("❌ Error: Region is required")
			return
		}

		if imei == "" {
			statusLabel.SetText("❌ Error: IMEI/TAC is required")
			return
		}

		// Validate IMEI length
		if len(imei) != 8 && len(imei) != 15 {
			statusLabel.SetText("❌ Error: IMEI must be 8 or 15 digits")
			return
		}

		if outputDir == "" {
			statusLabel.SetText("❌ Error: Output directory is required")
			return
		}

		// Start download
		downloadInProgress = true
		downloadButton.Disable()
		progressBar.Show()
		statusLabel.SetText("⏳ Initializing download...")

		// Create progress reporter
		progress := NewGUIProgressReporter(progressBar, statusLabel)

		// Run download in goroutine to keep UI responsive
		go func() {
			defer func() {
				downloadInProgress = false
				downloadButton.Enable()
			}()

			// Call download logic (will be implemented in Task 4.3)
			err := downloadFirmware(model, region, imei, version, outputDir, progress)

			if err != nil {
				statusLabel.SetText("❌ Error: " + err.Error())
				progressBar.Hide()
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
		widget.NewFormItem("Version", versionEntry),
		widget.NewFormItem("Output Directory *", outputDirEntry),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("Download Firmware",
			fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		downloadButton,
		widget.NewSeparator(),
		progressBar,
		statusLabel,
	)
}

// downloadFirmware implements the complete firmware download logic
func downloadFirmware(model, region, imei, version, outputDir string, progress ProgressReporter) error {
	// Step 1: Validate and parse IMEI
	progress.SetStatus("Validating IMEI...")
	effectiveIMEI, err := parseIMEI(imei, "", model, region)
	if err != nil {
		return fmt.Errorf("invalid IMEI: %w", err)
	}

	// Step 2: Create FUS client
	progress.SetStatus("Connecting to Samsung servers...")
	client := NewFUSClient()

	// Step 3: Get version if not specified (fetch latest)
	if version == "" {
		progress.SetStatus("Fetching latest firmware version...")
		ver, err := getLatestVersion(model, region)
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}
		version = ver
		progress.SetStatus(fmt.Sprintf("Latest version: %s", version))
	}

	// Step 4: Get binary file information
	progress.SetStatus("Retrieving firmware information...")
	path, filename, size, err := getBinaryFile(client, version, model, region, effectiveIMEI)
	if err != nil {
		return fmt.Errorf("failed to get firmware info: %w", err)
	}

	// Step 5: Determine output file path
	out := filepath.Join(outputDir, filename)

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	progress.SetStatus(fmt.Sprintf("Firmware: %s (%.2f GB)", filename, float64(size)/(1024*1024*1024)))

	// Step 6: Check if already decrypted
	decFile := strings.TrimSuffix(strings.TrimSuffix(out, ".enc4"), ".enc2")
	if _, err := os.Stat(decFile); err == nil {
		progress.SetStatus("✅ File already decrypted!")
		return nil
	}

	// Step 7: Check for existing file (resume support)
	var offset int64
	if info, err := os.Stat(out); err == nil {
		offset = info.Size()
		if offset == size {
			progress.SetStatus("File already downloaded, decrypting...")
			autoDecrypt(out, filename, version, model, region, effectiveIMEI)
			progress.Finish()
			return nil
		}
		progress.SetStatus(fmt.Sprintf("Resuming from %.1f%%", float64(offset)/float64(size)*100))
	}

	// Step 8: Initialize download session
	progress.SetStatus("Initializing download...")
	initDownload(client, filename)

	// Step 9: Start download
	resp, err := client.DownloadFile(path+filename, offset)
	if err != nil {
		return fmt.Errorf("failed to start download: %w", err)
	}
	defer resp.Body.Close()

	// Step 10: Open output file for writing
	flags := os.O_CREATE | os.O_WRONLY
	if offset > 0 {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	fd, err := os.OpenFile(out, flags, 0644)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer fd.Close()

	// Step 11: Setup progress tracking
	progress.SetTotal(size)
	progress.SetCurrent(offset)
	progress.SetStatus("Downloading...")

	// Step 12: Download file in chunks
	buf := make([]byte, 32768) // 32 KB chunks

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := fd.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("failed to write to file: %w", writeErr)
			}
			progress.Add(int64(n))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("download error: %w", err)
		}
	}

	// Step 13: Download complete
	progress.SetStatus("Download complete! Decrypting...")

	// Step 14: Auto-decrypt if applicable
	autoDecrypt(out, filename, version, model, region, effectiveIMEI)

	// Step 15: Mark as finished
	progress.Finish()

	return nil
}

func makeDecryptTab(window fyne.Window) *fyne.Container {
	// Create input fields
	modelEntry := widget.NewEntry()
	modelEntry.SetPlaceHolder("e.g., SM-S928B")

	regionEntry := widget.NewEntry()
	regionEntry.SetPlaceHolder("e.g., EUX")

	imeiEntry := widget.NewEntry()
	imeiEntry.SetPlaceHolder("8 digits (TAC) or 15 digits (full IMEI)")

	versionEntry := widget.NewEntry()
	versionEntry.SetPlaceHolder("e.g., S928BXXU1AXXX/S928BOXM1AXXX/...")

	inputFileEntry := widget.NewEntry()
	inputFileEntry.SetPlaceHolder("/path/to/firmware.enc4")

	outputFileEntry := widget.NewEntry()
	outputFileEntry.SetPlaceHolder("/path/to/output/firmware.zip")

	// Encryption version selector (default to 4)
	encVersionSelect := widget.NewSelect([]string{"2", "4"}, nil)
	encVersionSelect.SetSelected("4")

	// Browse button for input file
	browseInputButton := widget.NewButton("Browse...", func() {
		// Create file open dialog with .enc2 and .enc4 filter
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if reader == nil {
				return // User cancelled
			}
			defer reader.Close()

			// Set the file path in the entry
			inputFileEntry.SetText(reader.URI().Path())
		}, window)

		// Set file filter for .enc2 and .enc4 files
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".enc2", ".enc4"}))
		fileDialog.Show()
	})

	// Browse button for output file
	browseOutputButton := widget.NewButton("Browse...", func() {
		// Create file save dialog
		fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if writer == nil {
				return // User cancelled
			}
			defer writer.Close()

			// Set the file path in the entry
			outputFileEntry.SetText(writer.URI().Path())
		}, window)

		// Set default filename suggestion
		fileDialog.SetFileName("firmware.zip")
		fileDialog.Show()
	})

	// Progress widgets
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord

	// Track decrypt state
	var decryptInProgress bool

	// Decrypt button
	decryptButton := widget.NewButton("Decrypt Firmware", func() {
		if decryptInProgress {
			statusLabel.SetText("⚠️ Decryption already in progress")
			return
		}

		// Get and validate inputs
		model := strings.TrimSpace(modelEntry.Text)
		region := strings.TrimSpace(regionEntry.Text)
		imei := strings.TrimSpace(imeiEntry.Text)
		version := strings.TrimSpace(versionEntry.Text)
		inputFile := strings.TrimSpace(inputFileEntry.Text)
		outputFile := strings.TrimSpace(outputFileEntry.Text)
		encVersion := encVersionSelect.Selected

		// Validation
		if model == "" {
			statusLabel.SetText("❌ Error: Model is required")
			return
		}

		if region == "" {
			statusLabel.SetText("❌ Error: Region is required")
			return
		}

		if imei == "" {
			statusLabel.SetText("❌ Error: IMEI/TAC is required")
			return
		}

		// Validate IMEI length
		if len(imei) != 8 && len(imei) != 15 {
			statusLabel.SetText("❌ Error: IMEI must be 8 or 15 digits")
			return
		}

		if version == "" {
			statusLabel.SetText("❌ Error: Version is required")
			return
		}

		if inputFile == "" {
			statusLabel.SetText("❌ Error: Input file path is required")
			return
		}

		if outputFile == "" {
			statusLabel.SetText("❌ Error: Output file path is required")
			return
		}

		if encVersion == "" {
			statusLabel.SetText("❌ Error: Encryption version is required")
			return
		}

		// Check if input file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			statusLabel.SetText("❌ Error: Input file does not exist")
			return
		}

		// Start decryption
		decryptInProgress = true
		decryptButton.Disable()
		progressBar.Show()
		statusLabel.SetText("⏳ Starting decryption...")

		// Create progress reporter
		progress := NewGUIProgressReporter(progressBar, statusLabel)

		// Run decryption in goroutine to keep UI responsive
		go func() {
			defer func() {
				decryptInProgress = false
				decryptButton.Enable()
				progressBar.Hide()
			}()

			err := decryptFirmwareGUI(model, region, imei, version, inputFile, outputFile, encVersion, progress)

			if err != nil {
				statusLabel.SetText("❌ Error: " + err.Error())
			} else {
				statusLabel.SetText("✅ Decryption complete! File saved to: " + outputFile)
			}
		}()
	})

	// Layout
	form := widget.NewForm(
		widget.NewFormItem("Model *", modelEntry),
		widget.NewFormItem("Region *", regionEntry),
		widget.NewFormItem("IMEI/TAC *", imeiEntry),
		widget.NewFormItem("Version *", versionEntry),
		widget.NewFormItem("Input File *", container.NewBorder(nil, nil, nil, browseInputButton, inputFileEntry)),
		widget.NewFormItem("Output File *", container.NewBorder(nil, nil, nil, browseOutputButton, outputFileEntry)),
		widget.NewFormItem("Encryption Version *", encVersionSelect),
	)

	return container.NewVBox(
		widget.NewLabelWithStyle("Decrypt Firmware",
			fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		form,
		decryptButton,
		widget.NewSeparator(),
		progressBar,
		statusLabel,
	)
}

// decryptFirmwareGUI implements the complete firmware decryption logic for GUI
func decryptFirmwareGUI(model, region, imei, version, inputFile, outputFile, encVersion string, progress ProgressReporter) error {
	// Step 1: Validate and parse IMEI
	progress.SetStatus("⏳ Validating IMEI...")
	effectiveIMEI, err := parseIMEI(imei, "", model, region)
	if err != nil {
		return fmt.Errorf("invalid IMEI: %w", err)
	}

	// Step 2: Check if output file already exists
	if _, err := os.Stat(outputFile); err == nil {
		progress.SetStatus("⚠️ Output file already exists, overwriting...")
	}

	// Step 3: Get decryption key based on version
	var key []byte
	if encVersion == "2" {
		progress.SetStatus("⏳ Generating V2 decryption key...")
		key = getV2Key(version, model, region)
	} else {
		progress.SetStatus("⏳ Generating V4 decryption key...")
		key, err = getV4Key(version, model, region, effectiveIMEI)
		if err != nil {
			return fmt.Errorf("failed to generate V4 key: %w", err)
		}
	}

	// Step 4: Decrypt the firmware with progress reporting
	if err := decryptFirmwareWithProgress(inputFile, outputFile, key, progress); err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Step 5: Mark as finished
	progress.Finish()

	return nil
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("susgo - Samsung Firmware Downloader")

	tabs := container.NewAppTabs(
		container.NewTabItem("Check Update", makeCheckUpdateTab()),
		container.NewTabItem("Download", makeDownloadTab()),
		container.NewTabItem("Decrypt", makeDecryptTab(myWindow)),
	)

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(700, 500))
	myWindow.ShowAndRun()
}
