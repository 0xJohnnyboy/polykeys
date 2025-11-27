//go:build windows

package devices

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
	"github.com/0xJohnnyboy/polykeys/internal/logger"
	"github.com/StackExchange/wmi"
)

// WindowsDeviceDetector detects USB/HID devices on Windows
type WindowsDeviceDetector struct {
	onConnectedCallback    func(*domain.Device)
	onDisconnectedCallback func(*domain.Device)
	devices                map[string]*domain.Device
	mu                     sync.RWMutex
	stopChan               chan struct{}
	polling                bool
}

// NewWindowsDeviceDetector creates a new Windows device detector
func NewWindowsDeviceDetector() (*WindowsDeviceDetector, error) {
	return &WindowsDeviceDetector{
		devices:  make(map[string]*domain.Device),
		stopChan: make(chan struct{}),
	}, nil
}

// StartMonitoring begins monitoring for device connection/disconnection events
func (d *WindowsDeviceDetector) StartMonitoring(ctx context.Context) error {
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
func (d *WindowsDeviceDetector) StopMonitoring() error {
	d.polling = false
	close(d.stopChan)
	return nil
}

// GetConnectedDevices returns all currently connected keyboard devices
func (d *WindowsDeviceDetector) GetConnectedDevices(ctx context.Context) ([]*domain.Device, error) {
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
func (d *WindowsDeviceDetector) OnDeviceConnected(callback func(*domain.Device)) {
	d.onConnectedCallback = callback
}

// OnDeviceDisconnected registers a callback for device disconnection events
func (d *WindowsDeviceDetector) OnDeviceDisconnected(callback func(*domain.Device)) {
	d.onDisconnectedCallback = callback
}

// pollDevices polls for device changes
func (d *WindowsDeviceDetector) pollDevices(ctx context.Context) {
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
				log.Printf("[Detector] Polling active (count: %d, tracking %d devices)", pollCount, len(previousDevices))
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
									log.Printf("[Detector] Panic in connected callback: %v", r)
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
									log.Printf("[Detector] Panic in disconnected callback: %v", r)
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

// Win32_USBHub represents a WMI USB Hub
type Win32_USBHub struct {
	DeviceID string
	Name     string
}

// Win32_Keyboard represents a WMI Keyboard
type Win32_Keyboard struct {
	DeviceID    string
	Name        string
	Description string
}

// scanDevices scans for connected keyboard devices using WMI
func (d *WindowsDeviceDetector) scanDevices() error {
	var keyboards []Win32_Keyboard

	// Query WMI for keyboards
	query := "SELECT DeviceID, Name, Description FROM Win32_Keyboard"
	err := wmi.Query(query, &keyboards)
	if err != nil {
		return fmt.Errorf("WMI query failed: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// Clear current devices
	d.devices = make(map[string]*domain.Device)

	// Process each keyboard
	for _, kb := range keyboards {
		// Try to extract VID/PID from DeviceID
		// DeviceID format: something like "HID\VID_xxxx&PID_yyyy\..."
		vendorID, productID := extractVIDPID(kb.DeviceID)

		var deviceID string
		if vendorID != "" && productID != "" {
			deviceID = vendorID + ":" + productID
		} else {
			// Fallback to using the device name as ID
			deviceID = kb.DeviceID
		}

		// Skip if it's the standard PS/2 keyboard (internal laptop keyboard)
		if strings.Contains(kb.Name, "PS/2") || strings.Contains(kb.Name, "Standard") {
			continue
		}

		device := domain.NewDevice(vendorID, productID, kb.Name)
		device.ID = deviceID
		device.UpdateLastSeen()

		d.devices[deviceID] = device
	}

	return nil
}

// extractVIDPID extracts VID and PID from a Windows Device ID
func extractVIDPID(deviceID string) (string, string) {
	// Match patterns like VID_xxxx&PID_yyyy or VID_xxxx and PID_yyyy
	vidRegex := regexp.MustCompile(`VID[_&]([0-9A-Fa-f]{4})`)
	pidRegex := regexp.MustCompile(`PID[_&]([0-9A-Fa-f]{4})`)

	vidMatch := vidRegex.FindStringSubmatch(deviceID)
	pidMatch := pidRegex.FindStringSubmatch(deviceID)

	var vid, pid string
	if len(vidMatch) > 1 {
		vid = strings.ToLower(vidMatch[1])
	}
	if len(pidMatch) > 1 {
		pid = strings.ToLower(pidMatch[1])
	}

	return vid, pid
}
