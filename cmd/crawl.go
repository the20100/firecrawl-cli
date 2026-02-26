package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/the20100/firecrawl-cli/client"
)

var crawlCmd = &cobra.Command{
	Use:   "crawl <url>",
	Short: "Crawl a website across multiple pages",
	Long: `Crawl a website by following links across multiple pages.

The crawl runs asynchronously and returns a job ID. Use --wait to block until
completion, or use --status <id> to check on an existing job.

Examples:
  firecrawl crawl https://example.com
  firecrawl crawl https://example.com --wait --limit 50
  firecrawl crawl https://example.com --max-depth 3 --allow-subdomains
  firecrawl crawl https://example.com --include-paths /blog,/docs
  firecrawl crawl https://example.com --exclude-paths /admin,/login
  firecrawl crawl --status <job-id>`,
	RunE: runCrawl,
}

var (
	crawlWait              bool
	crawlProgress          bool
	crawlLimit             int
	crawlMaxDepth          int
	crawlIncludePaths      []string
	crawlExcludePaths      []string
	crawlAllowSubdomains   bool
	crawlAllowExternal     bool
	crawlIgnoreQueryParams bool
	crawlDelay             int
	crawlMaxConcurrency    int
	crawlPollInterval      int
	crawlTimeout           int
	crawlScrapeFormats     []string
	crawlStatus            string
	crawlJSON              bool
)

func init() {
	rootCmd.AddCommand(crawlCmd)

	f := crawlCmd.Flags()
	f.BoolVar(&crawlWait, "wait", false, "Block until crawl completes")
	f.BoolVar(&crawlProgress, "progress", false, "Show crawl progress while waiting")
	f.IntVar(&crawlLimit, "limit", 0, "Maximum pages to crawl")
	f.IntVar(&crawlMaxDepth, "max-depth", 0, "Maximum link depth to follow")
	f.StringSliceVar(&crawlIncludePaths, "include-paths", nil, "Only crawl paths matching these patterns")
	f.StringSliceVar(&crawlExcludePaths, "exclude-paths", nil, "Skip paths matching these patterns")
	f.BoolVar(&crawlAllowSubdomains, "allow-subdomains", false, "Include subdomains")
	f.BoolVar(&crawlAllowExternal, "allow-external-links", false, "Follow external links")
	f.BoolVar(&crawlIgnoreQueryParams, "ignore-query-parameters", false, "Deduplicate URLs ignoring query params")
	f.IntVar(&crawlDelay, "delay", 0, "Delay between requests in milliseconds")
	f.IntVar(&crawlMaxConcurrency, "max-concurrency", 0, "Maximum concurrent requests")
	f.IntVar(&crawlPollInterval, "poll-interval", 2, "Polling interval in seconds when --wait is set")
	f.IntVar(&crawlTimeout, "timeout", 0, "Request timeout in seconds")
	f.StringSliceVar(&crawlScrapeFormats, "format", []string{"markdown"}, "Scrape formats for each page")
	f.StringVar(&crawlStatus, "status", "", "Check status of an existing crawl job ID")
	f.BoolVar(&crawlJSON, "json", false, "Output raw JSON response")
}

func runCrawl(cmd *cobra.Command, args []string) error {
	c := newClient()

	// Status check mode
	if crawlStatus != "" {
		status, err := c.CrawlStatus(crawlStatus)
		if err != nil {
			return err
		}
		if crawlJSON {
			return printJSON(status)
		}
		fmt.Printf("Crawl %s: %s (%d/%d pages)\n", crawlStatus, status.Status, status.Completed, status.Total)
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("URL argument required (or use --status <id>)")
	}

	req := &client.CrawlRequest{
		URL: args[0],
	}
	if crawlLimit > 0 {
		req.Limit = crawlLimit
	}
	if crawlMaxDepth > 0 {
		req.MaxDepth = crawlMaxDepth
	}
	if len(crawlIncludePaths) > 0 {
		req.IncludePaths = crawlIncludePaths
	}
	if len(crawlExcludePaths) > 0 {
		req.ExcludePaths = crawlExcludePaths
	}
	if crawlAllowSubdomains {
		req.AllowSubdomains = true
	}
	if crawlAllowExternal {
		req.AllowExternalLinks = true
	}
	if crawlIgnoreQueryParams {
		req.IgnoreQueryParams = true
	}
	if crawlDelay > 0 {
		req.Delay = crawlDelay
	}
	if crawlMaxConcurrency > 0 {
		req.MaxConcurrency = crawlMaxConcurrency
	}

	var formats []interface{}
	for _, f := range crawlScrapeFormats {
		formats = append(formats, f)
	}
	req.ScrapeOptions = &client.ScrapeRequest{Formats: formats}

	job, err := c.CrawlStart(req)
	if err != nil {
		return err
	}

	if crawlJSON {
		return printJSON(job)
	}

	fmt.Printf("Crawl started: %s\n", job.ID)

	if !crawlWait {
		fmt.Printf("Check status with: firecrawl crawl --status %s\n", job.ID)
		return nil
	}

	// Poll until done
	poll := time.Duration(crawlPollInterval) * time.Second
	for {
		time.Sleep(poll)
		status, err := c.CrawlStatus(job.ID)
		if err != nil {
			return fmt.Errorf("poll crawl status: %w", err)
		}

		if crawlProgress {
			fmt.Fprintf(os.Stderr, "\rProgress: %d/%d pages (%s)   ", status.Completed, status.Total, status.Status)
		}

		if status.Status == "completed" || status.Status == "failed" {
			if crawlProgress {
				fmt.Println()
			}
			fmt.Printf("Crawl %s: %d pages crawled\n", status.Status, status.Completed)
			if status.Status == "completed" && len(status.Data) > 0 {
				for i, page := range status.Data {
					if page.Metadata != nil {
						fmt.Printf("[%d] %s\n", i+1, page.Metadata.SourceURL)
					}
				}
			}
			break
		}
	}

	return nil
}
