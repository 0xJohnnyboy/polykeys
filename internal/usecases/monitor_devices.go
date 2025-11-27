package usecases

import (
	"context"
	"fmt"
	"log"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
)

// MonitorDevicesUseCase handles the logic for monitoring device connections
type MonitorDevicesUseCase struct {
	deviceRepo       domain.DeviceRepository
	deviceDetector   domain.DeviceDetector
	switchLayoutUC   *SwitchLayoutUseCase
	enabled          bool
}

// NewMonitorDevicesUseCase creates a new MonitorDevicesUseCase
func NewMonitorDevicesUseCase(
	deviceRepo domain.DeviceRepository,
	deviceDetector domain.DeviceDetector,
	switchLayoutUC *SwitchLayoutUseCase,
) *MonitorDevicesUseCase {
	return &MonitorDevicesUseCase{
		deviceRepo:     deviceRepo,
		deviceDetector: deviceDetector,
		switchLayoutUC: switchLayoutUC,
		enabled:        true,
	}
}

// StartMonitoring begins monitoring for device connection/disconnection events
func (uc *MonitorDevicesUseCase) StartMonitoring(ctx context.Context) error {
	// Register callbacks for device events
	uc.deviceDetector.OnDeviceConnected(func(device *domain.Device) {
		if !uc.enabled {
			return
		}

		log.Printf("Device connected: %s (%s)", device.DisplayName(), device.ID)

		// Update device last seen
		device.UpdateLastSeen()

		// Save device
		if err := uc.deviceRepo.Save(ctx, device); err != nil {
			log.Printf("Error saving device: %v", err)
			return
		}

		// Switch layout for this device
		if err := uc.switchLayoutUC.SwitchForDevice(ctx, device); err != nil {
			log.Printf("Error switching layout for device %s: %v", device.DisplayName(), err)
			return
		}

		log.Printf("Successfully switched layout for device: %s", device.DisplayName())
	})

	uc.deviceDetector.OnDeviceDisconnected(func(device *domain.Device) {
		if !uc.enabled {
			return
		}

		log.Printf("Device disconnected: %s (%s)", device.DisplayName(), device.ID)

		// Switch to default layout when device is disconnected
		if err := uc.switchLayoutUC.SwitchToDefault(ctx); err != nil {
			log.Printf("Error switching to default layout: %v", err)
			return
		}

		log.Printf("Switched to default layout after device disconnection")
	})

	// Start monitoring
	if err := uc.deviceDetector.StartMonitoring(ctx); err != nil {
		return fmt.Errorf("failed to start device monitoring: %w", err)
	}

	log.Println("Device monitoring started")
	return nil
}

// StopMonitoring stops the monitoring process
func (uc *MonitorDevicesUseCase) StopMonitoring() error {
	if err := uc.deviceDetector.StopMonitoring(); err != nil {
		return fmt.Errorf("failed to stop device monitoring: %w", err)
	}

	log.Println("Device monitoring stopped")
	return nil
}

// Enable enables automatic layout switching
func (uc *MonitorDevicesUseCase) Enable() {
	uc.enabled = true
	log.Println("Polykeys enabled")
}

// Disable disables automatic layout switching
func (uc *MonitorDevicesUseCase) Disable() {
	uc.enabled = false
	log.Println("Polykeys disabled")
}

// IsEnabled returns whether automatic layout switching is enabled
func (uc *MonitorDevicesUseCase) IsEnabled() bool {
	return uc.enabled
}

// GetConnectedDevices returns all currently connected keyboard devices
func (uc *MonitorDevicesUseCase) GetConnectedDevices(ctx context.Context) ([]*domain.Device, error) {
	devices, err := uc.deviceDetector.GetConnectedDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connected devices: %w", err)
	}

	return devices, nil
}

// WaitForNextDevice waits for the next device to be connected and returns it
// This is useful for the "polykeys add --detect" command
func (uc *MonitorDevicesUseCase) WaitForNextDevice(ctx context.Context) (*domain.Device, error) {
	deviceChan := make(chan *domain.Device, 1)
	errChan := make(chan error, 1)

	// Temporarily register a callback
	uc.deviceDetector.OnDeviceConnected(func(device *domain.Device) {
		select {
		case deviceChan <- device:
		default:
		}
	})

	// Wait for device or context cancellation
	select {
	case device := <-deviceChan:
		return device, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
