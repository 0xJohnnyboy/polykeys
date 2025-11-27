package usecases

import (
	"context"
	"fmt"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
)

// SwitchLayoutUseCase handles the logic for switching keyboard layouts
type SwitchLayoutUseCase struct {
	mappingRepo    domain.MappingRepository
	layoutRepo     domain.LayoutRepository
	layoutSwitcher domain.LayoutSwitcher
}

// NewSwitchLayoutUseCase creates a new SwitchLayoutUseCase
func NewSwitchLayoutUseCase(
	mappingRepo domain.MappingRepository,
	layoutRepo domain.LayoutRepository,
	layoutSwitcher domain.LayoutSwitcher,
) *SwitchLayoutUseCase {
	return &SwitchLayoutUseCase{
		mappingRepo:    mappingRepo,
		layoutRepo:     layoutRepo,
		layoutSwitcher: layoutSwitcher,
	}
}

// SwitchForDevice switches the keyboard layout based on the connected device
func (uc *SwitchLayoutUseCase) SwitchForDevice(ctx context.Context, device *domain.Device) error {
	// Find mapping for this device
	mapping, err := uc.mappingRepo.FindByDeviceID(ctx, device.ID)
	if err != nil {
		// If no mapping found for this device, try system default
		fmt.Printf("[Switch] ⚠ No mapping found for device %s (%s), using system default\n",
			device.DisplayName(), device.ID)
		mapping, err = uc.mappingRepo.GetSystemDefault(ctx)
		if err != nil {
			return fmt.Errorf("no mapping found for device %s and no system default: %w", device.DisplayName(), err)
		}
	} else {
		fmt.Printf("[Switch] ✓ Found mapping for device %s (%s) → %s\n",
			device.DisplayName(), device.ID, mapping.LayoutName)
	}

	// Get the layout to switch to
	layout, err := uc.layoutRepo.FindByName(ctx, mapping.LayoutName, mapping.LayoutOS)
	if err != nil {
		return fmt.Errorf("layout %s not found: %w", mapping.LayoutName, err)
	}

	fmt.Printf("[Switch] → Switching to layout: %s (OS: %s, ID: %s)\n",
		layout.Name, layout.OS, layout.SystemIdentifier)

	// Switch to the layout
	if err := uc.layoutSwitcher.SwitchLayout(ctx, layout); err != nil {
		return fmt.Errorf("failed to switch layout: %w", err)
	}

	fmt.Printf("[Switch] ✓ Successfully switched to %s\n", layout.Name)

	return nil
}

// SwitchToDefault switches to the system default layout
func (uc *SwitchLayoutUseCase) SwitchToDefault(ctx context.Context) error {
	fmt.Printf("[Switch] → Switching to system default\n")

	// Get system default mapping
	mapping, err := uc.mappingRepo.GetSystemDefault(ctx)
	if err != nil {
		return fmt.Errorf("no system default mapping configured: %w", err)
	}

	// Get the layout
	layout, err := uc.layoutRepo.FindByName(ctx, mapping.LayoutName, mapping.LayoutOS)
	if err != nil {
		return fmt.Errorf("default layout %s not found: %w", mapping.LayoutName, err)
	}

	fmt.Printf("[Switch] → Switching to default layout: %s (OS: %s, ID: %s)\n",
		layout.Name, layout.OS, layout.SystemIdentifier)

	// Switch to the layout
	if err := uc.layoutSwitcher.SwitchLayout(ctx, layout); err != nil {
		return fmt.Errorf("failed to switch to default layout: %w", err)
	}

	fmt.Printf("[Switch] ✓ Successfully switched to %s\n", layout.Name)

	return nil
}
