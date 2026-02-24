package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vincentmaurin/firecrawl-cli/client"
)

var mapCmd = &cobra.Command{
	Use:   "map <url>",
	Short: "Discover all URLs on a website",
	Long: `Map a website to discover all its URLs without fetching full content.

Examples:
  firecrawl map https://example.com
  firecrawl map https://example.com --limit 100
  firecrawl map https://example.com --search "blog"
  firecrawl map https://example.com --include-subdomains
  firecrawl map https://example.com --sitemap only
  firecrawl map https://example.com --ignore-query-parameters`,
	Args: cobra.ExactArgs(1),
	RunE: runMap,
}

var (
	mapLimit              int
	mapSearch             string
	mapIncludeSubdomains  bool
	mapIgnoreQueryParams  bool
	mapSitemap            string
	mapTimeout            int
	mapJSON               bool
)

func init() {
	rootCmd.AddCommand(mapCmd)

	f := mapCmd.Flags()
	f.IntVar(&mapLimit, "limit", 0, "Maximum URLs to return")
	f.StringVar(&mapSearch, "search", "", "Filter URLs by search query")
	f.BoolVar(&mapIncludeSubdomains, "include-subdomains", false, "Include subdomains in results")
	f.BoolVar(&mapIgnoreQueryParams, "ignore-query-parameters", false, "Deduplicate URLs by ignoring query params")
	f.StringVar(&mapSitemap, "sitemap", "", "Sitemap mode: include, skip, only")
	f.IntVar(&mapTimeout, "timeout", 0, "Request timeout in seconds")
	f.BoolVar(&mapJSON, "json", false, "Output raw JSON response")
}

func runMap(cmd *cobra.Command, args []string) error {
	c := newClient()

	req := &client.MapRequest{
		URL: args[0],
	}
	if mapLimit > 0 {
		req.Limit = mapLimit
	}
	if mapSearch != "" {
		req.Search = mapSearch
	}
	if mapIncludeSubdomains {
		req.IncludeSubdomains = true
	}
	if mapIgnoreQueryParams {
		req.IgnoreQueryParams = true
	}
	if mapSitemap != "" {
		req.Sitemap = mapSitemap
	}
	if mapTimeout > 0 {
		req.Timeout = mapTimeout
	}

	resp, err := c.Map(req)
	if err != nil {
		return err
	}

	if mapJSON {
		return printJSON(resp)
	}

	fmt.Printf("Found %d URL(s) on %s\n\n", len(resp.Links), args[0])
	for _, link := range resp.Links {
		fmt.Println(link)
	}

	return nil
}
