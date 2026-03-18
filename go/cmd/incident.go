package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var incidentCmd = &cobra.Command{
	Use:   "incident",
	Short: "Manage data quality incidents",
}

var incidentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List incidents",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		page, err := client.ListIncidents(0, 100)
		if err != nil {
			return err
		}
		if len(page.Content) == 0 {
			fmt.Println(output.Dim.Render("  No incidents found."))
			return nil
		}

		// Client-side filters
		statusFilter, _ := cmd.Flags().GetString("status")
		datasetFilter, _ := cmd.Flags().GetString("dataset")

		rows := make([]map[string]string, 0, len(page.Content))
		for _, inc := range page.Content {
			if statusFilter != "" && !strings.EqualFold(inc.Status, statusFilter) {
				continue
			}
			if datasetFilter != "" && inc.Dataset.ID != datasetFilter && inc.Dataset.Name != datasetFilter {
				continue
			}
			rows = append(rows, map[string]string{
				"number":   strconv.Itoa(inc.Number),
				"name":     inc.DisplayName(),
				"severity": inc.Severity,
				"status":   inc.Status,
				"cloud url": inc.CloudURL,
			})
		}

		if len(rows) == 0 {
			fmt.Println(output.Dim.Render("  No incidents match the given filters."))
			return nil
		}

		cols := []string{"number", "name", "severity", "status", "cloud url"}
		output.Render(rows, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var incidentGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show incident details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		inc, err := client.GetIncident(args[0])
		if err != nil {
			return err
		}
		item := map[string]string{
			"id":               inc.ID,
			"number":           strconv.Itoa(inc.Number),
			"name":             inc.DisplayName(),
			"description":      inc.Description,
			"severity":         inc.Severity,
			"status":           inc.Status,
			"resolution notes": inc.ResolutionNotes,
			"cloud url":        inc.CloudURL,
		}
		keys := []string{"id", "number", "name", "description", "severity", "status", "resolution notes", "cloud url"}
		output.RenderOne(item, keys, GCtx)
		return nil
	},
}

var incidentUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an incident",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		var req api.UpdateIncidentRequest
		if v, _ := cmd.Flags().GetString("title"); v != "" {
			req.Title = &v
			req.Name = &v
		}
		if v, _ := cmd.Flags().GetString("severity"); v != "" {
			req.Severity = &v
		}
		if v, _ := cmd.Flags().GetString("description"); v != "" {
			req.Description = &v
		}
		if v, _ := cmd.Flags().GetString("assigned-to"); v != "" {
			req.AssignedTo = &v
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			req.Status = &v
		}

		inc, err := client.UpdateIncident(args[0], req)
		if err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Incident #%d updated.", inc.Number), GCtx)
		return nil
	},
}

func init() {
	incidentListCmd.Flags().String("status", "", "Filter by status: reported|investigating|fixing|resolved")
	incidentListCmd.Flags().String("dataset", "", "Filter by dataset ID")

	incidentUpdateCmd.Flags().String("title", "", "New title")
	incidentUpdateCmd.Flags().String("severity", "", "Severity: minor|major|critical")
	incidentUpdateCmd.Flags().String("description", "", "Description")
	incidentUpdateCmd.Flags().String("assigned-to", "", "Assigned user email")
	incidentUpdateCmd.Flags().String("status", "", "Status: reported|investigating|fixing|resolved")

	incidentCmd.AddCommand(incidentListCmd, incidentGetCmd, incidentUpdateCmd)
}
