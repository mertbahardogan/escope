package index

import (
	"fmt"

	"github.com/mertbahardogan/escope/internal/indexsession"
	"github.com/spf13/cobra"
)

var useClear bool

var useCmd = &cobra.Command{
	Use:                "use [index-or-alias]",
	Short:              "Remember index/alias for detail subcommands",
	Args:               cobra.MaximumNArgs(1),
	SilenceErrors:      true,
	DisableSuggestions: true,
	Long: `Stores the index or alias for the current Elasticsearch host so subcommands like
mapping, settings, analyzer, exists, and cardinality can omit --name/-n when a default is set.

With no arguments: print the current selection. --clear removes default_index for this host (calculator block, if any, is kept).`,
	Example: `  escope index use --help

  escope index use
  escope index use my-index
  escope index use --clear`,
	Run: func(cmd *cobra.Command, args []string) {
		if useClear && len(args) > 0 {
			fmt.Println("Error: do not pass an index name with --clear")
			return
		}
		if useClear {
			if err := indexsession.Clear(); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println("Cleared selected index for this host.")
			return
		}
		if len(args) == 0 {
			msg, err := indexsession.DescribeCurrent()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(msg)
			return
		}
		if err := indexsession.WriteSelectedIndex(args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Selected index for detail commands: %s\n", args[0])
	},
}

func init() {
	indexCmd.AddCommand(useCmd)
	useCmd.Flags().BoolVar(&useClear, "clear", false, "Clear default index")
}
