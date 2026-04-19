package index

import (
	"context"
	"fmt"
	"sort"

	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui"
	"github.com/mertbahardogan/escope/internal/util"
	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:                "settings",
	Short:              "Show index settings information",
	SilenceErrors:      true,
	DisableSuggestions: true,
	Args:               cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		flagName, _ := cmd.Flags().GetString("name")
		name := resolveIndexName(flagName)

		if name == "" {
			printIndexNameRequired()
			fmt.Println("Usage: escope index settings [--name <index-name>]")
			return
		}

		runIndexSettings(name)
	},
}

func runIndexSettings(indexName string) {
	client := elastic.NewClientWrapper(connection.GetClient())
	indexService := services.NewIndexService(client)

	settings, err := util.ExecuteWithTimeout(func() ([]models.IndexSettingInfo, error) {
		return indexService.GetIndexSettings(context.Background(), indexName)
	})
	if util.HandleServiceErrorWithReturn(err, "Index settings fetch") {
		return
	}

	if len(settings) == 0 {
		fmt.Printf("No settings found for index '%s'\n", indexName)
		return
	}

	// Sort settings by key for consistent display
	sort.Slice(settings, func(i, j int) bool {
		return settings[i].Key < settings[j].Key
	})

	headers := []string{"Setting", "Value"}
	rows := make([][]string, 0, len(settings))

	for _, setting := range settings {
		row := []string{
			setting.Key,
			setting.Value,
		}
		rows = append(rows, row)
	}

	formatter := ui.NewGenericTableFormatter()
	fmt.Printf("\nIndex: %s\n\n", indexName)
	fmt.Print(formatter.FormatTable(headers, rows))
	fmt.Printf("Total: %d settings\n", len(settings))
}

func init() {
	indexCmd.AddCommand(settingsCmd)
	settingsCmd.Flags().StringP("name", "n", "", "Index name (defaults to index from 'escope index use')")
}
