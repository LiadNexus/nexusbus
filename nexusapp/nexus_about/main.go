package nexus_about

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Current version of the app
const currentVersion = "0.3.0"

// GitHub API URL to fetch the latest release
const releaseURL = "https://api.github.com/repos/Nexus-Te/NexusScanner/releases/latest"

// Release represents the structure of the GitHub release response
type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	HTMLURL string  `json:"html_url"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a downloadable file in the release
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// AboutPage creates the UI for the About page
func AboutPage(win fyne.Window) fyne.CanvasObject {
	aboutTitle := widget.NewLabel("Nexus Scanner by Nexus Technologies")
	aboutText := widget.NewLabel(fmt.Sprintf(
		"Nexus Scanner is a versatile tool for scanning and interacting with Modbus devices.\n"+
			"It supports various protocols and provides a user-friendly interface for configuring and monitoring devices.\n\n"+
			"Version: %s\n\nContact: support@nexus-te.com", currentVersion))

	checkUpdatesButton := widget.NewButton("Check for Updates", func() {
		checkForUpdates(win)
	})

	return container.NewVBox(
		aboutTitle,
		aboutText,
		checkUpdatesButton,
	)
}

// Show initializes the About page and loads it into the given window
func Show(win fyne.Window) fyne.CanvasObject {
	return AboutPage(win)
}

// checkForUpdates fetches the latest release from GitHub and compares versions
func checkForUpdates(win fyne.Window) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(releaseURL)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to fetch release information: %v", err), win)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to read response: %v", err), win)
		return
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		dialog.ShowError(fmt.Errorf("Failed to parse release data: %v", err), win)
		return
	}

	if release.TagName > currentVersion {
		for _, asset := range release.Assets {
			if filepath.Ext(asset.Name) == ".exe" {
				showUpdateDialog(win, release, asset)
				return
			}
		}
		dialog.ShowInformation("Update Available", "New version found, but no installer available.", win)
	} else {
		dialog.ShowInformation("Up-to-Date", "You are using the latest version.", win)
	}
}

// showUpdateDialog presents an update dialog with a download button and clickable link
func showUpdateDialog(win fyne.Window, release Release, asset Asset) {
	releaseLink := widget.NewHyperlink("Release Page", parseURL(release.HTMLURL))
	downloadButton := widget.NewButton("Download and Install", func() {
		go downloadAndInstall(win, asset)
	})

	dialog.NewCustom(
		"Update Available",
		"Close",
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("A new version (%s) is available!", release.TagName)),
			releaseLink,
			downloadButton,
		),
		win,
	).Show()
}

// downloadAndInstall downloads the installer, saves it, executes it, and closes the app
func downloadAndInstall(win fyne.Window, asset Asset) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(asset.BrowserDownloadURL)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to download installer: %v", err), win)
		return
	}
	defer resp.Body.Close()

	// Get the absolute path for the installer
	filePath := filepath.Join(os.TempDir(), asset.Name)
	out, err := os.Create(filePath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to create file: %v", err), win)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to save installer: %v", err), win)
		return
	}

	dialog.ShowInformation("Download Complete", fmt.Sprintf("Installer saved at %s", filePath), win)

	// Execute the installer
	go func() {
		if err := exec.Command(filePath).Start(); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to execute installer: %v", err), win)
			return
		}

		// Close the current app after the installer starts
		time.Sleep(1 * time.Second) // Ensure installer starts before exiting
		os.Exit(0)
	}()
}

// parseURL safely parses a URL string
func parseURL(urlStr string) *url.URL {
	parsedURL, _ := url.Parse(urlStr)
	return parsedURL
}
