package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.firecrawl.dev"

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func New(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

func (c *Client) do(method, path string, body any, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// Scrape fetches and extracts content from a single URL.
func (c *Client) Scrape(req *ScrapeRequest) (*ScrapeResponse, error) {
	var resp ScrapeResponse
	if err := c.do("POST", "/v1/scrape", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Search performs a web search.
func (c *Client) Search(req *SearchRequest) (*SearchResponse, error) {
	var resp SearchResponse
	if err := c.do("POST", "/v1/search", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Map discovers all URLs on a site.
func (c *Client) Map(req *MapRequest) (*MapResponse, error) {
	var resp MapResponse
	if err := c.do("POST", "/v1/map", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CrawlStart initiates an async crawl job and returns the job ID.
func (c *Client) CrawlStart(req *CrawlRequest) (*CrawlJob, error) {
	var resp CrawlJob
	if err := c.do("POST", "/v1/crawl", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CrawlStatus returns the current status of a crawl job.
func (c *Client) CrawlStatus(id string) (*CrawlStatus, error) {
	var resp CrawlStatus
	if err := c.do("GET", "/v1/crawl/"+id, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AgentStart initiates an async agent extraction job.
func (c *Client) AgentStart(req *AgentRequest) (*AgentJob, error) {
	var resp AgentJob
	if err := c.do("POST", "/v1/agent", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AgentStatus returns the current status of an agent job.
func (c *Client) AgentStatus(id string) (*AgentStatus, error) {
	var resp AgentStatus
	if err := c.do("GET", "/v1/agent/"+id, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Credits returns the current credit usage.
func (c *Client) Credits() (*CreditUsage, error) {
	var resp CreditUsage
	if err := c.do("GET", "/v1/team/credit-usage", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
