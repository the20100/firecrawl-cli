package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/firecrawl-cli/client"
)

var scrapeCmd = &cobra.Command{
	Use:   "scrape <url>",
	Short: "Scrape a webpage and extract its content",
	Long: `Scrape a URL and extract content in various formats.

Supports markdown, HTML, links, screenshots, and structured JSON extraction
via a custom schema (--schema or --schema-file).

Examples:
  firecrawl scrape https://example.com
  firecrawl scrape https://example.com --format markdown,html
  firecrawl scrape https://example.com --format json --schema '{"type":"object","properties":{"title":{"type":"string"}}}'
  firecrawl scrape https://example.com --schema-file schema.json --format markdown,json
  firecrawl scrape https://example.com --only-main-content --parser pdf
  firecrawl scrape https://example.com --max-age 86400000 --wait-for 2000
  firecrawl scrape https://example.com --exclude-tags nav,footer --output result.md`,
	Args: cobra.ExactArgs(1),
	RunE: runScrape,
}

var (
	scrapeFormats      []string
	scrapeOnlyMain     bool
	scrapeIncludeTags  []string
	scrapeExcludeTags  []string
	scrapeWaitFor      int
	scrapeTimeout      int
	scrapeMaxAge       int64
	scrapeParsers      []string
	scrapeSchema       string
	scrapeSchemaFile   string
	scrapeOutput       string
	scrapeJSON         bool
	scrapePretty       bool
	scrapeTiming       bool
)

func init() {
	rootCmd.AddCommand(scrapeCmd)

	f := scrapeCmd.Flags()
	f.StringSliceVarP(&scrapeFormats, "format", "f", []string{"markdown"}, "Output formats: markdown,html,rawHtml,links,screenshot,json")
	f.BoolVar(&scrapeOnlyMain, "only-main-content", false, "Remove navigation, footers, and sidebars")
	f.StringSliceVar(&scrapeIncludeTags, "include-tags", nil, "HTML tags to include (e.g. article,main)")
	f.StringSliceVar(&scrapeExcludeTags, "exclude-tags", nil, "HTML tags to exclude (e.g. nav,footer)")
	f.IntVar(&scrapeWaitFor, "wait-for", 0, "Wait N milliseconds for JavaScript rendering")
	f.IntVar(&scrapeTimeout, "timeout", 0, "Request timeout in milliseconds")
	f.Int64Var(&scrapeMaxAge, "max-age", 0, "Max cache age in milliseconds")
	f.StringSliceVar(&scrapeParsers, "parser", nil, "Additional parsers: pdf")
	f.StringVar(&scrapeSchema, "schema", "", "JSON schema for structured extraction (inline JSON)")
	f.StringVar(&scrapeSchemaFile, "schema-file", "", "Path to JSON schema file for structured extraction")
	f.StringVarP(&scrapeOutput, "output", "o", "", "Save output to file")
	f.BoolVar(&scrapeJSON, "json", false, "Output raw JSON response")
	f.BoolVar(&scrapePretty, "pretty", false, "Pretty-print JSON output")
	f.BoolVar(&scrapeTiming, "timing", false, "Show request timing")
}

func runScrape(cmd *cobra.Command, args []string) error {
	c := newClient()
	url := args[0]

	req := &client.ScrapeRequest{
		URL: url,
	}

	// Parse schema
	var schemaJSON json.RawMessage
	if scrapeSchemaFile != "" {
		data, err := os.ReadFile(scrapeSchemaFile)
		if err != nil {
			return fmt.Errorf("read schema file: %w", err)
		}
		schemaJSON = json.RawMessage(data)
	} else if scrapeSchema != "" {
		schemaJSON = json.RawMessage(scrapeSchema)
	}

	// Build formats list
	var formats []interface{}
	for _, f := range scrapeFormats {
		f = strings.TrimSpace(f)
		if f == "json" && len(schemaJSON) > 0 {
			formats = append(formats, client.JSONFormat{
				Type:   "json",
				Schema: schemaJSON,
			})
		} else {
			formats = append(formats, f)
		}
	}
	// If schema provided but json not in formats, append json format automatically
	if len(schemaJSON) > 0 {
		hasJSON := false
		for _, f := range scrapeFormats {
			if strings.TrimSpace(f) == "json" {
				hasJSON = true
				break
			}
		}
		if !hasJSON {
			formats = append(formats, client.JSONFormat{
				Type:   "json",
				Schema: schemaJSON,
			})
		}
	}
	req.Formats = formats

	if scrapeOnlyMain {
		t := true
		req.OnlyMainContent = &t
	}
	if len(scrapeIncludeTags) > 0 {
		req.IncludeTags = scrapeIncludeTags
	}
	if len(scrapeExcludeTags) > 0 {
		req.ExcludeTags = scrapeExcludeTags
	}
	if scrapeWaitFor > 0 {
		req.WaitFor = scrapeWaitFor
	}
	if scrapeTimeout > 0 {
		req.Timeout = scrapeTimeout
	}
	if scrapeMaxAge > 0 {
		req.MaxAge = scrapeMaxAge
	}
	if len(scrapeParsers) > 0 {
		req.Parsers = scrapeParsers
	}

	resp, err := c.Scrape(req)
	if err != nil {
		return err
	}

	if scrapeJSON || scrapePretty {
		return printJSON(resp)
	}

	output := buildScrapeOutput(resp)

	if scrapeOutput != "" {
		if err := os.WriteFile(scrapeOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Saved to %s\n", scrapeOutput)
		return nil
	}

	fmt.Print(output)
	return nil
}

func buildScrapeOutput(resp *client.ScrapeResponse) string {
	var sb strings.Builder

	if resp.Data.Metadata != nil {
		m := resp.Data.Metadata
		if m.Title != "" {
			sb.WriteString("# " + m.Title + "\n")
			sb.WriteString("URL: " + m.SourceURL + "\n\n")
		}
	}

	if resp.Data.Markdown != "" {
		sb.WriteString(resp.Data.Markdown)
		sb.WriteString("\n")
	}

	if resp.Data.HTML != "" {
		if resp.Data.Markdown != "" {
			sb.WriteString("\n---HTML---\n")
		}
		sb.WriteString(resp.Data.HTML)
		sb.WriteString("\n")
	}

	if len(resp.Data.Links) > 0 {
		sb.WriteString("\n---LINKS---\n")
		for _, l := range resp.Data.Links {
			sb.WriteString(l + "\n")
		}
	}

	if len(resp.Data.Extract) > 0 && string(resp.Data.Extract) != "null" {
		sb.WriteString("\n---EXTRACTED JSON---\n")
		pretty, err := json.MarshalIndent(resp.Data.Extract, "", "  ")
		if err == nil {
			sb.WriteString(string(pretty) + "\n")
		} else {
			sb.WriteString(string(resp.Data.Extract) + "\n")
		}
	}

	return sb.String()
}
