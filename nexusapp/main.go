package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	modbus_scanner "nexusapp/ip_scanner"
	"nexusapp/nexus_about"
	"nexusapp/nexus_modbus_bits"
	modbus_rtu_scanner "nexusapp/rtu_scanner"

	"github.com/fyne-io/examples/img/icon"
)

type appInfo struct {
	name string
	icon fyne.Resource
	canv bool
	run  func(fyne.Window) fyne.CanvasObject
}

var apps = []appInfo{
	{"RTU Scanner", icon.BugBitmap, true, modbus_rtu_scanner.Show},
	{"Bits", icon.BugBitmap, true, nexus_modbus_bits.Show},
	{"IP Scanner", icon.BugBitmap, true, modbus_scanner.Show},
	{"About", icon.BugBitmap, true, nexus_about.Show},
}

func main() {
	a := app.New()
	resourceIconPng, err := fyne.LoadResourceFromPath("NEXUS.ico")
	if err != nil {
		panic(err)
	}
	a.SetIcon(resourceIconPng)

	w := a.NewWindow("Nexus Scanner")

	apps[0].icon = theme.RadioButtonIcon() // lazy load Fyne resource to avoid error

	// Create a slice to hold pointers to TabItem
	var tabItems []*container.TabItem
	for _, app := range apps {
		tab := container.NewTabItem(app.name, container.NewMax(app.run(w)))
		tabItems = append(tabItems, tab) // Append the pointer to TabItem
	}

	// Create the tabs with the slice of pointers
	tabs := container.NewAppTabs(tabItems...)
	tabs.SetTabLocation(container.TabLocationTop)

	// Create a top bar
	topBar := container.NewBorder(
		nil, nil, nil, tabs,
	)

	w.SetContent(topBar)
	//w.Resize(fyne.NewSize(1100, 710)) // Adjust the window size as needed
	w.SetFullScreen(false) // Ensure the window is not full-screen on launch
	// Listen for changes in window size and prevent full-screen
	//w.SetFixedSize(true)

	w.ShowAndRun()
}
