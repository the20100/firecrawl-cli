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
  Set your API key via the FIRECRAWL_API_KEY environment variable (or aliases:
  FIRECRAWL_KEY, FIRECRAWL_API, API_KEY_FIRECRAWL, ...),
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

// resolveEnv returns the value of the first non-empty environment variable from the given names.
func resolveEnv(names ...string) string {
	for _, name := range names {
		if v := os.Getenv(name); v != "" {
			return v
		}
	}
	return ""
}

func newClient() *client.Client {
	key := apiKey
	if key == "" {
		key = resolveEnv(
			"FIRECRAWL_API_KEY", "FIRECRAWL_KEY", "FIRECRAWL_API", "API_KEY_FIRECRAWL", "API_FIRECRAWL", "FIRECRAWL_PK", "FIRECRAWL_PUBLIC",
			"FIRECRAWL_API_SECRET", "FIRECRAWL_SECRET_KEY", "FIRECRAWL_API_SECRET_KEY", "FIRECRAWL_SECRET", "SECRET_FIRECRAWL", "API_SECRET_FIRECRAWL", "SK_FIRECRAWL", "FIRECRAWL_SK",
		)
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
