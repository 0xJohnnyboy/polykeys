package domain

import (
	"testing"
	"time"
)

func TestNewDevice(t *testing.T) {
	vendorID := "046d"
	productID := "c52b"
	name := "Logitech USB Keyboard"

	device := NewDevice(vendorID, productID, name)

	if device.ID != "046d:c52b" {
		t.Errorf("Expected ID to be '046d:c52b', got '%s'", device.ID)
	}

	if device.VendorID != vendorID {
		t.Errorf("Expected VendorID to be '%s', got '%s'", vendorID, device.VendorID)
	}

	if device.ProductID != productID {
		t.Errorf("Expected ProductID to be '%s', got '%s'", productID, device.ProductID)
	}

	if device.Name != name {
		t.Errorf("Expected Name to be '%s', got '%s'", name, device.Name)
	}

	if device.Alias != "" {
		t.Errorf("Expected Alias to be empty, got '%s'", device.Alias)
	}

	// LastSeen should be recent
	if time.Since(device.LastSeen) > time.Second {
		t.Errorf("Expected LastSeen to be recent, got %v", device.LastSeen)
	}
}

func TestDevice_UpdateLastSeen(t *testing.T) {
	device := NewDevice("1234", "5678", "Test Device")

	// Wait a bit to ensure time difference
	time.Sleep(10 * time.Millisecond)
	oldTime := device.LastSeen

	device.UpdateLastSeen()

	if !device.LastSeen.After(oldTime) {
		t.Errorf("Expected LastSeen to be updated, old: %v, new: %v", oldTime, device.LastSeen)
	}
}

func TestDevice_SetAlias(t *testing.T) {
	device := NewDevice("1234", "5678", "Test Device")
	alias := "MyKeyboard"

	device.SetAlias(alias)

	if device.Alias != alias {
		t.Errorf("Expected Alias to be '%s', got '%s'", alias, device.Alias)
	}
}

func TestDevice_DisplayName(t *testing.T) {
	tests := []struct {
		name           string
		deviceName     string
		alias          string
		expectedDisplay string
	}{
		{
			name:           "No alias - returns device name",
			deviceName:     "Logitech Keyboard",
			alias:          "",
			expectedDisplay: "Logitech Keyboard",
		},
		{
			name:           "With alias - returns alias",
			deviceName:     "Logitech Keyboard",
			alias:          "Office Keyboard",
			expectedDisplay: "Office Keyboard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := NewDevice("1234", "5678", tt.deviceName)
			if tt.alias != "" {
				device.SetAlias(tt.alias)
			}

			displayName := device.DisplayName()
			if displayName != tt.expectedDisplay {
				t.Errorf("Expected DisplayName to be '%s', got '%s'", tt.expectedDisplay, displayName)
			}
		})
	}
}
