package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [device-id]",
	Short: "Remove a device-to-layout mapping",
	Long:  `Remove a mapping for the specified device ID.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	deviceID := args[0]
	fmt.Printf("Removing mapping for device: %s\n", deviceID)
	fmt.Println("(Not yet implemented)")
	// TODO: Implement remove mapping
	return nil
}
