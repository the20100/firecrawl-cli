package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/the20100/firecrawl-cli/client"
	"github.com/the20100/firecrawl-cli/internal/config"
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

Token resolution order:
  1. FIRECRAWL_API_KEY env var (or aliases: FIRECRAWL_KEY, FIRECRAWL_API, ...)
  2. Config file (~/.config/firecrawl/config.json via: firecrawl auth set-key)
  3. --api-key flag

  For self-hosted instances, set FIRECRAWL_API_URL or use --api-url.`,
	SilenceUsage: true,
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

// resolveAPIKey returns the best available API key.
func resolveAPIKey() (string, error) {
	// 1. Flag
	if apiKey != "" {
		return apiKey, nil
	}

	// 2. Env var aliases
	if k := resolveEnv(
		"FIRECRAWL_API_KEY", "FIRECRAWL_KEY", "FIRECRAWL_API", "API_KEY_FIRECRAWL", "API_FIRECRAWL", "FIRECRAWL_PK", "FIRECRAWL_PUBLIC",
		"FIRECRAWL_API_SECRET", "FIRECRAWL_SECRET_KEY", "FIRECRAWL_API_SECRET_KEY", "FIRECRAWL_SECRET", "SECRET_FIRECRAWL", "API_SECRET_FIRECRAWL", "SK_FIRECRAWL", "FIRECRAWL_SK",
	); k != "" {
		return k, nil
	}

	// 3. Config file
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.APIKey != "" {
		return cfg.APIKey, nil
	}

	return "", fmt.Errorf("not authenticated — run: firecrawl auth set-key\nor set FIRECRAWL_API_KEY env var")
}

// isAuthCommand returns true if cmd is a child of the "auth" command.
func isAuthCommand(cmd *cobra.Command) bool {
	if cmd.Name() == "auth" {
		return true
	}
	p := cmd.Parent()
	for p != nil {
		if p.Name() == "auth" {
			return true
		}
		p = p.Parent()
	}
	return false
}

func maskOrEmpty(v string) string {
	if v == "" {
		return "(not set)"
	}
	if len(v) <= 8 {
		return "***"
	}
	return v[:4] + "..." + v[len(v)-4:]
}

func newClient() *client.Client {
	key, err := resolveAPIKey()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
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
