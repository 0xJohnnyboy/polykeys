package commands

import (
	"context"
	"fmt"

	"github.com/0xJohnnyboy/polykeys/internal/infrastructure"
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

	// Initialize app
	app, err := infrastructure.NewApp()
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	ctx := context.Background()

	// Load current config
	if err := app.ManageMappingsUC.LoadFromConfig(ctx); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Remove mapping
	if err := app.ManageMappingsUC.RemoveMapping(ctx, deviceID); err != nil {
		return fmt.Errorf("failed to remove mapping: %w", err)
	}

	// Save config
	if err := app.ManageMappingsUC.SaveToConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Removed mapping for: %s\n", deviceID)
	return nil
}
