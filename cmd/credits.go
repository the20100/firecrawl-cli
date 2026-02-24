package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var creditsCmd = &cobra.Command{
	Use:   "credit-usage",
	Short: "Show current credit usage",
	Long: `Display the current credit consumption for your Firecrawl account.

Examples:
  firecrawl credit-usage
  firecrawl credit-usage --json`,
	Args: cobra.NoArgs,
	RunE: runCredits,
}

var creditsJSON bool

func init() {
	rootCmd.AddCommand(creditsCmd)
	creditsCmd.Flags().BoolVar(&creditsJSON, "json", false, "Output raw JSON response")
}

func runCredits(cmd *cobra.Command, args []string) error {
	c := newClient()

	resp, err := c.Credits()
	if err != nil {
		return err
	}

	if creditsJSON {
		return printJSON(resp)
	}

	fmt.Printf("Credits used: %d\n", resp.Data.Credits)
	return nil
}
