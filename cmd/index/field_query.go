package index

import (
	"context"
	"fmt"

	"github.com/mertbahardogan/escope/internal/connection"
	"github.com/mertbahardogan/escope/internal/elastic"
	"github.com/mertbahardogan/escope/internal/services"
	"github.com/mertbahardogan/escope/internal/ui"
	"github.com/mertbahardogan/escope/internal/util"
	"github.com/spf13/cobra"
)

var existsCmd = &cobra.Command{
	Use:   "exists",
	Short: "Count documents by field (exists or exact term)",
	Long: `Runs _count on the index or alias given by --name.

Without --value: exists query (documents that have the field).

With --value: term query on field = value (exact match; numbers and booleans are detected from the string).

Use --nested for nested fields: path is the first segment of --field
(e.g. --field comments.author uses path "comments").`,
	SilenceErrors:      true,
	DisableSuggestions: true,
	Args:               cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		field, _ := cmd.Flags().GetString("field")
		value, _ := cmd.Flags().GetString("value")
		nested, _ := cmd.Flags().GetBool("nested")

		name = resolveIndexName(name)
		if name == "" || field == "" {
			if name == "" {
				printIndexNameRequired()
			} else {
				fmt.Println("Error: --field is required")
			}
			fmt.Println("Usage: escope index exists [--name <index-or-alias>] --field <field> [--value <exact>] [--nested]")
			return
		}

		runFieldExists(name, field, value, nested)
	},
}

var cardinalityCmd = &cobra.Command{
	Use:   "cardinality",
	Short: "Approximate distinct values (cardinality aggregation)",
	Long: `Runs a cardinality aggregation on the given field (approximate unique values).

Use --nested for values inside a nested mapping: path is the first segment of --field.

For analyzed text you usually need a .keyword (or other doc_values) subfield.`,
	SilenceErrors:      true,
	DisableSuggestions: true,
	Args:               cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		field, _ := cmd.Flags().GetString("field")
		nested, _ := cmd.Flags().GetBool("nested")

		name = resolveIndexName(name)
		if name == "" || field == "" {
			if name == "" {
				printIndexNameRequired()
			} else {
				fmt.Println("Error: --field is required")
			}
			fmt.Println("Usage: escope index cardinality [--name <index-or-alias>] --field <field> [--nested]")
			return
		}

		runFieldCardinality(name, field, nested)
	},
}

func runFieldExists(indexName, field, value string, nested bool) {
	client := elastic.NewClientWrapper(connection.GetClient())
	indexService := services.NewIndexService(client)

	count, err := util.ExecuteWithTimeout(func() (int64, error) {
		return indexService.CountDocumentsByFieldQuery(context.Background(), indexName, field, value, nested)
	})
	if util.HandleServiceErrorWithReturn(err, "Field query") {
		return
	}

	exists := "No"
	if count > 0 {
		exists = "Yes"
	}

	headers := []string{"Documents", "Exists"}
	rows := [][]string{
		{
			util.FormatDocsCount(count),
			exists,
		},
	}

	formatter := ui.NewGenericTableFormatter()
	fmt.Print(formatter.FormatTable(headers, rows))
}

func runFieldCardinality(indexName, field string, nested bool) {
	client := elastic.NewClientWrapper(connection.GetClient())
	indexService := services.NewIndexService(client)

	distinct, err := util.ExecuteWithTimeout(func() (int64, error) {
		return indexService.FieldValueCardinality(context.Background(), indexName, field, nested)
	})
	if util.HandleServiceErrorWithReturn(err, "Cardinality query") {
		return
	}

	headers := []string{"Approx. distinct values"}
	rows := [][]string{
		{util.FormatDocsCount(distinct)},
	}

	formatter := ui.NewGenericTableFormatter()
	fmt.Print(formatter.FormatTable(headers, rows))
}

func init() {
	indexCmd.AddCommand(existsCmd)
	indexCmd.AddCommand(cardinalityCmd)

	registerFieldQueryFlags := func(cmd *cobra.Command, nestedUsage string) {
		cmd.Flags().StringP("name", "n", "", "Index or alias (defaults to index from 'escope index use')")
		cmd.Flags().StringP("field", "f", "", "Field path (required)")
		cmd.Flags().Bool("nested", false, nestedUsage)
		cmd.MarkFlagRequired("field")
	}

	registerFieldQueryFlags(existsCmd, "Nested query: path is the first segment of --field (e.g. comments.author → path comments)")
	existsCmd.Flags().StringP("value", "v", "", "If set, count documents where field equals this value (term query); omit for exists-only")

	registerFieldQueryFlags(cardinalityCmd, "Nested cardinality: path is the first segment of --field")
}
