package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kerns/hitmaker/internal/tui"
)

// newFrameCommand prints a single rendered TUI frame from fixture data and
// exits. It is hidden — its only purpose is regenerating documentation
// screenshots deterministically (pipe it through a terminal-to-image tool).
func newFrameCommand(root *rootOptions) *cobra.Command {
	var width, height int
	cmd := &cobra.Command{
		Use:    "frame [dashboard|config]",
		Short:  "Render one sample TUI frame (for docs screenshots)",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			which := "dashboard"
			if len(args) > 0 {
				which = args[0]
			}
			switch which {
			case "config":
				fmt.Println(tui.SampleConfig(width, height))
			default:
				fmt.Println(tui.SampleDashboard(width, height))
			}
			return nil
		},
	}
	cmd.Flags().IntVarP(&width, "width", "W", 92, "frame width")
	cmd.Flags().IntVarP(&height, "height", "H", 20, "frame height")
	return cmd
}
