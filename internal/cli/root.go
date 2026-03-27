package cli

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "rag",
	Short: "Query your documents locally with AI (Ollama)",
	Long: `A local-first CLI to index documents and ask questions.
No cloud, runs on your machine via Ollama.`,
}

// Execute runs the root command and returns any parse or runtime error from Cobra.
func Execute() error {
	return rootCmd.Execute()
}
