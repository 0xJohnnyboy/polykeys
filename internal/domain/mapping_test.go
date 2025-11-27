package domain

import "testing"

func TestNewMapping(t *testing.T) {
	deviceID := "046d:c52b"
	displayName := "Logitech Keyboard"
	layoutName := "US International"
	layoutOS := OSLinux

	mapping := NewMapping(deviceID, displayName, layoutName, layoutOS)

	if mapping.DeviceID != deviceID {
		t.Errorf("Expected DeviceID to be '%s', got '%s'", deviceID, mapping.DeviceID)
	}

	if mapping.DeviceDisplayName != displayName {
		t.Errorf("Expected DeviceDisplayName to be '%s', got '%s'", displayName, mapping.DeviceDisplayName)
	}

	if mapping.LayoutName != layoutName {
		t.Errorf("Expected LayoutName to be '%s', got '%s'", layoutName, mapping.LayoutName)
	}

	if mapping.LayoutOS != layoutOS {
		t.Errorf("Expected LayoutOS to be '%s', got '%s'", layoutOS, mapping.LayoutOS)
	}
}

func TestMapping_IsSystemDefault(t *testing.T) {
	tests := []struct {
		name       string
		deviceID   string
		isDefault  bool
	}{
		{
			name:      "System default mapping",
			deviceID:  "system_default",
			isDefault: true,
		},
		{
			name:      "Regular device mapping",
			deviceID:  "046d:c52b",
			isDefault: false,
		},
		{
			name:      "Empty device ID",
			deviceID:  "",
			isDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := NewMapping(tt.deviceID, "Display", "Layout", OSLinux)

			if mapping.IsSystemDefault() != tt.isDefault {
				t.Errorf("Expected IsSystemDefault to be %v, got %v", tt.isDefault, mapping.IsSystemDefault())
			}
		})
	}
}
