package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
	"github.com/0xJohnnyboy/polykeys/internal/infrastructure"
	"github.com/spf13/cobra"
)

var followFlag bool

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show device events logs",
	Long:  `Display USB/HID device events for debugging purposes.`,
	RunE:  runLogs,
}

func init() {
	logsCmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Follow log output")
}

func runLogs(cmd *cobra.Command, args []string) error {
	// Initialize app
	app, err := infrastructure.NewApp()
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if !followFlag {
		// Just show currently connected devices
		devices, err := app.MonitorDevicesUC.GetConnectedDevices(ctx)
		if err != nil {
			return fmt.Errorf("failed to get connected devices: %w", err)
		}

		fmt.Println("Currently connected devices:")
		if len(devices) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, device := range devices {
				fmt.Printf("  • %s (%s)\n", device.Name, device.ID)
			}
		}
		return nil
	}

	// Follow mode
	fmt.Println("📡 Monitoring device events... (Ctrl+C to stop)")
	fmt.Println()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nStopping...")
		cancel()
	}()

	// Register callbacks
	app.DeviceDetector.OnDeviceConnected(func(device *domain.Device) {
		fmt.Printf("[CONNECTED] %s (%s)\n", device.Name, device.ID)
	})

	app.DeviceDetector.OnDeviceDisconnected(func(device *domain.Device) {
		fmt.Printf("[DISCONNECTED] %s (%s)\n", device.Name, device.ID)
	})

	// Start monitoring
	if err := app.DeviceDetector.StartMonitoring(ctx); err != nil {
		return fmt.Errorf("failed to start monitoring: %w", err)
	}
	defer app.DeviceDetector.StopMonitoring()

	// Wait for context cancellation
	<-ctx.Done()

	return nil
}
