//go:build windows

package layouts

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
	"github.com/0xJohnnyboy/polykeys/internal/errors"
	"golang.org/x/sys/windows"
)

// WindowsLayoutSwitcher switches keyboard layouts on Windows
type WindowsLayoutSwitcher struct{}

// NewWindowsLayoutSwitcher creates a new Windows layout switcher
func NewWindowsLayoutSwitcher() *WindowsLayoutSwitcher {
	return &WindowsLayoutSwitcher{}
}

// SwitchLayout changes the system keyboard layout while preserving the system language
func (s *WindowsLayoutSwitcher) SwitchLayout(ctx context.Context, layout *domain.KeyboardLayout) error {
	if layout.OS != domain.OSWindows {
		return errors.WithDetails(
			errors.New(errors.ErrCodeLayoutInvalidOS, "layout is not for Windows"),
			map[string]any{
				"layout": layout.Name,
				"os":     layout.OS,
			},
		)
	}

	// Get the KLID
	klid := s.getKLID(layout)

	// Load the keyboard layout with current language preserved
	hkl, err := s.loadKeyboardLayoutPreservingLanguage(klid)
	if err != nil {
		if pkErr, ok := err.(*errors.PolykeysError); ok {
			return errors.WithDetails(pkErr, map[string]any{
				"layout": layout.Name,
				"klid":   klid,
			})
		}
		return errors.Wrap(errors.ErrCodeLayoutNotFound, "failed to load keyboard layout", err)
	}

	// Activate the layout for the current thread
	if err := s.activateKeyboardLayout(hkl); err != nil {
		return errors.WithDetails(
			errors.Wrap(errors.ErrCodeLayoutSelectFailed, "failed to activate keyboard layout", err),
			map[string]any{"layout": layout.Name},
		)
	}

	// Broadcast to all windows to switch layout
	if err := s.broadcastLayoutChange(hkl); err != nil {
		return errors.Wrap(errors.ErrCodeLayoutSelectFailed, "failed to broadcast layout change", err)
	}

	return nil
}

// loadKeyboardLayout loads a keyboard layout by KLID
func (s *WindowsLayoutSwitcher) loadKeyboardLayout(klid string) (windows.Handle, error) {
	user32 := windows.NewLazySystemDLL("user32.dll")
	procLoadKeyboardLayout := user32.NewProc("LoadKeyboardLayoutW")

	klidUTF16, err := windows.UTF16PtrFromString(klid)
	if err != nil {
		return 0, errors.Wrap(errors.ErrCodeLayoutStringFailed, "failed to convert KLID to UTF16", err)
	}

	// KLF_ACTIVATE = 0x00000001
	const KLF_ACTIVATE = 0x00000001

	ret, _, err := procLoadKeyboardLayout.Call(
		uintptr(unsafe.Pointer(klidUTF16)),
		uintptr(KLF_ACTIVATE),
	)

	if ret == 0 {
		return 0, errors.Wrap(errors.ErrCodeLayoutNotFound, "LoadKeyboardLayout failed", err)
	}

	return windows.Handle(ret), nil
}

// loadKeyboardLayoutPreservingLanguage loads a keyboard layout while preserving the current input language
func (s *WindowsLayoutSwitcher) loadKeyboardLayoutPreservingLanguage(klid string) (windows.Handle, error) {
	user32 := windows.NewLazySystemDLL("user32.dll")
	procGetKeyboardLayout := user32.NewProc("GetKeyboardLayout")
	procLoadKeyboardLayout := user32.NewProc("LoadKeyboardLayoutW")

	// Get current HKL to extract current language
	currentHKL, _, _ := procGetKeyboardLayout.Call(0)

	// Extract current language ID (upper 16 bits of HKL)
	currentLangID := uint16((currentHKL >> 16) & 0xFFFF)

	// Parse the layout ID from KLID
	// KLID format: "00020409" where the last 4 chars (0409) are the layout ID
	if len(klid) < 4 {
		return 0, errors.New(errors.ErrCodeLayoutInvalidIdentifier, "KLID too short")
	}

	// Extract layout ID from KLID (last 4 hex chars)
	layoutIDStr := klid[len(klid)-4:]

	// First try: construct HKL with current language + new layout and try to activate directly
	var layoutID uint64
	_, err := fmt.Sscanf(layoutIDStr, "%04x", &layoutID)
	if err != nil {
		return 0, errors.Wrap(errors.ErrCodeLayoutInvalidIdentifier, "failed to parse layout ID from KLID", err)
	}

	// Construct HKL: (currentLangID << 16) | layoutID
	preservedHKL := (uintptr(currentLangID) << 16) | uintptr(layoutID)

	// Try to activate this HKL directly (in case it's already loaded)
	procActivateKeyboardLayout := user32.NewProc("ActivateKeyboardLayout")
	ret, _, _ := procActivateKeyboardLayout.Call(preservedHKL, uintptr(0x00000001)) // KLF_ACTIVATE

	if ret != 0 {
		// Success! The layout was already loaded with our current language
		return windows.Handle(preservedHKL), nil
	}

	// If direct activation failed, load the layout with original KLID
	// This ensures the layout is installed, even if with a different language
	klidUTF16, err := windows.UTF16PtrFromString(klid)
	if err != nil {
		return 0, errors.Wrap(errors.ErrCodeLayoutStringFailed, "failed to convert KLID to UTF16", err)
	}

	const KLF_ACTIVATE = 0x00000001
	ret, _, err = procLoadKeyboardLayout.Call(
		uintptr(unsafe.Pointer(klidUTF16)),
		uintptr(KLF_ACTIVATE),
	)

	if ret == 0 {
		return 0, errors.Wrap(errors.ErrCodeLayoutNotFound, "LoadKeyboardLayout failed", err)
	}

	// Now that the layout is loaded, try again to activate it with our current language
	ret2, _, _ := procActivateKeyboardLayout.Call(preservedHKL, uintptr(KLF_ACTIVATE))
	if ret2 != 0 {
		return windows.Handle(preservedHKL), nil
	}

	// Fallback: return the HKL from LoadKeyboardLayout (may change language)
	return windows.Handle(ret), nil
}

// activateKeyboardLayout activates a loaded keyboard layout
func (s *WindowsLayoutSwitcher) activateKeyboardLayout(hkl windows.Handle) error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	procActivateKeyboardLayout := user32.NewProc("ActivateKeyboardLayout")

	// Flags for ActivateKeyboardLayout
	const (
		KLF_ACTIVATE      = 0x00000001 // Activate for current thread
		KLF_SETFORPROCESS = 0x00000100 // Set for entire process
		KLF_REORDER       = 0x00000008 // Reorder layout list
	)

	// Activate with multiple flags to ensure it takes effect
	flags := KLF_ACTIVATE | KLF_SETFORPROCESS | KLF_REORDER

	ret, _, err := procActivateKeyboardLayout.Call(
		uintptr(hkl),
		uintptr(flags),
	)

	if ret == 0 {
		return errors.Wrap(errors.ErrCodeLayoutSelectFailed, "ActivateKeyboardLayout failed", err)
	}

	return nil
}

// broadcastLayoutChange broadcasts the layout change to all windows
func (s *WindowsLayoutSwitcher) broadcastLayoutChange(hkl windows.Handle) error {
	user32 := windows.NewLazySystemDLL("user32.dll")
	procPostMessage := user32.NewProc("PostMessageW")

	// Windows messages
	const (
		WM_INPUTLANGCHANGEREQUEST = 0x0050
		WM_INPUTLANGCHANGE        = 0x0051
		HWND_BROADCAST            = 0xFFFF
	)

	// Use PostMessage for async broadcast (recommended for broadcasts)
	// IMPORTANT: Never use SendMessage with HWND_BROADCAST as it's blocking
	// and can freeze the application if any window doesn't respond
	procPostMessage.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_INPUTLANGCHANGEREQUEST),
		0,
		uintptr(hkl),
	)

	// Also post WM_INPUTLANGCHANGE (use PostMessage, NOT SendMessage)
	procPostMessage.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_INPUTLANGCHANGE),
		0,
		uintptr(hkl),
	)

	return nil
}

// getKLID returns the Windows KLID for a layout
func (s *WindowsLayoutSwitcher) getKLID(layout *domain.KeyboardLayout) string {
	// Use the system identifier from the layout (set by the repository)
	if layout.SystemIdentifier != "" {
		return layout.SystemIdentifier
	}

	// Fallback to US layout if no identifier is set
	return "00000409"
}
