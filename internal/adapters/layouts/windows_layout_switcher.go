//go:build windows

package layouts

import (
	"context"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
	"golang.org/x/sys/windows"
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

	// Load the keyboard layout
	hkl, err := s.loadKeyboardLayout(klid)
	if err != nil {
		return fmt.Errorf("failed to load keyboard layout %s: %w", klid, err)
	}

	// Activate the layout for the current thread
	if err := s.activateKeyboardLayout(hkl); err != nil {
		return fmt.Errorf("failed to activate keyboard layout: %w", err)
	}

	// Broadcast to all windows to switch layout
	if err := s.broadcastLayoutChange(hkl); err != nil {
		return fmt.Errorf("failed to broadcast layout change: %w", err)
	}

	return nil
}

// loadKeyboardLayout loads a keyboard layout by KLID
func (s *WindowsLayoutSwitcher) loadKeyboardLayout(klid string) (windows.Handle, error) {
	user32 := windows.NewLazySystemDLL("user32.dll")
	procLoadKeyboardLayout := user32.NewProc("LoadKeyboardLayoutW")

	klidUTF16, err := windows.UTF16PtrFromString(klid)
	if err != nil {
		return 0, fmt.Errorf("failed to convert KLID to UTF16: %w", err)
	}

	// KLF_ACTIVATE = 0x00000001
	const KLF_ACTIVATE = 0x00000001

	ret, _, err := procLoadKeyboardLayout.Call(
		uintptr(unsafe.Pointer(klidUTF16)),
		uintptr(KLF_ACTIVATE),
	)

	if ret == 0 {
		return 0, fmt.Errorf("LoadKeyboardLayout failed: %w", err)
	}

	return windows.Handle(ret), nil
}

// activateKeyboardLayout activates a loaded keyboard layout
func (s *WindowsLayoutSwitcher) activateKeyboardLayout(hkl windows.Handle) error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	procActivateKeyboardLayout := user32.NewProc("ActivateKeyboardLayout")

	// HKL_PREV = 0, HKL_NEXT = 1
	const KLF_SETFORPROCESS = 0x00000100

	ret, _, err := procActivateKeyboardLayout.Call(
		uintptr(hkl),
		uintptr(KLF_SETFORPROCESS),
	)

	if ret == 0 {
		return fmt.Errorf("ActivateKeyboardLayout failed: %w", err)
	}

	return nil
}

// broadcastLayoutChange broadcasts the layout change to all windows
func (s *WindowsLayoutSwitcher) broadcastLayoutChange(hkl windows.Handle) error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	procSendMessage := user32.NewProc("SendMessageW")

	// WM_INPUTLANGCHANGEREQUEST = 0x0050
	const WM_INPUTLANGCHANGEREQUEST = 0x0050
	const HWND_BROADCAST = 0xFFFF

	_, _, _ = procSendMessage.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_INPUTLANGCHANGEREQUEST),
		0,
		uintptr(hkl),
	)

	// SendMessage doesn't fail for broadcasts
	return nil
}

// GetCurrentLayout retrieves the currently active layout
func (s *WindowsLayoutSwitcher) GetCurrentLayout(ctx context.Context) (*domain.KeyboardLayout, error) {
	user32 := windows.NewLazySystemDLL("user32.dll")
	procGetKeyboardLayout := user32.NewProc("GetKeyboardLayout")

	// Get current thread ID (0 = current thread)
	ret, _, _ := procGetKeyboardLayout.Call(0)

	if ret == 0 {
		return nil, fmt.Errorf("GetKeyboardLayout failed")
	}

	hkl := windows.Handle(ret)

	// Extract KLID from HKL (lower 16 bits contain language ID)
	langID := uint16(hkl & 0xFFFF)
	klid := fmt.Sprintf("%08x", langID)

	// Try to find matching layout
	for name, mappedKLID := range s.layoutMap {
		if mappedKLID == klid {
			return domain.NewKeyboardLayout(name, domain.OSWindows, klid), nil
		}
	}

	// Return generic layout with KLID
	return domain.NewKeyboardLayout(fmt.Sprintf("Layout-%s", klid), domain.OSWindows, klid), nil
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
