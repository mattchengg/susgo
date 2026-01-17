package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
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

// downloadFirmware is a placeholder that will be implemented in Task 4.3
func downloadFirmware(model, region, imei, version, outputDir string, progress ProgressReporter) error {
	// Validate IMEI
	effectiveIMEI, err := parseIMEI(imei, "", model, region)
	if err != nil {
		return fmt.Errorf("invalid IMEI: %w", err)
	}

	progress.SetStatus(fmt.Sprintf("Validating inputs (IMEI: %s)...", effectiveIMEI))

	// TODO: Task 4.3 will implement the actual download logic
	// This includes:
	// - Creating FUSClient
	// - Getting version if not specified
	// - Getting binary file info
	// - Downloading the file with progress updates
	// - Auto-decrypting if needed

	return fmt.Errorf("download logic not yet implemented (Task 4.3)")
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("susgo - Samsung Firmware Downloader")

	tabs := container.NewAppTabs(
		container.NewTabItem("Check Update", makeCheckUpdateTab()),
		container.NewTabItem("Download", makeDownloadTab()),
	)

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(700, 500))
	myWindow.ShowAndRun()
}
