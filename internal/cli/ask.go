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

var askChatModel string

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
		prompt := buildRAGPrompt(q, top)

		out := cmd.OutOrStdout()
		if err := client.GenerateStream(askChatModel, prompt, out); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(out)
		return nil
	},
}

func buildRAGPrompt(question string, chunks []store.Chunk) string {
	var b strings.Builder
	b.WriteString("Use the following context to answer the question. If the answer is not contained in the context, say that you do not know.\n\nContext:\n\n")
	for i, c := range chunks {
		fmt.Fprintf(&b, "[Source %d: %s]\n%s\n\n", i+1, c.DocumentPath, c.Text)
	}
	b.WriteString("Question: ")
	b.WriteString(question)
	b.WriteString("\n\nAnswer:")
	return b.String()
}

func init() {
	askCmd.Flags().StringVar(&askChatModel, "model", "llama3.2:3b", "Ollama model for generation (must be pulled locally)")
	rootCmd.AddCommand(askCmd)
}
