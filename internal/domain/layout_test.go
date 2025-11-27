package domain

import "testing"

func TestNewKeyboardLayout(t *testing.T) {
	name := "US International"
	os := OSLinux
	systemID := "us"

	layout := NewKeyboardLayout(name, os, systemID)

	expectedID := "US International:linux"
	if layout.ID != expectedID {
		t.Errorf("Expected ID to be '%s', got '%s'", expectedID, layout.ID)
	}

	if layout.Name != name {
		t.Errorf("Expected Name to be '%s', got '%s'", name, layout.Name)
	}

	if layout.OS != os {
		t.Errorf("Expected OS to be '%s', got '%s'", os, layout.OS)
	}

	if layout.SystemIdentifier != systemID {
		t.Errorf("Expected SystemIdentifier to be '%s', got '%s'", systemID, layout.SystemIdentifier)
	}
}

func TestOperatingSystem_Constants(t *testing.T) {
	tests := []struct {
		name     string
		os       OperatingSystem
		expected string
	}{
		{"Linux OS", OSLinux, "linux"},
		{"macOS OS", OSMacOS, "macos"},
		{"Windows OS", OSWindows, "windows"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.os) != tt.expected {
				t.Errorf("Expected OS constant to be '%s', got '%s'", tt.expected, string(tt.os))
			}
		})
	}
}

func TestKeyboardLayout_DifferentOS(t *testing.T) {
	tests := []struct {
		name     string
		os       OperatingSystem
		systemID string
	}{
		{
			name:     "Linux layout",
			os:       OSLinux,
			systemID: "us",
		},
		{
			name:     "macOS layout",
			os:       OSMacOS,
			systemID: "com.apple.keylayout.US",
		},
		{
			name:     "Windows layout",
			os:       OSWindows,
			systemID: "00000409",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := NewKeyboardLayout("US", tt.os, tt.systemID)

			if layout.OS != tt.os {
				t.Errorf("Expected OS to be '%s', got '%s'", tt.os, layout.OS)
			}

			if layout.SystemIdentifier != tt.systemID {
				t.Errorf("Expected SystemIdentifier to be '%s', got '%s'", tt.systemID, layout.SystemIdentifier)
			}
		})
	}
}
