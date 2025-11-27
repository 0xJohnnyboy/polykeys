package commands

import (
	"context"
	"fmt"

	"github.com/0xJohnnyboy/polykeys/internal/infrastructure"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all device-to-layout mappings",
	Long:  `Display all configured mappings between devices and keyboard layouts.`,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	// Initialize app
	app, err := infrastructure.NewApp()
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	ctx := context.Background()

	// Try to load mappings from config first
	if err := app.ManageMappingsUC.LoadFromConfig(ctx); err != nil {
		// Config might not exist yet, that's okay
		fmt.Println("No configuration file found. Use 'polykeys add' to create mappings.")
		return nil
	}

	// Get all mappings
	mappings, err := app.ManageMappingsUC.ListMappings(ctx)
	if err != nil {
		return fmt.Errorf("failed to list mappings: %w", err)
	}

	if len(mappings) == 0 {
		fmt.Println("No mappings configured.")
		fmt.Println("Use 'polykeys add --detect' to add a mapping.")
		return nil
	}

	fmt.Println("Current mappings:")
	fmt.Println()
	for _, mapping := range mappings {
		if mapping.IsSystemDefault() {
			fmt.Printf("  • System Default → %s\n", mapping.LayoutName)
		} else {
			fmt.Printf("  • %s → %s\n", mapping.DeviceDisplayName, mapping.LayoutName)
		}
	}

	return nil
}
