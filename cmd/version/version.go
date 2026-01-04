package version

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mertbahardogan/escope/cmd/core"
	"github.com/spf13/cobra"
)

const githubAPIURL = "https://api.github.com/repos/mertbahardogan/escope/releases/latest"

type githubRelease struct {
	TagName string `json:"tag_name"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the current version of escope",
	Run: func(cmd *cobra.Command, args []string) {
		version, err := getLatestVersion()
		if err != nil {
			fmt.Printf("escope version: unable to fetch (%v)\n", err)
			return
		}
		fmt.Printf("escope version %s\n", version)
	},
}

func init() {
	core.RootCmd.AddCommand(versionCmd)
}

func getLatestVersion() (string, error) {
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
