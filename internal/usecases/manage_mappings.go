package usecases

import (
	"context"
	"fmt"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
)

// ManageMappingsUseCase handles the logic for managing device-to-layout mappings
type ManageMappingsUseCase struct {
	deviceRepo  domain.DeviceRepository
	mappingRepo domain.MappingRepository
	layoutRepo  domain.LayoutRepository
	configLoader domain.ConfigLoader
}

// NewManageMappingsUseCase creates a new ManageMappingsUseCase
func NewManageMappingsUseCase(
	deviceRepo domain.DeviceRepository,
	mappingRepo domain.MappingRepository,
	layoutRepo domain.LayoutRepository,
	configLoader domain.ConfigLoader,
) *ManageMappingsUseCase {
	return &ManageMappingsUseCase{
		deviceRepo:   deviceRepo,
		mappingRepo:  mappingRepo,
		layoutRepo:   layoutRepo,
		configLoader: configLoader,
	}
}

// AddMapping creates a new mapping between a device and a layout
func (uc *ManageMappingsUseCase) AddMapping(
	ctx context.Context,
	device *domain.Device,
	layoutName string,
	layoutOS domain.OperatingSystem,
) error {
	// Verify the layout exists
	layout, err := uc.layoutRepo.FindByName(ctx, layoutName, layoutOS)
	if err != nil {
		return fmt.Errorf("layout %s not found: %w", layoutName, err)
	}

	// Save the device if not already saved
	if err := uc.deviceRepo.Save(ctx, device); err != nil {
		return fmt.Errorf("failed to save device: %w", err)
	}

	// Create and save the mapping
	mapping := domain.NewMapping(device.ID, device.DisplayName(), layout.Name, layout.OS)
	if err := uc.mappingRepo.Save(ctx, mapping); err != nil {
		return fmt.Errorf("failed to save mapping: %w", err)
	}

	return nil
}

// RemoveMapping removes a mapping for a device
func (uc *ManageMappingsUseCase) RemoveMapping(ctx context.Context, deviceID string) error {
	// Check if mapping exists
	_, err := uc.mappingRepo.FindByDeviceID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("mapping not found for device %s: %w", deviceID, err)
	}

	// Delete the mapping
	if err := uc.mappingRepo.Delete(ctx, deviceID); err != nil {
		return fmt.Errorf("failed to delete mapping: %w", err)
	}

	return nil
}

// ListMappings returns all configured mappings
func (uc *ManageMappingsUseCase) ListMappings(ctx context.Context) ([]*domain.Mapping, error) {
	mappings, err := uc.mappingRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve mappings: %w", err)
	}

	return mappings, nil
}

// SetSystemDefault sets the default layout for unmapped devices
func (uc *ManageMappingsUseCase) SetSystemDefault(
	ctx context.Context,
	layoutName string,
	layoutOS domain.OperatingSystem,
) error {
	// Verify the layout exists
	layout, err := uc.layoutRepo.FindByName(ctx, layoutName, layoutOS)
	if err != nil {
		return fmt.Errorf("layout %s not found: %w", layoutName, err)
	}

	// Create system default mapping
	mapping := domain.NewMapping("system_default", "System Default", layout.Name, layout.OS)
	if err := uc.mappingRepo.Save(ctx, mapping); err != nil {
		return fmt.Errorf("failed to save system default: %w", err)
	}

	return nil
}

// LoadFromConfig loads mappings from the configuration file
func (uc *ManageMappingsUseCase) LoadFromConfig(ctx context.Context) error {
	config, err := uc.configLoader.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Save all mappings from config
	for _, mapping := range config.Mappings {
		if err := uc.mappingRepo.Save(ctx, mapping); err != nil {
			return fmt.Errorf("failed to save mapping from config: %w", err)
		}
	}

	return nil
}

// SaveToConfig saves current mappings to the configuration file
func (uc *ManageMappingsUseCase) SaveToConfig(ctx context.Context) error {
	mappings, err := uc.mappingRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve mappings: %w", err)
	}

	config := &domain.Config{
		Mappings: mappings,
		Enabled:  true,
	}

	if err := uc.configLoader.Save(ctx, config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}
