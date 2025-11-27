package layouts

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
)

// LinuxLayoutSwitcher switches keyboard layouts on Linux using setxkbmap
type LinuxLayoutSwitcher struct {
	// Map of layout names to setxkbmap identifiers
	layoutMap map[string]string
}

// NewLinuxLayoutSwitcher creates a new Linux layout switcher
func NewLinuxLayoutSwitcher() *LinuxLayoutSwitcher {
	return &LinuxLayoutSwitcher{
		layoutMap: getDefaultLayoutMap(),
	}
}

// getDefaultLayoutMap returns a map of layout names to setxkbmap identifiers
func getDefaultLayoutMap() map[string]string {
	return map[string]string{
		// US layouts
		domain.LayoutUSQwerty:                "us",
		domain.LayoutUSInternational:         "us -variant intl",
		domain.LayoutUSInternationalDeadKeys: "us -variant altgr-intl",

		// French layouts
		domain.LayoutFrenchAzerty: "fr",

		// UK layouts
		domain.LayoutUKQwerty: "gb",

		// Alternative layouts
		domain.LayoutColemak: "us -variant colemak",
		domain.LayoutDvorak:  "us -variant dvorak",

		// Other languages
		domain.LayoutGerman:     "de",
		domain.LayoutSpanish:    "es",
		domain.LayoutItalian:    "it",
		domain.LayoutPortuguese: "pt",
		domain.LayoutRussian:    "ru",
		domain.LayoutJapanese:   "jp",
	}
}

// SwitchLayout changes the system keyboard layout
func (s *LinuxLayoutSwitcher) SwitchLayout(ctx context.Context, layout *domain.KeyboardLayout) error {
	if layout.OS != domain.OSLinux {
		return fmt.Errorf("layout %s is not for Linux", layout.Name)
	}

	// Get the setxkbmap identifier
	identifier := s.getSetxkbmapIdentifier(layout)

	// Split identifier into command parts (e.g., "us -variant intl")
	parts := strings.Fields(identifier)
	if len(parts) == 0 {
		return fmt.Errorf("invalid layout identifier: %s", identifier)
	}

	// Build setxkbmap command
	args := append([]string{parts[0]}, parts[1:]...)
	cmd := exec.CommandContext(ctx, "setxkbmap", args...)

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to switch layout: %w (output: %s)", err, string(output))
	}

	return nil
}

// GetCurrentLayout retrieves the currently active layout
func (s *LinuxLayoutSwitcher) GetCurrentLayout(ctx context.Context) (*domain.KeyboardLayout, error) {
	// Use setxkbmap -query to get current layout
	cmd := exec.CommandContext(ctx, "setxkbmap", "-query")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to query current layout: %w", err)
	}

	// Parse output to find layout
	lines := strings.Split(string(output), "\n")
	var layoutCode string
	var variant string

	for _, line := range lines {
		if strings.HasPrefix(line, "layout:") {
			layoutCode = strings.TrimSpace(strings.TrimPrefix(line, "layout:"))
		} else if strings.HasPrefix(line, "variant:") {
			variant = strings.TrimSpace(strings.TrimPrefix(line, "variant:"))
		}
	}

	if layoutCode == "" {
		return nil, fmt.Errorf("could not determine current layout")
	}

	// Build layout name from code and variant
	layoutName := s.getLayoutName(layoutCode, variant)
	systemIdentifier := layoutCode
	if variant != "" {
		systemIdentifier = fmt.Sprintf("%s -variant %s", layoutCode, variant)
	}

	return domain.NewKeyboardLayout(layoutName, domain.OSLinux, systemIdentifier), nil
}

// GetAvailableLayouts returns all available layouts for Linux
func (s *LinuxLayoutSwitcher) GetAvailableLayouts(ctx context.Context) ([]*domain.KeyboardLayout, error) {
	layouts := make([]*domain.KeyboardLayout, 0, len(s.layoutMap))

	for name, identifier := range s.layoutMap {
		layout := domain.NewKeyboardLayout(name, domain.OSLinux, identifier)
		layouts = append(layouts, layout)
	}

	return layouts, nil
}

// getSetxkbmapIdentifier returns the setxkbmap identifier for a layout
func (s *LinuxLayoutSwitcher) getSetxkbmapIdentifier(layout *domain.KeyboardLayout) string {
	// First try to use the system identifier directly
	if layout.SystemIdentifier != "" {
		return layout.SystemIdentifier
	}

	// Otherwise, try to map the layout name
	if identifier, exists := s.layoutMap[layout.Name]; exists {
		return identifier
	}

	// Fallback: assume the layout name is the identifier
	return strings.ToLower(layout.Name)
}

// getLayoutName converts a setxkbmap layout code and variant to a friendly name
func (s *LinuxLayoutSwitcher) getLayoutName(layoutCode, variant string) string {
	// Try to find a matching name in the map
	searchIdentifier := layoutCode
	if variant != "" {
		searchIdentifier = fmt.Sprintf("%s -variant %s", layoutCode, variant)
	}

	for name, identifier := range s.layoutMap {
		if identifier == searchIdentifier {
			return name
		}
	}

	// Fallback: use layout code
	if variant != "" {
		return fmt.Sprintf("%s (%s)", layoutCode, variant)
	}
	return layoutCode
}

// AddLayoutMapping adds a custom layout mapping
func (s *LinuxLayoutSwitcher) AddLayoutMapping(name, setxkbmapIdentifier string) {
	s.layoutMap[name] = setxkbmapIdentifier
}
