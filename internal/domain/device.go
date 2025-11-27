package domain

import "time"

// Device represents a physical keyboard device
type Device struct {
	// ID is a unique identifier for the device (e.g., VID:PID combination)
	ID string
	// VendorID is the USB Vendor ID
	VendorID string
	// ProductID is the USB Product ID
	ProductID string
	// Name is the human-readable device name
	Name string
	// Alias is an optional user-defined alias for easier identification
	Alias string
	// LastSeen is the timestamp when the device was last detected
	LastSeen time.Time
}

// NewDevice creates a new Device with the given parameters
func NewDevice(vendorID, productID, name string) *Device {
	id := vendorID + ":" + productID
	return &Device{
		ID:        id,
		VendorID:  vendorID,
		ProductID: productID,
		Name:      name,
		LastSeen:  time.Now(),
	}
}

// UpdateLastSeen updates the LastSeen timestamp to now
func (d *Device) UpdateLastSeen() {
	d.LastSeen = time.Now()
}

// SetAlias sets a user-defined alias for the device
func (d *Device) SetAlias(alias string) {
	d.Alias = alias
}

// DisplayName returns the alias if set, otherwise the device name
func (d *Device) DisplayName() string {
	if d.Alias != "" {
		return d.Alias
	}
	return d.Name
}
