package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List indexed documents",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Listing documents")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
