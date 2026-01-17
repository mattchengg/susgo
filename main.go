package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("susgo - Samsung Firmware Downloader")

	// Placeholder content
	content := widget.NewLabel("susgo GUI - Coming Soon")

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(700, 500))
	myWindow.ShowAndRun()
}
