//go:build windows

package layouts

import (
	"context"
	"fmt"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
)

// WindowsLayoutSwitcher switches keyboard layouts on Windows
type WindowsLayoutSwitcher struct {
	// Map of layout names to Windows layout identifiers (KLID)
	layoutMap map[string]string
}

// NewWindowsLayoutSwitcher creates a new Windows layout switcher
func NewWindowsLayoutSwitcher() *WindowsLayoutSwitcher {
	return &WindowsLayoutSwitcher{
		layoutMap: getDefaultWindowsLayoutMap(),
	}
}

// getDefaultWindowsLayoutMap returns a map of layout names to Windows KLIDs
// KLIDs are 8-digit hex values representing keyboard layouts
func getDefaultWindowsLayoutMap() map[string]string {
	return map[string]string{
		// US layouts
		domain.LayoutUSQwerty:                "00000409", // US
		domain.LayoutUSInternational:         "00020409", // US International
		domain.LayoutUSInternationalDeadKeys: "00020409", // US International

		// French layouts
		domain.LayoutFrenchAzerty: "0000040c", // French

		// UK layouts
		domain.LayoutUKQwerty: "00000809", // UK

		// Alternative layouts
		domain.LayoutColemak: "00000409", // Colemak (requires separate installation)
		domain.LayoutDvorak:  "00010409", // US Dvorak

		// Other languages
		domain.LayoutGerman:     "00000407", // German
		domain.LayoutSpanish:    "0000040a", // Spanish
		domain.LayoutItalian:    "00000410", // Italian
		domain.LayoutPortuguese: "00000816", // Portuguese
		domain.LayoutRussian:    "00000419", // Russian
		domain.LayoutJapanese:   "00000411", // Japanese
	}
}

// SwitchLayout changes the system keyboard layout
func (s *WindowsLayoutSwitcher) SwitchLayout(ctx context.Context, layout *domain.KeyboardLayout) error {
	if layout.OS != domain.OSWindows {
		return fmt.Errorf("layout %s is not for Windows", layout.Name)
	}

	// Get the KLID
	klid := s.getKLID(layout)

	// Windows implementation would use:
	// - LoadKeyboardLayout Win32 API
	// - ActivateKeyboardLayout
	// - Or send WM_INPUTLANGCHANGEREQUEST message

	// Simplified placeholder
	_ = klid
	return fmt.Errorf("Windows layout switching not yet fully implemented (would switch to KLID: %s)", klid)
}

// GetCurrentLayout retrieves the currently active layout
func (s *WindowsLayoutSwitcher) GetCurrentLayout(ctx context.Context) (*domain.KeyboardLayout, error) {
	// Windows implementation would use:
	// - GetKeyboardLayout to get HKL
	// - Extract KLID from HKL
	// - Map back to layout name

	return nil, fmt.Errorf("Windows current layout detection not yet implemented")
}

// GetAvailableLayouts returns all available layouts for Windows
func (s *WindowsLayoutSwitcher) GetAvailableLayouts(ctx context.Context) ([]*domain.KeyboardLayout, error) {
	layouts := make([]*domain.KeyboardLayout, 0, len(s.layoutMap))

	for name, klid := range s.layoutMap {
		layout := domain.NewKeyboardLayout(name, domain.OSWindows, klid)
		layouts = append(layouts, layout)
	}

	return layouts, nil
}

// getKLID returns the Windows KLID for a layout
func (s *WindowsLayoutSwitcher) getKLID(layout *domain.KeyboardLayout) string {
	// First try to use the system identifier directly
	if layout.SystemIdentifier != "" {
		return layout.SystemIdentifier
	}

	// Otherwise, try to map the layout name
	if klid, exists := s.layoutMap[layout.Name]; exists {
		return klid
	}

	// Fallback
	return "00000409" // US layout
}

// AddLayoutMapping adds a custom layout mapping
func (s *WindowsLayoutSwitcher) AddLayoutMapping(name, klid string) {
	s.layoutMap[name] = klid
}
