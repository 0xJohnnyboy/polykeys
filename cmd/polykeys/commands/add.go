package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
	"github.com/0xJohnnyboy/polykeys/internal/infrastructure"
	"github.com/0xJohnnyboy/polykeys/internal/logger"
	"github.com/spf13/cobra"
)

var (
	detectFlag bool
	deviceFlag string
	layoutFlag string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new device-to-layout mapping",
	Long: `Add a new mapping between a keyboard device and a layout.

Use --detect to interactively detect a device when you plug it in.
Use --device and --layout to manually specify a mapping.`,
	RunE: runAdd,
}

func init() {
	addCmd.Flags().BoolVar(&detectFlag, "detect", false, "Detect next connected device")
	addCmd.Flags().StringVar(&deviceFlag, "device", "", "Device name or ID")
	addCmd.Flags().StringVar(&layoutFlag, "layout", "", "Layout name")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if detectFlag {
		return runAddDetect()
	}

	if deviceFlag == "" || layoutFlag == "" {
		return fmt.Errorf("either use --detect or provide both --device and --layout")
	}

	return runAddManual(deviceFlag, layoutFlag)
}

func runAddDetect() error {
	// Set debug mode
	logger.SetDebug(Debug)

	// Initialize app
	app, err := infrastructure.NewApp()
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	ctx := context.Background()

	// Load current config
	_ = app.ManageMappingsUC.LoadFromConfig(ctx)

	fmt.Println("🔍 Detection mode")
	fmt.Println("Please connect a keyboard now...")
	fmt.Println()

	// Get currently connected devices
	currentDevices, err := app.MonitorDevicesUC.GetConnectedDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connected devices: %w", err)
	}

	// Start monitoring
	if err := app.DeviceDetector.StartMonitoring(ctx); err != nil {
		return fmt.Errorf("failed to start monitoring: %w", err)
	}
	defer app.DeviceDetector.StopMonitoring()

	// Wait for a new device
	deviceChan := make(chan *domain.Device, 1)
	app.DeviceDetector.OnDeviceConnected(func(device *domain.Device) {
		// Check if this is a new device
		isNew := true
		for _, existing := range currentDevices {
			if existing.ID == device.ID {
				isNew = false
				break
			}
		}

		if isNew {
			select {
			case deviceChan <- device:
			default:
			}
		}
	})

	// Wait for device
	device := <-deviceChan

	fmt.Printf("✓ Detected: %s (%s)\n", device.Name, device.ID)
	fmt.Println()

	// Ask for optional alias
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter an alias (optional, press Enter to skip): ")
	alias, _ := reader.ReadString('\n')
	alias = strings.TrimSpace(alias)

	if alias != "" {
		device.SetAlias(alias)
	}

	// Get available layouts
	layouts, err := app.LayoutSwitcher.GetAvailableLayouts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get layouts: %w", err)
	}

	// Display available layouts
	fmt.Println()
	fmt.Println("Available layouts:")
	for i, layout := range layouts {
		fmt.Printf("  %d. %s\n", i+1, layout.Name)
	}
	fmt.Println()

	// Ask user to choose
	fmt.Print("Choose a layout (number): ")
	var choice int
	_, err = fmt.Scanf("%d", &choice)
	if err != nil || choice < 1 || choice > len(layouts) {
		return fmt.Errorf("invalid choice")
	}

	selectedLayout := layouts[choice-1]

	// Get current OS
	var os domain.OperatingSystem
	switch runtime.GOOS {
	case "linux":
		os = domain.OSLinux
	case "darwin":
		os = domain.OSMacOS
	case "windows":
		os = domain.OSWindows
	default:
		os = domain.OSLinux
	}

	// Add mapping
	if err := app.ManageMappingsUC.AddMapping(ctx, device, selectedLayout.Name, os); err != nil {
		return fmt.Errorf("failed to add mapping: %w", err)
	}

	// Save to config
	if err := app.ManageMappingsUC.SaveToConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println()
	fmt.Printf("✓ Mapping created: %s → %s\n", device.DisplayName(), selectedLayout.Name)

	return nil
}

func runAddManual(deviceID, layoutName string) error {
	// Initialize app
	app, err := infrastructure.NewApp()
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	ctx := context.Background()

	// Load current config
	_ = app.ManageMappingsUC.LoadFromConfig(ctx)

	// Get current OS
	var os domain.OperatingSystem
	switch runtime.GOOS {
	case "linux":
		os = domain.OSLinux
	case "darwin":
		os = domain.OSMacOS
	case "windows":
		os = domain.OSWindows
	default:
		os = domain.OSLinux
	}

	// Create a device object
	device := domain.NewDevice("unknown", "unknown", deviceID)
	device.ID = deviceID

	// Add mapping
	if err := app.ManageMappingsUC.AddMapping(ctx, device, layoutName, os); err != nil {
		return fmt.Errorf("failed to add mapping: %w", err)
	}

	// Save to config
	if err := app.ManageMappingsUC.SaveToConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✓ Mapping created: %s → %s\n", deviceID, layoutName)

	return nil
}
