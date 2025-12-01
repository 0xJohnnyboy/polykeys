//go:build windows

package layouts

import (
	"context"
	"strings"
	"unsafe"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
	"github.com/0xJohnnyboy/polykeys/internal/errors"
	"github.com/0xJohnnyboy/polykeys/internal/logger"
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

// loadKeyboardLayoutPreservingLanguage loads a keyboard layout while preserving the current input language
func (s *WindowsLayoutSwitcher) loadKeyboardLayoutPreservingLanguage(klid string) (windows.Handle, error) {
	user32 := windows.NewLazySystemDLL("user32.dll")
	procGetKeyboardLayout := user32.NewProc("GetKeyboardLayout")
	procLoadKeyboardLayout := user32.NewProc("LoadKeyboardLayoutW")
	procActivateKeyboardLayout := user32.NewProc("ActivateKeyboardLayout")
	procGetKeyboardLayoutList := user32.NewProc("GetKeyboardLayoutList")
	procGetKeyboardLayoutName := user32.NewProc("GetKeyboardLayoutNameW")

	const (
		KLF_ACTIVATE  = 0x00000001 // Activate keyboard layout
		KL_NAMELENGTH = 9          // Keyboard layout name length (8 chars + null terminator)
	)

	// Get current HKL to extract current language
	currentHKL, _, _ := procGetKeyboardLayout.Call(0)

	// Extract current language ID (lower 16 bits of HKL)
	currentLangID := uint16(currentHKL & 0xFFFF)

	// Load the target layout to get its device ID
	klidUTF16, err := windows.UTF16PtrFromString(klid)
	if err != nil {
		return 0, errors.Wrap(errors.ErrCodeLayoutStringFailed, "failed to convert KLID to UTF16", err)
	}
	loadedHKL, _, err := procLoadKeyboardLayout.Call(
		uintptr(unsafe.Pointer(klidUTF16)),
		uintptr(KLF_ACTIVATE),
	)

	if loadedHKL == 0 {
		return 0, errors.Wrap(errors.ErrCodeLayoutNotFound, "LoadKeyboardLayout failed", err)
	}

	// Extract the device identifier from the loaded HKL (upper 16 bits)
	targetDeviceID := uint16((loadedHKL >> 16) & 0xFFFF)

	if logger.IsDebug() {
		logger.Debug("[Switcher] Current HKL: 0x%08x, Current LangID: 0x%04x, Target KLID: %s\n",
			currentHKL, currentLangID, klid)
		logger.Debug("[Switcher] Loaded HKL: 0x%08x, Target DeviceID: 0x%04x\n",
			loadedHKL, targetDeviceID)
	}

	// Get list of all installed keyboard layouts
	numLayouts, _, _ := procGetKeyboardLayoutList.Call(0, 0)
	if numLayouts == 0 {
		if logger.IsDebug() {
			logger.Debug("[Switcher] No layouts found in list, using loaded HKL\n")
		}
		return windows.Handle(loadedHKL), nil
	}

	// Allocate buffer for HKL list
	hklList := make([]uintptr, numLayouts)
	procGetKeyboardLayoutList.Call(numLayouts, uintptr(unsafe.Pointer(&hklList[0])))

	// Search for a layout with matching device ID and current language
	for _, hkl := range hklList {
		deviceID := uint16((hkl >> 16) & 0xFFFF)
		langID := uint16(hkl & 0xFFFF)

		if deviceID == targetDeviceID && langID == currentLangID {
			// Found a candidate with matching DeviceID and LangID
			// Now verify the actual KLID matches what we want

			// Activate this HKL temporarily to get its name
			ret, _, _ := procActivateKeyboardLayout.Call(hkl, uintptr(KLF_ACTIVATE))
			if ret == 0 {
				continue // Activation failed, try next candidate
			}

			// Get the actual KLID of this layout
			var actualKLID [KL_NAMELENGTH]uint16
			procGetKeyboardLayoutName.Call(uintptr(unsafe.Pointer(&actualKLID[0])))
			actualKLIDStr := windows.UTF16ToString(actualKLID[:])

			if logger.IsDebug() {
				logger.Debug("[Switcher] Checking candidate: HKL 0x%08x (DeviceID: 0x%04x, LangID: 0x%04x), actual KLID: %s\n",
					hkl, deviceID, langID, actualKLIDStr)
			}

			// Check if the actual KLID matches our target
			if strings.EqualFold(actualKLIDStr, klid) {
				// Perfect match: same layout variant, current language!
				if logger.IsDebug() {
					logger.Debug("[Switcher] ✓ Found exact matching layout with preserved language\n")
				}
				return windows.Handle(hkl), nil
			}
		}
	}

	// Fallback: no exact match found with current language and KLID
	// This might happen if the layout with current language isn't installed
	if logger.IsDebug() {
		logger.Debug("[Switcher] ⚠ No exact matching layout variant with current language found, using loaded HKL\n")
	}
	return windows.Handle(loadedHKL), nil
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
