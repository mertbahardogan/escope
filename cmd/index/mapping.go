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

var mappingCmd = &cobra.Command{
	Use:                "mapping",
	Short:              "Show index mapping information",
	SilenceErrors:      true,
	DisableSuggestions: true,
	Args:               cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			fmt.Println("Error: --name flag is required")
			fmt.Println("Usage: escope index mapping --name <index-name>")
			return
		}

		runIndexMapping(name)
	},
}

func runIndexMapping(indexName string) {
	client := elastic.NewClientWrapper(connection.GetClient())
	indexService := services.NewIndexService(client)

	fields, err := util.ExecuteWithTimeout(func() ([]models.FieldMapping, error) {
		return indexService.GetIndexMapping(context.Background(), indexName)
	})
	if util.HandleServiceErrorWithReturn(err, "Index mapping fetch") {
		return
	}

	if len(fields) == 0 {
		fmt.Printf("No mapping found for index '%s'\n", indexName)
		return
	}

	// Separate root fields and nested fields
	var rootFields []models.FieldMapping
	var nestedFields []models.FieldMapping

	for _, field := range fields {
		if field.Depth == 0 && field.Type != "object" {
			rootFields = append(rootFields, field)
		} else {
			nestedFields = append(nestedFields, field)
		}
	}

	// Sort both slices by path
	sort.Slice(rootFields, func(i, j int) bool {
		return rootFields[i].Path < rootFields[j].Path
	})
	sort.Slice(nestedFields, func(i, j int) bool {
		return nestedFields[i].Path < nestedFields[j].Path
	})

	formatter := ui.NewGenericTableFormatter()
	headers := []string{"Field Path", "Type", "Index"}

	fmt.Printf("\nIndex: %s\n", indexName)

	// Print root fields
	if len(rootFields) > 0 {
		fmt.Printf("\nFields (%d)\n", len(rootFields))
		rows := make([][]string, 0, len(rootFields))
		for _, field := range rootFields {
			rows = append(rows, []string{field.Path, field.Type, field.Index})
		}
		fmt.Print(formatter.FormatTable(headers, rows))
	}

	// Print nested fields
	if len(nestedFields) > 0 {
		fmt.Printf("\nNested Fields (%d)\n", len(nestedFields))
		rows := make([][]string, 0, len(nestedFields))
		for _, field := range nestedFields {
			rows = append(rows, []string{field.Path, field.Type, field.Index})
		}
		fmt.Print(formatter.FormatTable(headers, rows))
	}

	fmt.Printf("\nTotal: %d fields\n", len(fields))
}

func init() {
	indexCmd.AddCommand(mappingCmd)
	mappingCmd.Flags().StringP("name", "n", "", "Index name to show mapping for (required)")
	mappingCmd.MarkFlagRequired("name")
}
