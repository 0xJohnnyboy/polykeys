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

    // Check if source is enabled, if not enable it first (like Windows does)
    CFBooleanRef isEnabled = (CFBooleanRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceIsEnabled);
    if (!isEnabled || !CFBooleanGetValue(isEnabled)) {
        OSStatus enableStatus = TISEnableInputSource(source);
        if (enableStatus != noErr) {
            CFRelease(sources);
            return -4; // Failed to enable
        }
    }

    OSStatus status = TISSelectInputSource(source);
    CFRelease(sources);

    return (status == noErr) ? 0 : -3;
}

// Helper function to find and select input source by name (tries multiple properties)
int selectInputSourceByName(const char* name) {
    CFStringRef nameRef = CFStringCreateWithCString(NULL, name, kCFStringEncodingUTF8);
    if (!nameRef) return -1;

    // Also try with com.apple.keylayout. prefix for keyboard layouts
    char idWithPrefix[256];
    snprintf(idWithPrefix, sizeof(idWithPrefix), "com.apple.keylayout.%s", name);
    CFStringRef idWithPrefixRef = CFStringCreateWithCString(NULL, idWithPrefix, kCFStringEncodingUTF8);

    // Get all input sources
    CFArrayRef sources = TISCreateInputSourceList(NULL, false);
    if (!sources) {
        CFRelease(nameRef);
        if (idWithPrefixRef) CFRelease(idWithPrefixRef);
        return -2;
    }

    TISInputSourceRef foundSource = NULL;
    CFIndex count = CFArrayGetCount(sources);

    for (CFIndex i = 0; i < count; i++) {
        TISInputSourceRef source = (TISInputSourceRef)CFArrayGetValueAtIndex(sources, i);

        // Try localized name
        CFStringRef localizedName = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyLocalizedName);
        if (localizedName && CFStringCompare(localizedName, nameRef, kCFCompareCaseInsensitive) == kCFCompareEqualTo) {
            foundSource = source;
            break;
        }

        // Try input source ID (full ID or just the name part)
        CFStringRef sourceID = (CFStringRef)TISGetInputSourceProperty(source, kTISPropertyInputSourceID);
        if (sourceID) {
            // Try exact match with com.apple.keylayout.NAME
            if (idWithPrefixRef && CFStringCompare(sourceID, idWithPrefixRef, kCFCompareCaseInsensitive) == kCFCompareEqualTo) {
                foundSource = source;
                break;
            }
            // Try if sourceID ends with the name (for layouts like com.apple.keylayout.ABC-AZERTY)
            if (CFStringHasSuffix(sourceID, nameRef)) {
                foundSource = source;
                break;
            }
        }
    }

    CFRelease(nameRef);
    if (idWithPrefixRef) CFRelease(idWithPrefixRef);

    if (!foundSource) {
        CFRelease(sources);
        return -2;
    }

    // Check if source is enabled, if not enable it first (like Windows does)
    CFBooleanRef isEnabled = (CFBooleanRef)TISGetInputSourceProperty(foundSource, kTISPropertyInputSourceIsEnabled);
    if (!isEnabled || !CFBooleanGetValue(isEnabled)) {
        OSStatus enableStatus = TISEnableInputSource(foundSource);
        if (enableStatus != noErr) {
            CFRelease(sources);
            return -4; // Failed to enable
        }
    }

    OSStatus status = TISSelectInputSource(foundSource);
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
	"github.com/0xJohnnyboy/polykeys/internal/errors"
	"github.com/0xJohnnyboy/polykeys/internal/logger"
)

// darwinErrorInfo maps C result codes to error information
type darwinErrorInfo struct {
	code    errors.ErrorCode
	message string
}

var darwinErrorMap = map[int]darwinErrorInfo{
	-1: {errors.ErrCodeLayoutStringFailed, "failed to create C string"},
	-2: {errors.ErrCodeLayoutNotFound, "input source not found"},
	-3: {errors.ErrCodeLayoutSelectFailed, "failed to select input source"},
	-4: {errors.ErrCodeLayoutEnableFailed, "failed to enable input source"},
}

// DarwinLayoutSwitcher switches keyboard layouts on macOS
type DarwinLayoutSwitcher struct{}

// NewDarwinLayoutSwitcher creates a new macOS layout switcher
func NewDarwinLayoutSwitcher() *DarwinLayoutSwitcher {
	return &DarwinLayoutSwitcher{}
}

// selectInputSourceByID selects an input source by its ID
func (s *DarwinLayoutSwitcher) selectInputSourceByID(sourceID string) error {
	cSourceID := C.CString(sourceID)
	defer C.free(unsafe.Pointer(cSourceID))

	result := C.selectInputSourceByID(cSourceID)

	if result == 0 {
		return nil
	}

	if errInfo, exists := darwinErrorMap[int(result)]; exists {
		return errors.New(errInfo.code, errInfo.message)
	}

	return errors.New(errors.ErrCodeUnknown, "unknown error selecting input source by ID")
}

// selectInputSourceByName selects an input source by its name
func (s *DarwinLayoutSwitcher) selectInputSourceByName(name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	result := C.selectInputSourceByName(cName)

	if result == 0 {
		return nil
	}

	if errInfo, exists := darwinErrorMap[int(result)]; exists {
		return errors.New(errInfo.code, errInfo.message)
	}

	return errors.New(errors.ErrCodeUnknown, "unknown error selecting input source by name")
}

// SwitchLayout changes the system keyboard layout
func (s *DarwinLayoutSwitcher) SwitchLayout(ctx context.Context, layout *domain.KeyboardLayout) error {
	if layout.OS != domain.OSMacOS {
		return errors.WithDetails(
			errors.New(errors.ErrCodeLayoutInvalidOS, "layout is not for macOS"),
			map[string]interface{}{
				"layout": layout.Name,
				"os":     layout.OS,
			},
		)
	}

	// Get the input source ID
	sourceID := s.getSourceID(layout)

	// Try to select by ID first
	err := s.selectInputSourceByID(sourceID)
	if err == nil {
		return nil
	}

	// If selection by ID failed, try by localized name as fallback
	logger.Debug("[Switcher] Layout ID %s not found, trying by name: %s\n", sourceID, layout.Name)

	err = s.selectInputSourceByName(layout.Name)
	if err == nil {
		logger.Debug("[Switcher] Successfully switched to %s by name\n", layout.Name)
		return nil
	}

	// Add details to the error
	return errors.WithDetails(err.(*errors.PolykeysError), map[string]interface{}{
		"layout":   layout.Name,
		"sourceID": sourceID,
	})
}

// getSourceID returns the macOS input source ID for a layout
func (s *DarwinLayoutSwitcher) getSourceID(layout *domain.KeyboardLayout) string {
	// Use the system identifier from the layout (set by the repository)
	if layout.SystemIdentifier != "" {
		return layout.SystemIdentifier
	}

	// Fallback to US if no identifier is set
	return "com.apple.keylayout.US"
}
