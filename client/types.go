package client

import "encoding/json"

// --- Scrape API ---

type ScrapeFormat interface{}

// JSONFormat represents the structured JSON extraction format with a schema.
type JSONFormat struct {
	Type   string          `json:"type"`
	Schema json.RawMessage `json:"schema"`
}

type ScrapeRequest struct {
	URL             string        `json:"url"`
	Formats         []interface{} `json:"formats,omitempty"`
	OnlyMainContent *bool         `json:"onlyMainContent,omitempty"`
	IncludeTags     []string      `json:"includeTags,omitempty"`
	ExcludeTags     []string      `json:"excludeTags,omitempty"`
	WaitFor         int           `json:"waitFor,omitempty"`
	Timeout         int           `json:"timeout,omitempty"`
	MaxAge          int64         `json:"maxAge,omitempty"`
	Parsers         []string      `json:"parsers,omitempty"`
}

type ScrapeData struct {
	Markdown   string          `json:"markdown,omitempty"`
	HTML       string          `json:"html,omitempty"`
	RawHTML    string          `json:"rawHtml,omitempty"`
	Links      []string        `json:"links,omitempty"`
	Screenshot string          `json:"screenshot,omitempty"`
	Extract    json.RawMessage `json:"extract,omitempty"`
	Metadata   *ScrapeMetadata `json:"metadata,omitempty"`
}

type ScrapeMetadata struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Language    string `json:"language,omitempty"`
	SourceURL   string `json:"sourceURL,omitempty"`
	StatusCode  int    `json:"statusCode,omitempty"`
}

type ScrapeResponse struct {
	Success bool       `json:"success"`
	Data    ScrapeData `json:"data"`
}

// --- Search API ---

type SearchRequest struct {
	Query          string   `json:"query"`
	Limit          int      `json:"limit,omitempty"`
	Lang           string   `json:"lang,omitempty"`
	Country        string   `json:"country,omitempty"`
	Location       string   `json:"location,omitempty"`
	Tbs            string   `json:"tbs,omitempty"`
	Sources        []string `json:"sources,omitempty"`
	Categories     []string `json:"categories,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
	ScrapeOptions  *ScrapeRequest `json:"scrapeOptions,omitempty"`
}

type SearchResult struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Markdown    string `json:"markdown,omitempty"`
}

type SearchResponse struct {
	Success bool           `json:"success"`
	Data    []SearchResult `json:"data"`
}

// --- Map API ---

type MapRequest struct {
	URL                  string `json:"url"`
	Search               string `json:"search,omitempty"`
	Limit                int    `json:"limit,omitempty"`
	IncludeSubdomains    bool   `json:"includeSubdomains,omitempty"`
	IgnoreQueryParams    bool   `json:"ignoreQueryParams,omitempty"`
	Sitemap              string `json:"sitemap,omitempty"` // include, skip, only
	Timeout              int    `json:"timeout,omitempty"`
}

type MapResponse struct {
	Success bool     `json:"success"`
	Links   []string `json:"links"`
}

// --- Crawl API ---

type CrawlRequest struct {
	URL                  string        `json:"url"`
	Limit                int           `json:"limit,omitempty"`
	MaxDepth             int           `json:"maxDepth,omitempty"`
	IncludePaths         []string      `json:"includePaths,omitempty"`
	ExcludePaths         []string      `json:"excludePaths,omitempty"`
	AllowSubdomains      bool          `json:"allowSubdomains,omitempty"`
	AllowExternalLinks   bool          `json:"allowExternalLinks,omitempty"`
	IgnoreQueryParams    bool          `json:"ignoreQueryParams,omitempty"`
	Delay                int           `json:"delay,omitempty"`
	MaxConcurrency       int           `json:"maxConcurrency,omitempty"`
	ScrapeOptions        *ScrapeRequest `json:"scrapeOptions,omitempty"`
}

type CrawlJob struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	URL     string `json:"url,omitempty"`
}

type CrawlStatus struct {
	Status    string       `json:"status"`
	Total     int          `json:"total"`
	Completed int          `json:"completed"`
	Data      []ScrapeData `json:"data,omitempty"`
}

// --- Agent API ---

type AgentRequest struct {
	Prompt      string          `json:"prompt"`
	URLs        []string        `json:"urls,omitempty"`
	Model       string          `json:"model,omitempty"`
	Schema      json.RawMessage `json:"schema,omitempty"`
	MaxCredits  int             `json:"maxCredits,omitempty"`
}

type AgentJob struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
}

type AgentStatus struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// --- Credits API ---

type CreditUsage struct {
	Success bool `json:"success"`
	Data    struct {
		Credits int `json:"credits"`
	} `json:"data"`
}

// --- Error ---

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return "firecrawl API error " + itoa(e.StatusCode) + ": " + e.Body
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	b := make([]byte, 0, 10)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
