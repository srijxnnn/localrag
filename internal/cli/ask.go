package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/srijxnnn/localrag/internal/ollama"
)

var askCmd = &cobra.Command{
	Use:   "ask <question>",
	Short: "Ask a question about indexed documents",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		q := strings.Join(args, " ")

		client := ollama.New("http://localhost:11434")
		vec, err := client.Embed("nomic-embed-text", q)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Printf("embedding length: %d\n", len(vec))
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
