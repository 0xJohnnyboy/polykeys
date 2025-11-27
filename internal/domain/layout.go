package domain

// OperatingSystem represents the target OS for a keyboard layout
type OperatingSystem string

const (
	OSLinux   OperatingSystem = "linux"
	OSMacOS   OperatingSystem = "macos"
	OSWindows OperatingSystem = "windows"
)

// KeyboardLayout represents a keyboard layout configuration
type KeyboardLayout struct {
	// ID is a unique identifier for the layout
	ID string
	// Name is the human-readable layout name
	Name string
	// OS is the operating system this layout is for
	OS OperatingSystem
	// SystemIdentifier is the OS-specific identifier for the layout
	// (e.g., "us" for Linux, "com.apple.keylayout.US" for macOS)
	SystemIdentifier string
}

// NewKeyboardLayout creates a new KeyboardLayout
func NewKeyboardLayout(name string, os OperatingSystem, systemIdentifier string) *KeyboardLayout {
	return &KeyboardLayout{
		ID:               name + ":" + string(os),
		Name:             name,
		OS:               os,
		SystemIdentifier: systemIdentifier,
	}
}
