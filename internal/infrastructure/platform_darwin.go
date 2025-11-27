//go:build darwin

package infrastructure

import (
	"fmt"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
)

func createPlatformDeviceDetector() (domain.DeviceDetector, error) {
	return nil, fmt.Errorf("macOS device detector not yet implemented")
}

func createPlatformLayoutSwitcher() (domain.LayoutSwitcher, error) {
	return nil, fmt.Errorf("macOS layout switcher not yet implemented")
}
