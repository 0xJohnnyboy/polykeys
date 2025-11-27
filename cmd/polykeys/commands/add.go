package commands

import (
	"fmt"

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
	fmt.Println("Detection mode - please connect a keyboard...")
	fmt.Println("(Not yet implemented)")
	// TODO: Implement device detection
	return nil
}

func runAddManual(device, layout string) error {
	fmt.Printf("Adding mapping: %s -> %s\n", device, layout)
	fmt.Println("(Not yet implemented)")
	// TODO: Implement manual mapping
	return nil
}
