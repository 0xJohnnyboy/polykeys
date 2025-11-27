package domain

import "context"

// DeviceDetector defines the interface for detecting USB/HID devices
type DeviceDetector interface {
	// StartMonitoring begins monitoring for device connection/disconnection events
	StartMonitoring(ctx context.Context) error
	// StopMonitoring stops the monitoring process
	StopMonitoring() error
	// GetConnectedDevices returns all currently connected keyboard devices
	GetConnectedDevices(ctx context.Context) ([]*Device, error)
	// OnDeviceConnected registers a callback for device connection events
	OnDeviceConnected(callback func(*Device))
	// OnDeviceDisconnected registers a callback for device disconnection events
	OnDeviceDisconnected(callback func(*Device))
}

// LayoutSwitcher defines the interface for switching keyboard layouts
type LayoutSwitcher interface {
	// SwitchLayout changes the system keyboard layout
	SwitchLayout(ctx context.Context, layout *KeyboardLayout) error
}

// ConfigLoader defines the interface for loading configuration
type ConfigLoader interface {
	// Load loads the configuration from the appropriate location
	Load(ctx context.Context) (*Config, error)
	// Save persists the configuration
	Save(ctx context.Context, config *Config) error
	// GetConfigPath returns the path to the configuration file
	GetConfigPath() (string, error)
}

// Config represents the application configuration
type Config struct {
	// Mappings contains all device-to-layout mappings
	Mappings []*Mapping
	// Enabled indicates if polykeys is currently active
	Enabled bool
}
