package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var askCmd = &cobra.Command{
	Use:   "ask <question>",
	Short: "Ask a question about indexed documents",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		q := strings.Join(args, " ")
		fmt.Printf("Asking: %s\n", q)
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
