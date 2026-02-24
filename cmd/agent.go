package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vincentmaurin/firecrawl-cli/client"
)

var agentCmd = &cobra.Command{
	Use:   "agent <prompt>",
	Short: "AI-powered autonomous extraction",
	Long: `Use an AI agent to browse and extract structured data from websites.

The agent interprets natural language instructions to navigate pages and extract
the information you need. Jobs run asynchronously (typically 2-5 minutes).

Examples:
  firecrawl agent "Find the pricing plans on this page" --urls https://example.com
  firecrawl agent "Extract all product names and prices" --urls https://shop.example.com --schema-file schema.json
  firecrawl agent "Get the author and title" --schema '{"type":"object","properties":{"author":{"type":"string"},"title":{"type":"string"}}}'
  firecrawl agent --status <job-id>
  firecrawl agent "Extract data" --urls https://example.com --wait --model spark-1-pro`,
	RunE: runAgent,
}

var (
	agentURLs         []string
	agentModel        string
	agentSchema       string
	agentSchemaFile   string
	agentMaxCredits   int
	agentWait         bool
	agentPollInterval int
	agentTimeout      int
	agentStatus       string
	agentOutput       string
	agentJSON         bool
)

func init() {
	rootCmd.AddCommand(agentCmd)

	f := agentCmd.Flags()
	f.StringSliceVar(&agentURLs, "urls", nil, "URLs for the agent to focus on")
	f.StringVar(&agentModel, "model", "", "Model: spark-1-mini (default), spark-1-pro")
	f.StringVar(&agentSchema, "schema", "", "JSON schema for structured output (inline)")
	f.StringVar(&agentSchemaFile, "schema-file", "", "Path to JSON schema file")
	f.IntVar(&agentMaxCredits, "max-credits", 0, "Maximum credits to spend")
	f.BoolVar(&agentWait, "wait", false, "Block until agent job completes")
	f.IntVar(&agentPollInterval, "poll-interval", 5, "Polling interval in seconds when --wait is set")
	f.IntVar(&agentTimeout, "timeout", 300, "Timeout in seconds when --wait is set")
	f.StringVar(&agentStatus, "status", "", "Check status of an existing agent job ID")
	f.StringVarP(&agentOutput, "output", "o", "", "Save result to file")
	f.BoolVar(&agentJSON, "json", false, "Output raw JSON response")
}

func runAgent(cmd *cobra.Command, args []string) error {
	c := newClient()

	// Status check mode
	if agentStatus != "" {
		status, err := c.AgentStatus(agentStatus)
		if err != nil {
			return err
		}
		if agentJSON {
			return printJSON(status)
		}
		fmt.Printf("Agent %s: %s\n", agentStatus, status.Status)
		if len(status.Data) > 0 {
			fmt.Println(string(status.Data))
		}
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("prompt argument required (or use --status <id>)")
	}

	prompt := strings.Join(args, " ")

	req := &client.AgentRequest{
		Prompt: prompt,
	}

	if len(agentURLs) > 0 {
		req.URLs = agentURLs
	}
	if agentModel != "" {
		req.Model = agentModel
	}
	if agentMaxCredits > 0 {
		req.MaxCredits = agentMaxCredits
	}

	// Load schema
	if agentSchemaFile != "" {
		data, err := os.ReadFile(agentSchemaFile)
		if err != nil {
			return fmt.Errorf("read schema file: %w", err)
		}
		req.Schema = data
	} else if agentSchema != "" {
		req.Schema = []byte(agentSchema)
	}

	job, err := c.AgentStart(req)
	if err != nil {
		return err
	}

	if agentJSON && !agentWait {
		return printJSON(job)
	}

	fmt.Printf("Agent job started: %s\n", job.ID)

	if !agentWait {
		fmt.Printf("Check status with: firecrawl agent --status %s\n", job.ID)
		return nil
	}

	// Poll until done
	poll := time.Duration(agentPollInterval) * time.Second
	deadline := time.Now().Add(time.Duration(agentTimeout) * time.Second)

	fmt.Fprint(os.Stderr, "Waiting for agent")
	for {
		time.Sleep(poll)
		fmt.Fprint(os.Stderr, ".")

		if time.Now().After(deadline) {
			fmt.Fprintln(os.Stderr)
			return fmt.Errorf("timed out waiting for agent job %s", job.ID)
		}

		status, err := c.AgentStatus(job.ID)
		if err != nil {
			return fmt.Errorf("poll agent status: %w", err)
		}

		if status.Status == "completed" || status.Status == "failed" {
			fmt.Fprintln(os.Stderr)
			if status.Status == "failed" {
				return fmt.Errorf("agent job failed: %s", status.Error)
			}

			if agentJSON {
				return printJSON(status)
			}

			output := string(status.Data)
			if agentOutput != "" {
				if err := os.WriteFile(agentOutput, status.Data, 0644); err != nil {
					return fmt.Errorf("write output file: %w", err)
				}
				fmt.Fprintf(os.Stderr, "Saved to %s\n", agentOutput)
				return nil
			}

			fmt.Println(output)
			return nil
		}
	}
}
