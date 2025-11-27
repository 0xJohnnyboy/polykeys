//go:build windows

package infrastructure

import (
	"github.com/0xJohnnyboy/polykeys/internal/adapters/devices"
	"github.com/0xJohnnyboy/polykeys/internal/adapters/layouts"
	"github.com/0xJohnnyboy/polykeys/internal/domain"
)

func createPlatformDeviceDetector() (domain.DeviceDetector, error) {
	return devices.NewWindowsDeviceDetector()
}

func createPlatformLayoutSwitcher() (domain.LayoutSwitcher, error) {
	return layouts.NewWindowsLayoutSwitcher(), nil
}
