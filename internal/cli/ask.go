package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/srijxnnn/localrag/internal/ollama"
	"github.com/srijxnnn/localrag/internal/search"
	"github.com/srijxnnn/localrag/internal/store"
)

const defaultTopK = 3

var askCmd = &cobra.Command{
	Use:   "ask <question>",
	Short: "Ask a question about indexed documents",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat(store.DBPath()); err != nil {
			return fmt.Errorf("workspace not initialized; run: rag init")
		}
		q := strings.Join(args, " ")

		s, err := store.Init(store.DBPath())
		if err != nil {
			return err
		}
		defer s.Close()

		chunks, err := s.AllChunks()
		if err != nil {
			return err
		}
		if len(chunks) == 0 {
			return fmt.Errorf("no indexed chunks; run: rag add <file>")
		}

		client := ollama.New(defaultOllamaURL)
		queryVec, err := client.Embed(embedModel, q)
		if err != nil {
			return err
		}

		top := search.TopK(queryVec, chunks, defaultTopK)
		for i, c := range top {
			fmt.Printf("--- match %d (score %.4f, %s) ---\n%s\n", i+1, search.Cosine(queryVec, c.Embedding), c.DocumentPath, c.Text)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
