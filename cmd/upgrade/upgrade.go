package upgrade

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mertbahardogan/escope/cmd/core"
	"github.com/mertbahardogan/escope/internal/version"
	"github.com/spf13/cobra"
)

const (
	githubAPIURL = "https://api.github.com/repos/mertbahardogan/escope/releases/latest"
	modulePath   = "github.com/mertbahardogan/escope@latest"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade escope to the latest version",
	Long:  "Check for updates and upgrade escope to the latest version using go install",
	Run: func(cmd *cobra.Command, args []string) {
		runUpgrade()
	},
}

func init() {
	core.RootCmd.AddCommand(upgradeCmd)
}

func runUpgrade() {
	fmt.Println("Checking for updates...")

	latestVersion, err := getLatestVersion()
	if err != nil {
		fmt.Printf("Error checking for updates: %v\n", err)
		return
	}

	currentVersion := normalizeVersion(version.Version)
	latestNormalized := normalizeVersion(latestVersion)

	if currentVersion == "dev" {
		fmt.Printf("Current version: dev (development build)\n")
		fmt.Printf("Latest version: %s\n", latestVersion)
	} else {
		cmp := compareVersions(currentVersion, latestNormalized)
		if cmp >= 0 {
			fmt.Printf("Already up to date (%s)\n", version.Version)
			return
		}
		fmt.Printf("New version available: %s → %s\n", version.Version, latestVersion)
	}

	fmt.Println("Upgrading...")

	if err := runGoInstall(); err != nil {
		fmt.Printf("Error upgrading: %v\n", err)
		fmt.Println("\nYou can manually upgrade by running:")
		fmt.Printf("  go install %s\n", modulePath)
		return
	}

	fmt.Printf("Successfully upgraded to %s\n", latestVersion)
}

func getLatestVersion() (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(githubAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to connect to GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("no releases found")
	}

	return release.TagName, nil
}

func normalizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimSpace(v)
	// Remove -dirty suffix (local uncommitted changes)
	if idx := strings.Index(v, "-dirty"); idx != -1 {
		v = v[:idx]
	}
	// Remove commit hash suffix (e.g., v0.2.0-5-g1234567)
	if idx := strings.Index(v, "-"); idx != -1 {
		v = v[:idx]
	}
	return v
}

func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}
	return 0
}

func runGoInstall() error {
	cmd := exec.Command("go", "install", modulePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
