package layouts

import (
	"context"
	"os/exec"
	"strings"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
	"github.com/0xJohnnyboy/polykeys/internal/errors"
)

// LinuxLayoutSwitcher switches keyboard layouts on Linux using setxkbmap
type LinuxLayoutSwitcher struct{}

// NewLinuxLayoutSwitcher creates a new Linux layout switcher
func NewLinuxLayoutSwitcher() *LinuxLayoutSwitcher {
	return &LinuxLayoutSwitcher{}
}

// SwitchLayout changes the system keyboard layout
func (s *LinuxLayoutSwitcher) SwitchLayout(ctx context.Context, layout *domain.KeyboardLayout) error {
	if layout.OS != domain.OSLinux {
		return errors.WithDetails(
			errors.New(errors.ErrCodeLayoutInvalidOS, "layout is not for Linux"),
			map[string]interface{}{
				"layout": layout.Name,
				"os":     layout.OS,
			},
		)
	}

	// Get the setxkbmap identifier
	identifier := s.getSetxkbmapIdentifier(layout)

	// Split identifier into command parts (e.g., "us -variant intl")
	parts := strings.Fields(identifier)
	if len(parts) == 0 {
		return errors.WithDetails(
			errors.New(errors.ErrCodeLayoutInvalidIdentifier, "invalid layout identifier"),
			map[string]interface{}{
				"layout":     layout.Name,
				"identifier": identifier,
			},
		)
	}

	// Build setxkbmap command
	args := append([]string{parts[0]}, parts[1:]...)
	cmd := exec.CommandContext(ctx, "setxkbmap", args...)

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.WithDetails(
			errors.Wrap(errors.ErrCodeLayoutSelectFailed, "failed to switch layout", err),
			map[string]interface{}{
				"layout": layout.Name,
				"output": string(output),
			},
		)
	}

	return nil
}

// getSetxkbmapIdentifier returns the setxkbmap identifier for a layout
func (s *LinuxLayoutSwitcher) getSetxkbmapIdentifier(layout *domain.KeyboardLayout) string {
	// Use the system identifier from the layout (set by the repository)
	if layout.SystemIdentifier != "" {
		return layout.SystemIdentifier
	}

	// Fallback to "us" if no identifier is set
	return "us"
}
