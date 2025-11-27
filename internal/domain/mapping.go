package domain

// Mapping represents the association between a device and a keyboard layout
type Mapping struct {
	// DeviceID is the ID of the device (can be "system_default" for the default mapping)
	DeviceID string
	// DeviceDisplayName is the display name of the device
	DeviceDisplayName string
	// LayoutName is the name of the keyboard layout
	LayoutName string
	// LayoutOS is the operating system for this layout
	LayoutOS OperatingSystem
}

// NewMapping creates a new Mapping
func NewMapping(deviceID, deviceDisplayName, layoutName string, layoutOS OperatingSystem) *Mapping {
	return &Mapping{
		DeviceID:          deviceID,
		DeviceDisplayName: deviceDisplayName,
		LayoutName:        layoutName,
		LayoutOS:          layoutOS,
	}
}

// IsSystemDefault returns true if this is the system default mapping
func (m *Mapping) IsSystemDefault() bool {
	return m.DeviceID == "system_default"
}
