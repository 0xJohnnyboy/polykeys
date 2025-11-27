package infrastructure

import (
	"fmt"

	"github.com/0xJohnnyboy/polykeys/internal/adapters/config"
	"github.com/0xJohnnyboy/polykeys/internal/domain"
	"github.com/0xJohnnyboy/polykeys/internal/usecases"
)

// App holds all the application components
type App struct {
	ConfigLoader       domain.ConfigLoader
	DeviceDetector     domain.DeviceDetector
	LayoutSwitcher     domain.LayoutSwitcher
	LayoutRepo         domain.LayoutRepository
	SwitchLayoutUC     *usecases.SwitchLayoutUseCase
	ManageMappingsUC   *usecases.ManageMappingsUseCase
	MonitorDevicesUC   *usecases.MonitorDevicesUseCase
}

// NewApp creates and initializes the application with all dependencies
func NewApp() (*App, error) {
	// Initialize config loader
	configLoader := config.NewLuaConfigLoader()

	// Initialize platform-specific adapters
	deviceDetector, err := createDeviceDetector()
	if err != nil {
		return nil, fmt.Errorf("failed to create device detector: %w", err)
	}

	layoutSwitcher, err := createLayoutSwitcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create layout switcher: %w", err)
	}

	// Create in-memory repositories (for now)
	deviceRepo := NewInMemoryDeviceRepository()
	mappingRepo := NewInMemoryMappingRepository()
	layoutRepo := NewInMemoryLayoutRepository()

	// Initialize use cases
	switchLayoutUC := usecases.NewSwitchLayoutUseCase(mappingRepo, layoutRepo, layoutSwitcher)
	manageMappingsUC := usecases.NewManageMappingsUseCase(deviceRepo, mappingRepo, layoutRepo, configLoader)
	monitorDevicesUC := usecases.NewMonitorDevicesUseCase(deviceRepo, deviceDetector, switchLayoutUC)

	return &App{
		ConfigLoader:       configLoader,
		DeviceDetector:     deviceDetector,
		LayoutSwitcher:     layoutSwitcher,
		LayoutRepo:         layoutRepo,
		SwitchLayoutUC:     switchLayoutUC,
		ManageMappingsUC:   manageMappingsUC,
		MonitorDevicesUC:   monitorDevicesUC,
	}, nil
}

// createDeviceDetector creates a platform-specific device detector
func createDeviceDetector() (domain.DeviceDetector, error) {
	return createPlatformDeviceDetector()
}

// createLayoutSwitcher creates a platform-specific layout switcher
func createLayoutSwitcher() (domain.LayoutSwitcher, error) {
	return createPlatformLayoutSwitcher()
}
