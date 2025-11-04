package root

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/peterldowns/localias/cmd/localias/shared"
)

var listFlags struct { //nolint:gochecknoglobals
	JSON *bool
}

var listCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:     "list",
	Aliases: []string{"l", "ls", "show"},
	Short:   "list all aliases",
	RunE:    listImpl,
}

func listImpl(_ *cobra.Command, _ []string) error {
	cfg := shared.Config()

	if *listFlags.JSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(cfg.Entries)
	}

	for _, entry := range cfg.Entries {
		fmt.Printf(
			"%s -> %s\n",
			color.New(color.FgBlue).Sprint(entry.Alias),
			color.New(color.FgWhite).Sprintf("%d", entry.Port),
		)
	}
	return nil
}

func init() {
	listFlags.JSON = listCmd.Flags().BoolP("json", "j", false, "output in JSON format")
	Command.AddCommand(listCmd)
}
