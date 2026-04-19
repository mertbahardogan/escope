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

var analyzerCmd = &cobra.Command{
	Use:                "analyzer",
	Short:              "Show index field analyzer information",
	SilenceErrors:      true,
	DisableSuggestions: true,
	Args:               cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		flagName, _ := cmd.Flags().GetString("name")
		name := resolveIndexName(flagName)

		if name == "" {
			printIndexNameRequired()
			fmt.Println("Usage: escope index analyzer [--name <index-name>]")
			return
		}

		runIndexAnalyzer(name)
	},
}

func runIndexAnalyzer(indexName string) {
	client := elastic.NewClientWrapper(connection.GetClient())
	indexService := services.NewIndexService(client)

	fields, err := util.ExecuteWithTimeout(func() ([]models.FieldMapping, error) {
		return indexService.GetIndexMapping(context.Background(), indexName)
	})
	if util.HandleServiceErrorWithReturn(err, "Index mapping fetch") {
		return
	}

	// Filter only fields that have analyzer, search_analyzer, or normalizer
	var analyzerFields []models.FieldMapping
	for _, field := range fields {
		if field.Analyzer != "-" || field.SearchAnalyzer != "-" || field.Normalizer != "-" {
			analyzerFields = append(analyzerFields, field)
		}
	}

	if len(analyzerFields) == 0 {
		fmt.Printf("No analyzer configuration found for index '%s'\n", indexName)
		return
	}

	// Sort fields by path for consistent display
	sort.Slice(analyzerFields, func(i, j int) bool {
		return analyzerFields[i].Path < analyzerFields[j].Path
	})

	headers := []string{"Field Path", "Analyzer", "Search Analyzer", "Normalizer"}
	rows := make([][]string, 0, len(analyzerFields))

	for _, field := range analyzerFields {
		row := []string{
			field.Path,
			field.Analyzer,
			field.SearchAnalyzer,
			field.Normalizer,
		}
		rows = append(rows, row)
	}

	formatter := ui.NewGenericTableFormatter()
	fmt.Printf("\nIndex: %s\n\n", indexName)
	fmt.Print(formatter.FormatTable(headers, rows))
	fmt.Printf("Total: %d fields with analyzer config\n", len(analyzerFields))
}

func init() {
	indexCmd.AddCommand(analyzerCmd)
	analyzerCmd.Flags().StringP("name", "n", "", "Index name (defaults to index from 'escope index use')")
}
