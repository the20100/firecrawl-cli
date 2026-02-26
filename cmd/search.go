package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/firecrawl-cli/client"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search the web and retrieve results",
	Long: `Search the web using Firecrawl's search API.

Optionally scrape result pages to get full content.

Examples:
  firecrawl search "golang web scraping"
  firecrawl search "AI news" --limit 10 --sources news
  firecrawl search "golang" --scrape --scrape-formats markdown
  firecrawl search "papers" --categories research --tbs qdr:w
  firecrawl search "local news" --location "Paris, France"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSearch,
}

var (
	searchLimit         int
	searchLang          string
	searchCountry       string
	searchLocation      string
	searchTbs           string
	searchSources       []string
	searchCategories    []string
	searchTimeout       int
	searchScrape        bool
	searchScrapeFormats []string
	searchJSON          bool
)

func init() {
	rootCmd.AddCommand(searchCmd)

	f := searchCmd.Flags()
	f.IntVarP(&searchLimit, "limit", "n", 5, "Number of results (1-100)")
	f.StringVar(&searchLang, "lang", "", "Language code (e.g. en, fr)")
	f.StringVar(&searchCountry, "country", "", "Country code (e.g. us, fr)")
	f.StringVar(&searchLocation, "location", "", "Location for geo-targeted results")
	f.StringVar(&searchTbs, "tbs", "", "Time filter: qdr:h, qdr:d, qdr:w, qdr:m, qdr:y")
	f.StringSliceVar(&searchSources, "sources", nil, "Sources: web, images, news")
	f.StringSliceVar(&searchCategories, "categories", nil, "Categories: github, research, pdf")
	f.IntVar(&searchTimeout, "timeout", 0, "Request timeout in milliseconds")
	f.BoolVar(&searchScrape, "scrape", false, "Scrape each result URL for full content")
	f.StringSliceVar(&searchScrapeFormats, "scrape-formats", []string{"markdown"}, "Formats for scraped content")
	f.BoolVar(&searchJSON, "json", false, "Output raw JSON response")
}

func runSearch(cmd *cobra.Command, args []string) error {
	c := newClient()

	query := strings.Join(args, " ")

	req := &client.SearchRequest{
		Query: query,
		Limit: searchLimit,
	}
	if searchLang != "" {
		req.Lang = searchLang
	}
	if searchCountry != "" {
		req.Country = searchCountry
	}
	if searchLocation != "" {
		req.Location = searchLocation
	}
	if searchTbs != "" {
		req.Tbs = searchTbs
	}
	if len(searchSources) > 0 {
		req.Sources = searchSources
	}
	if len(searchCategories) > 0 {
		req.Categories = searchCategories
	}
	if searchTimeout > 0 {
		req.Timeout = searchTimeout
	}
	if searchScrape {
		var formats []interface{}
		for _, f := range searchScrapeFormats {
			formats = append(formats, f)
		}
		req.ScrapeOptions = &client.ScrapeRequest{
			Formats: formats,
		}
	}

	resp, err := c.Search(req)
	if err != nil {
		return err
	}

	if searchJSON {
		return printJSON(resp)
	}

	fmt.Printf("Search results for: %q\n", query)
	if len(resp.Data) == 0 {
		fmt.Println("No results found.")
		return nil
	}
	fmt.Printf("Found %d result(s)\n\n", len(resp.Data))

	for i, r := range resp.Data {
		fmt.Printf("[%d] %s\n", i+1, r.Title)
		fmt.Printf("    URL: %s\n", r.URL)
		if r.Description != "" {
			fmt.Printf("    %s\n", r.Description)
		}
		if r.Markdown != "" {
			preview := r.Markdown
			if len(preview) > 300 {
				preview = preview[:300] + "..."
			}
			fmt.Printf("\n%s\n", preview)
		}
		fmt.Println()
	}

	return nil
}
