package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all device-to-layout mappings",
	Long:  `Display all configured mappings between devices and keyboard layouts.`,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	fmt.Println("Current mappings:")
	fmt.Println("(Not yet implemented)")
	// TODO: Implement list mappings
	return nil
}
