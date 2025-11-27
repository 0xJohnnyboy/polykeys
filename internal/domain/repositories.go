package domain

import "context"

// DeviceRepository defines the interface for device persistence
type DeviceRepository interface {
	// Save persists a device
	Save(ctx context.Context, device *Device) error
	// FindByID retrieves a device by its ID
	FindByID(ctx context.Context, id string) (*Device, error)
	// FindAll retrieves all devices
	FindAll(ctx context.Context) ([]*Device, error)
	// Delete removes a device by its ID
	Delete(ctx context.Context, id string) error
}

// MappingRepository defines the interface for mapping persistence
type MappingRepository interface {
	// Save persists a mapping
	Save(ctx context.Context, mapping *Mapping) error
	// FindByDeviceID retrieves a mapping for a specific device
	FindByDeviceID(ctx context.Context, deviceID string) (*Mapping, error)
	// FindAll retrieves all mappings
	FindAll(ctx context.Context) ([]*Mapping, error)
	// Delete removes a mapping by device ID
	Delete(ctx context.Context, deviceID string) error
	// GetSystemDefault retrieves the system default mapping
	GetSystemDefault(ctx context.Context) (*Mapping, error)
}

// LayoutRepository defines the interface for layout persistence
type LayoutRepository interface {
	// Save persists a keyboard layout
	Save(ctx context.Context, layout *KeyboardLayout) error
	// FindByName retrieves a layout by name and OS
	FindByName(ctx context.Context, name string, os OperatingSystem) (*KeyboardLayout, error)
	// FindByOS retrieves all layouts for a specific OS
	FindByOS(ctx context.Context, os OperatingSystem) ([]*KeyboardLayout, error)
	// FindAll retrieves all layouts
	FindAll(ctx context.Context) ([]*KeyboardLayout, error)
}
