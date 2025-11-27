package commands

import (
	"github.com/spf13/cobra"
)

var (
	Debug bool
)

var rootCmd = &cobra.Command{
	Use:   "polykeys",
	Short: "Manage keyboard layouts based on connected devices",
	Long: `Polykeys automatically switches keyboard layouts based on which
keyboard you have connected. Configure device-to-layout mappings
and let polykeys handle the rest.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "Enable debug logging")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(logsCmd)
}
