//go:build darwin

package layouts

/*
#cgo LDFLAGS: -framework Carbon
#include <Carbon/Carbon.h>

// Helper function to select input source by ID
int selectInputSourceByID(const char* sourceID) {
    CFStringRef sourceIDRef = CFStringCreateWithCString(NULL, sourceID, kCFStringEncodingUTF8);
    if (!sourceIDRef) return -1;

    CFDictionaryRef filter = CFDictionaryCreate(
        NULL,
        (const void **)&kTISPropertyInputSourceID,
        (const void **)&sourceIDRef,
        1,
        &kCFTypeDictionaryKeyCallBacks,
        &kCFTypeDictionaryValueCallBacks
    );

    CFArrayRef sources = TISCreateInputSourceList(filter, false);
    CFRelease(filter);
    CFRelease(sourceIDRef);

    if (!sources || CFArrayGetCount(sources) == 0) {
        if (sources) CFRelease(sources);
        return -2;
    }

    TISInputSourceRef source = (TISInputSourceRef)CFArrayGetValueAtIndex(sources, 0);
    OSStatus status = TISSelectInputSource(source);
    CFRelease(sources);

    return (status == noErr) ? 0 : -3;
}

// Helper to get current input source ID
char* getCurrentInputSourceID() {
    TISInputSourceRef currentSource = TISCopyCurrentKeyboardInputSource();
    if (!currentSource) return NULL;

    CFStringRef sourceID = (CFStringRef)TISGetInputSourceProperty(currentSource, kTISPropertyInputSourceID);
    if (!sourceID) {
        CFRelease(currentSource);
        return NULL;
    }

    CFIndex length = CFStringGetLength(sourceID);
    CFIndex maxSize = CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;
    char *buffer = (char*)malloc(maxSize);

    if (CFStringGetCString(sourceID, buffer, maxSize, kCFStringEncodingUTF8)) {
        CFRelease(currentSource);
        return buffer;
    }

    free(buffer);
    CFRelease(currentSource);
    return NULL;
}
*/
import "C"
import (
	"context"
	"fmt"
	"unsafe"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
)

// DarwinLayoutSwitcher switches keyboard layouts on macOS
type DarwinLayoutSwitcher struct {
	// Map of layout names to macOS input source IDs
	layoutMap map[string]string
}

// NewDarwinLayoutSwitcher creates a new macOS layout switcher
func NewDarwinLayoutSwitcher() *DarwinLayoutSwitcher {
	return &DarwinLayoutSwitcher{
		layoutMap: getDefaultDarwinLayoutMap(),
	}
}

// getDefaultDarwinLayoutMap returns a map of layout names to macOS input source IDs
func getDefaultDarwinLayoutMap() map[string]string {
	return map[string]string{
		// US layouts
		domain.LayoutUSQwerty:                "com.apple.keylayout.US",
		domain.LayoutUSInternational:         "com.apple.keylayout.USInternational-PC",
		domain.LayoutUSInternationalDeadKeys: "com.apple.keylayout.USInternational-PC",

		// French layouts
		domain.LayoutFrenchAzerty: "com.apple.keylayout.French",

		// UK layouts
		domain.LayoutUKQwerty: "com.apple.keylayout.British",

		// Alternative layouts
		domain.LayoutColemak: "com.apple.keylayout.Colemak",
		domain.LayoutDvorak:  "com.apple.keylayout.Dvorak",

		// Other languages
		domain.LayoutGerman:     "com.apple.keylayout.German",
		domain.LayoutSpanish:    "com.apple.keylayout.Spanish",
		domain.LayoutItalian:    "com.apple.keylayout.Italian",
		domain.LayoutPortuguese: "com.apple.keylayout.Portuguese",
		domain.LayoutRussian:    "com.apple.keylayout.Russian",
		domain.LayoutJapanese:   "com.apple.inputmethod.Kotoeri.Japanese",
	}
}

// SwitchLayout changes the system keyboard layout
func (s *DarwinLayoutSwitcher) SwitchLayout(ctx context.Context, layout *domain.KeyboardLayout) error {
	if layout.OS != domain.OSMacOS {
		return fmt.Errorf("layout %s is not for macOS", layout.Name)
	}

	// Get the input source ID
	sourceID := s.getSourceID(layout)

	// Select the input source using Carbon API
	cSourceID := C.CString(sourceID)
	defer C.free(unsafe.Pointer(cSourceID))

	result := C.selectInputSourceByID(cSourceID)
	switch result {
	case 0:
		return nil
	case -1:
		return fmt.Errorf("failed to create source ID string")
	case -2:
		return fmt.Errorf("input source not found: %s", sourceID)
	case -3:
		return fmt.Errorf("failed to select input source")
	default:
		return fmt.Errorf("unknown error selecting input source")
	}
}

// GetCurrentLayout retrieves the currently active layout
func (s *DarwinLayoutSwitcher) GetCurrentLayout(ctx context.Context) (*domain.KeyboardLayout, error) {
	cSourceID := C.getCurrentInputSourceID()
	if cSourceID == nil {
		return nil, fmt.Errorf("failed to get current input source")
	}
	defer C.free(unsafe.Pointer(cSourceID))

	sourceID := C.GoString(cSourceID)

	// Try to find matching layout
	for name, mappedSourceID := range s.layoutMap {
		if mappedSourceID == sourceID {
			return domain.NewKeyboardLayout(name, domain.OSMacOS, sourceID), nil
		}
	}

	// Return generic layout with source ID
	return domain.NewKeyboardLayout(fmt.Sprintf("Layout-%s", sourceID), domain.OSMacOS, sourceID), nil
}

// GetAvailableLayouts returns all available layouts for macOS
func (s *DarwinLayoutSwitcher) GetAvailableLayouts(ctx context.Context) ([]*domain.KeyboardLayout, error) {
	layouts := make([]*domain.KeyboardLayout, 0, len(s.layoutMap))

	for name, sourceID := range s.layoutMap {
		layout := domain.NewKeyboardLayout(name, domain.OSMacOS, sourceID)
		layouts = append(layouts, layout)
	}

	return layouts, nil
}

// getSourceID returns the macOS input source ID for a layout
func (s *DarwinLayoutSwitcher) getSourceID(layout *domain.KeyboardLayout) string {
	// First try to use the system identifier directly
	if layout.SystemIdentifier != "" {
		return layout.SystemIdentifier
	}

	// Otherwise, try to map the layout name
	if sourceID, exists := s.layoutMap[layout.Name]; exists {
		return sourceID
	}

	// Fallback
	return "com.apple.keylayout.US"
}

// AddLayoutMapping adds a custom layout mapping
func (s *DarwinLayoutSwitcher) AddLayoutMapping(name, sourceID string) {
	s.layoutMap[name] = sourceID
}
