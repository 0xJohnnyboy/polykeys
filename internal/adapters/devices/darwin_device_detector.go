//go:build darwin

package devices

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
	"github.com/0xJohnnyboy/polykeys/internal/logger"
)

// DarwinDeviceDetector detects USB/HID devices on macOS
type DarwinDeviceDetector struct {
	onConnectedCallback    func(*domain.Device)
	onDisconnectedCallback func(*domain.Device)
	devices                map[string]*domain.Device
	mu                     sync.RWMutex
	stopChan               chan struct{}
	polling                bool
}

// NewDarwinDeviceDetector creates a new macOS device detector
func NewDarwinDeviceDetector() (*DarwinDeviceDetector, error) {
	return &DarwinDeviceDetector{
		devices:  make(map[string]*domain.Device),
		stopChan: make(chan struct{}),
	}, nil
}

// StartMonitoring begins monitoring for device connection/disconnection events
func (d *DarwinDeviceDetector) StartMonitoring(ctx context.Context) error {
	d.polling = true

	// Get initial device list
	if err := d.scanDevices(); err != nil {
		return fmt.Errorf("failed to scan initial devices: %w", err)
	}

	// Start polling for device changes
	go d.pollDevices(ctx)

	return nil
}

// StopMonitoring stops the monitoring process
func (d *DarwinDeviceDetector) StopMonitoring() error {
	d.polling = false
	close(d.stopChan)
	return nil
}

// GetConnectedDevices returns all currently connected keyboard devices
func (d *DarwinDeviceDetector) GetConnectedDevices(ctx context.Context) ([]*domain.Device, error) {
	// Scan for current devices
	if err := d.scanDevices(); err != nil {
		return nil, fmt.Errorf("failed to scan devices: %w", err)
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	devices := make([]*domain.Device, 0, len(d.devices))
	for _, device := range d.devices {
		devices = append(devices, device)
	}

	return devices, nil
}

// OnDeviceConnected registers a callback for device connection events
func (d *DarwinDeviceDetector) OnDeviceConnected(callback func(*domain.Device)) {
	d.onConnectedCallback = callback
}

// OnDeviceDisconnected registers a callback for device disconnection events
func (d *DarwinDeviceDetector) OnDeviceDisconnected(callback func(*domain.Device)) {
	d.onDisconnectedCallback = callback
}

// pollDevices polls for device changes
func (d *DarwinDeviceDetector) pollDevices(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Store previous device IDs and full device info for disconnection
	previousDevices := make(map[string]*domain.Device)
	d.mu.RLock()
	for id, device := range d.devices {
		previousDevices[id] = device
	}
	d.mu.RUnlock()

	pollCount := 0
	for {
		select {
		case <-ticker.C:
			pollCount++
			// Log every 30 polls (every minute) to show we're still alive
			if pollCount%30 == 0 {
				logger.Debug("[Detector] Polling active (count: %d, tracking %d devices)\n", pollCount, len(previousDevices))
			}
			if !d.polling {
				return
			}

			// Scan for current devices
			if err := d.scanDevices(); err != nil {
				logger.Debug("[Detector] Error scanning devices: %v\n", err)
				continue
			}

			// Compare with previous state
			d.mu.RLock()
			currentDevices := make(map[string]*domain.Device)
			for id, device := range d.devices {
				currentDevices[id] = device
			}
			d.mu.RUnlock()

			// Check for new devices (connected)
			for id, device := range currentDevices {
				if _, existed := previousDevices[id]; !existed {
					if d.onConnectedCallback != nil {
						// Run callback in goroutine to avoid blocking polling
						go func(dev *domain.Device) {
							defer func() {
								if r := recover(); r != nil {
									logger.Debug("[Detector] Panic in connected callback: %v\n", r)
								}
							}()
							d.onConnectedCallback(dev)
						}(device)
					}
				}
			}

			// Check for removed devices (disconnected)
			for id := range previousDevices {
				if _, exists := currentDevices[id]; !exists {
					// Device was disconnected - use the stored device info
					device := previousDevices[id]
					if d.onDisconnectedCallback != nil {
						// Run callback in goroutine to avoid blocking polling
						go func(dev *domain.Device) {
							defer func() {
								if r := recover(); r != nil {
									logger.Debug("[Detector] Panic in disconnected callback: %v\n", r)
								}
							}()
							d.onDisconnectedCallback(dev)
						}(device)
					}
				}
			}

			// Update previous devices with full device info
			previousDevices = make(map[string]*domain.Device)
			for id, device := range currentDevices {
				previousDevices[id] = device
			}

		case <-d.stopChan:
			return

		case <-ctx.Done():
			return
		}
	}
}

// SPUSBDevice represents a USB device from system_profiler
type SPUSBDevice struct {
	Name       string `json:"_name"`
	VendorID   string `json:"vendor_id,omitempty"`
	ProductID  string `json:"product_id,omitempty"`
	SerialNum  string `json:"serial_num,omitempty"`
	Items      []SPUSBDevice `json:"_items,omitempty"`
}

// SPUSBDataType represents the root of system_profiler output
type SPUSBDataType struct {
	Devices []SPUSBDevice `json:"SPUSBDataType"`
}

// scanDevices scans for connected keyboard devices using system_profiler
func (d *DarwinDeviceDetector) scanDevices() error {
	// Run system_profiler to get USB devices in JSON format
	cmd := exec.Command("system_profiler", "SPUSBDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		logger.Debug("[Detector] system_profiler failed: %v\n", err)
		return fmt.Errorf("system_profiler failed: %w", err)
	}

	var result SPUSBDataType
	if err := json.Unmarshal(output, &result); err != nil {
		logger.Debug("[Detector] Failed to parse JSON: %v\n", err)
		return fmt.Errorf("failed to parse system_profiler output: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Clear current devices
	d.devices = make(map[string]*domain.Device)

	// Process all USB devices recursively
	for _, usbBus := range result.Devices {
		d.processUSBDevice(usbBus)
	}

	logger.Debug("[Detector] Scan complete: found %d keyboards\n", len(d.devices))

	return nil
}

// processUSBDevice recursively processes USB devices and their children
func (d *DarwinDeviceDetector) processUSBDevice(usbDevice SPUSBDevice) {
	// Process any USB device with VID and PID (don't filter by name)
	// Custom keyboards like Corne, Lily58, etc. don't have "keyboard" in their name
	if usbDevice.VendorID != "" && usbDevice.ProductID != "" {
		// Parse VID and PID (format: "0x046d" -> "046d")
		vendorID := strings.TrimPrefix(strings.ToLower(usbDevice.VendorID), "0x")
		productID := strings.TrimPrefix(strings.ToLower(usbDevice.ProductID), "0x")

		deviceID := vendorID + ":" + productID

		logger.Debug("[Detector] USB Device: %s (%s)\n", usbDevice.Name, deviceID)

		// Skip some known non-keyboard devices
		skipDevices := []string{
			"hub",           // USB hubs
			"camera",        // Cameras
			"bluetooth",     // Bluetooth adapters
			"card reader",   // Card readers
		}

		deviceNameLower := strings.ToLower(usbDevice.Name)
		shouldSkip := false
		for _, skip := range skipDevices {
			if strings.Contains(deviceNameLower, skip) {
				shouldSkip = true
				break
			}
		}

		if shouldSkip {
			logger.Debug("[Detector] Skipping non-keyboard device: %s\n", usbDevice.Name)
		} else {
			logger.Debug("[Detector] Found potential keyboard: %s (%s)\n", usbDevice.Name, deviceID)

			device := domain.NewDevice(vendorID, productID, usbDevice.Name)
			device.ID = deviceID
			device.UpdateLastSeen()

			d.devices[deviceID] = device
		}
	}

	// Recursively process child items
	for _, item := range usbDevice.Items {
		d.processUSBDevice(item)
	}
}
