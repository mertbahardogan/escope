package upgrade

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mertbahardogan/escope/internal/config"
)

const (
	githubAPIURL = "https://api.github.com/repos/mertbahardogan/escope/releases/latest"
	modulePath   = "github.com/mertbahardogan/escope@latest"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// GetLatestVersion fetches the latest version from GitHub Releases API
func GetLatestVersion() (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(githubAPIURL)
	if err != nil {
		return "", fmt.Errorf("connection failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("parse error")
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no releases found")
	}

	return release.TagName, nil
}

// CheckAndUpgrade checks for updates and upgrades if a new version is available
func CheckAndUpgrade() {
	fmt.Println("Checking for updates...")

	latestVersion, err := GetLatestVersion()
	if err != nil {
		fmt.Printf("Error checking for updates: %v\n", err)
		return
	}

	// Get installed version from config
	installedVersion, _ := config.GetInstalledVersion()

	// Normalize versions for comparison
	latestNormalized := normalizeVersion(latestVersion)
	installedNormalized := normalizeVersion(installedVersion)

	// Check if already up to date
	if installedNormalized != "" && installedNormalized == latestNormalized {
		fmt.Printf("Already up to date %s\n", latestVersion)
		return
	}

	if installedVersion != "" {
		fmt.Printf("Current version: %s\n", installedVersion)
		fmt.Printf("Latest version: %s\n", latestVersion)
	} else {
		fmt.Printf("Latest version: %s\n", latestVersion)
	}

	fmt.Println("Upgrading...")

	if err := runGoInstall(); err != nil {
		fmt.Printf("Error upgrading: %v\n", err)
		fmt.Println("\nYou can manually upgrade by running:")
		fmt.Printf("  go install %s\n", modulePath)
		return
	}

	// Save installed version to config
	if err := config.SetInstalledVersion(latestVersion); err != nil {
		fmt.Printf("Warning: could not save installed version: %v\n", err)
	}

	fmt.Printf("Successfully upgraded to %s\n", latestVersion)
}

func normalizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimSpace(v)
	return v
}

func runGoInstall() error {
	cmd := exec.Command("go", "install", modulePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
