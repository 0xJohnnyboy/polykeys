//go:build linux

package infrastructure

import (
	"github.com/0xJohnnyboy/polykeys/internal/adapters/devices"
	"github.com/0xJohnnyboy/polykeys/internal/adapters/layouts"
	"github.com/0xJohnnyboy/polykeys/internal/domain"
)

func createPlatformDeviceDetector() (domain.DeviceDetector, error) {
	return devices.NewLinuxDeviceDetector()
}

func createPlatformLayoutSwitcher() (domain.LayoutSwitcher, error) {
	return layouts.NewLinuxLayoutSwitcher(), nil
}
