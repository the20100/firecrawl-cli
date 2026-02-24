# firecrawl-cli

A Go CLI for the [Firecrawl API](https://firecrawl.dev). Supports scraping, searching, mapping, crawling, and AI-powered extraction — including **structured JSON extraction via custom schemas**.

## Installation

```bash
cd firecrawl-cli
go build -o firecrawl .
mv firecrawl /usr/local/bin/  # or anywhere in your PATH
```

## Authentication

```bash
export FIRECRAWL_API_KEY=fc-xxxxxxxxxxxxxxxx
```

Or pass `--api-key <key>` to any command.

For self-hosted instances:

```bash
export FIRECRAWL_API_URL=http://localhost:3002
```

---

## Commands

### `scrape` — Extract content from a URL

```bash
firecrawl scrape <url> [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-f, --format` | `markdown` | Output formats: `markdown`, `html`, `rawHtml`, `links`, `screenshot`, `json` |
| `--schema <json>` | | Inline JSON schema for structured extraction |
| `--schema-file <path>` | | Path to a JSON schema file |
| `--only-main-content` | | Strip navigation, footers, sidebars |
| `--include-tags <tags>` | | HTML tags to include (e.g. `article,main`) |
| `--exclude-tags <tags>` | | HTML tags to exclude (e.g. `nav,footer`) |
| `--wait-for <ms>` | | Wait N ms for JS rendering |
| `--timeout <ms>` | | Request timeout in milliseconds |
| `--max-age <ms>` | | Max cache age in milliseconds |
| `--parser <parsers>` | | Additional parsers: `pdf` |
| `-o, --output <path>` | | Save output to file |
| `--json` | | Raw JSON response |
| `--pretty` | | Pretty-print JSON |
| `--timing` | | Show request timing |

**Examples:**

```bash
# Basic markdown scrape
firecrawl scrape https://example.com

# Multiple formats
firecrawl scrape https://example.com --format markdown,html,links

# Structured JSON extraction — inline schema
firecrawl scrape https://example.com \
  --schema '{"type":"object","properties":{"company_name":{"type":"string"},"company_description":{"type":"string"}}}'

# Structured JSON extraction — schema file + markdown
firecrawl scrape https://example.com \
  --schema-file schema.json \
  --format markdown,json

# Equivalent to the advanced curl pattern (PDF parser + maxAge + schema)
firecrawl scrape https://example.com \
  --parser pdf \
  --max-age 172800000 \
  --schema-file schema.json

# Clean content, save to file
firecrawl scrape https://example.com --only-main-content --output result.md

# Wait for JS rendering, filter tags
firecrawl scrape https://spa.example.com --wait-for 2000 --exclude-tags nav,footer
```

> **How the schema works:** when `--schema` or `--schema-file` is provided, the CLI automatically wraps it as `{"type":"json","schema":{...}}` in the API request — the same format as the `curl` call with an explicit JSON format object. No manual wrapping needed.

---

### `search` — Search the web

```bash
firecrawl search <query> [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-n, --limit` | `5` | Number of results (1–100) |
| `--lang` | | Language code (e.g. `en`, `fr`) |
| `--country` | | Country code (e.g. `us`, `fr`) |
| `--location` | | Geo-targeted location |
| `--tbs` | | Time filter: `qdr:h`, `qdr:d`, `qdr:w`, `qdr:m`, `qdr:y` |
| `--sources` | | `web`, `images`, `news` |
| `--categories` | | `github`, `research`, `pdf` |
| `--timeout` | | Request timeout in milliseconds |
| `--scrape` | | Scrape each result URL for full content |
| `--scrape-formats` | `markdown` | Formats for scraped content |
| `--json` | | Raw JSON response |

**Examples:**

```bash
firecrawl search "golang web scraping"
firecrawl search "AI news" --limit 10 --sources news --tbs qdr:w
firecrawl search "LLM papers" --categories research --limit 20
firecrawl search "golang tutorials" --scrape --scrape-formats markdown
```

---

### `map` — Discover all URLs on a site

```bash
firecrawl map <url> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--limit <n>` | Maximum URLs to return |
| `--search <query>` | Filter URLs by search query |
| `--include-subdomains` | Include subdomains |
| `--ignore-query-parameters` | Deduplicate by ignoring query params |
| `--sitemap <mode>` | `include`, `skip`, or `only` |
| `--timeout <s>` | Request timeout in seconds |
| `--json` | Raw JSON response |

**Examples:**

```bash
firecrawl map https://example.com
firecrawl map https://example.com --search "blog" --limit 100
firecrawl map https://example.com --sitemap only
firecrawl map https://example.com --include-subdomains --json
```

---

### `crawl` — Multi-page crawl

Crawls a site by following links. Runs **asynchronously** and returns a job ID. Use `--wait` to block until done.

```bash
firecrawl crawl <url> [flags]
firecrawl crawl --status <job-id>
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--wait` | | Block until crawl completes |
| `--progress` | | Show live progress |
| `--limit <n>` | | Max pages to crawl |
| `--max-depth <n>` | | Max link depth |
| `--include-paths <list>` | | Only crawl matching paths |
| `--exclude-paths <list>` | | Skip matching paths |
| `--allow-subdomains` | | Include subdomains |
| `--allow-external-links` | | Follow external links |
| `--ignore-query-parameters` | | Deduplicate by ignoring query params |
| `--delay <ms>` | | Delay between requests |
| `--max-concurrency <n>` | | Max parallel requests |
| `--format <list>` | `markdown` | Scrape formats per page |
| `--poll-interval <s>` | `2` | Poll interval when `--wait` |
| `--status <job-id>` | | Check an existing job |
| `--json` | | Raw JSON response |

**Examples:**

```bash
# Start crawl, get job ID
firecrawl crawl https://example.com

# Crawl and wait for completion
firecrawl crawl https://example.com --wait --limit 100 --progress

# Scoped crawl
firecrawl crawl https://example.com --include-paths /blog --max-depth 3

# Check a running job
firecrawl crawl --status abc123def456
```

---

### `agent` — AI-powered extraction

Uses an AI agent to browse and extract data from natural language instructions. Runs **asynchronously** (typically 2–5 min). Use `--wait` to block for results.

```bash
firecrawl agent <prompt> [flags]
firecrawl agent --status <job-id>
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--urls <list>` | | URLs to focus on |
| `--model` | | `spark-1-mini` or `spark-1-pro` |
| `--schema <json>` | | Inline JSON schema for structured output |
| `--schema-file <path>` | | Path to JSON schema file |
| `--max-credits <n>` | | Credit spend ceiling |
| `--wait` | | Block until job completes |
| `--poll-interval <s>` | `5` | Poll interval when `--wait` |
| `--timeout <s>` | `300` | Max wait time |
| `--status <job-id>` | | Check an existing job |
| `-o, --output <path>` | | Save result to file |
| `--json` | | Raw JSON response |

**Examples:**

```bash
# Extract pricing info
firecrawl agent "Find all pricing tiers" --urls https://example.com/pricing --wait

# Structured output with schema
firecrawl agent "Extract company info" \
  --urls https://example.com \
  --schema '{"type":"object","properties":{"company_name":{"type":"string"},"company_description":{"type":"string"}}}' \
  --wait

# Schema file, save to JSON
firecrawl agent "Get all products" \
  --urls https://shop.example.com \
  --schema-file schema.json \
  --wait \
  --output products.json

# Check job status
firecrawl agent --status abc123def456
```

---

### `credit-usage` — Account credit usage

```bash
firecrawl credit-usage
firecrawl credit-usage --json
```

---

## Global flags

| Flag | Description |
|------|-------------|
| `--api-key <key>` | API key (or `FIRECRAWL_API_KEY` env var) |
| `--api-url <url>` | API base URL (or `FIRECRAWL_API_URL` env var) |

---

## Project structure

```
firecrawl-cli/
├── main.go
├── go.mod
├── client/
│   ├── types.go    # API request/response types
│   └── client.go   # HTTP client
└── cmd/
    ├── root.go     # Global config, newClient()
    ├── scrape.go
    ├── search.go
    ├── crawl.go
    ├── map.go
    ├── agent.go
    └── credits.go
```
