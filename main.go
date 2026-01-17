package main

import (
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

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("susgo - Samsung Firmware Downloader")

	tabs := container.NewAppTabs(
		container.NewTabItem("Check Update", makeCheckUpdateTab()),
	)

	myWindow.SetContent(tabs)
	myWindow.Resize(fyne.NewSize(700, 500))
	myWindow.ShowAndRun()
}
