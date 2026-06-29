package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/version"
)

// VersionCheckURL is the GitHub API URL for the latest release.
const VersionCheckURL = "https://api.github.com/repos/petitorium/petitorium/releases/latest"

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckLatestVersion checks if there is a newer version of Petitorium available on GitHub.
// If a new version is found, it updates the footer right text.
func CheckLatestVersion(app *tview.Application, footerRight *tview.TextView, colors *ColorManager) {
	if config.C.DisableVersionCheck || version.Version == "dev" {
		return
	}

	// Run in background
	go func() {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		req, err := http.NewRequest(http.MethodGet, VersionCheckURL, nil)
		if err != nil {
			return
		}

		// GitHub API requires a User-Agent header
		req.Header.Set("User-Agent", "Petitorium/"+version.Version)

		resp, err := client.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return
		}

		var release githubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return
		}

		latestVersion := strings.TrimSpace(release.TagName)
		if latestVersion == "" {
			return
		}

		// Normalize versions by removing 'v' prefix for comparison
		currentNormalized := strings.TrimPrefix(version.Version, "v")
		latestNormalized := strings.TrimPrefix(latestVersion, "v")

		if latestNormalized != "" && latestNormalized != currentNormalized {
			app.QueueUpdateDraw(func() {
				footerRight.SetText(fmt.Sprintf("%s [yellow](%s available!) ", footerBrandText(colors), latestVersion))
			})
		}
	}()
}
