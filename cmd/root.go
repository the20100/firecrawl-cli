package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/the20100/firecrawl-cli/client"
)

var (
	apiKey string
	apiURL string
)

var rootCmd = &cobra.Command{
	Use:   "firecrawl",
	Short: "Firecrawl CLI",
	Long: `A command-line interface for the Firecrawl API.

Supports scraping, searching, mapping, crawling, and AI-powered extraction.

Configuration:
  Set your API key via the FIRECRAWL_API_KEY environment variable,
  or pass it with --api-key.

  For self-hosted instances, set FIRECRAWL_API_URL or use --api-url.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Firecrawl API key (or set FIRECRAWL_API_KEY)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "", "Firecrawl API URL (or set FIRECRAWL_API_URL)")
}

func newClient() *client.Client {
	key := apiKey
	if key == "" {
		key = os.Getenv("FIRECRAWL_API_KEY")
	}
	if key == "" {
		fmt.Fprintln(os.Stderr, "error: API key required. Set FIRECRAWL_API_KEY or use --api-key")
		os.Exit(1)
	}

	url := apiURL
	if url == "" {
		url = os.Getenv("FIRECRAWL_API_URL")
	}

	return client.New(key, url)
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
