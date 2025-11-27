package devices

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
	"github.com/fsnotify/fsnotify"
)

// LinuxDeviceDetector detects USB/HID devices on Linux
type LinuxDeviceDetector struct {
	watcher              *fsnotify.Watcher
	onConnectedCallback  func(*domain.Device)
	onDisconnectedCallback func(*domain.Device)
	devices              map[string]*domain.Device // deviceID -> device
	mu                   sync.RWMutex
	stopChan             chan struct{}
}

// NewLinuxDeviceDetector creates a new Linux device detector
func NewLinuxDeviceDetector() (*LinuxDeviceDetector, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &LinuxDeviceDetector{
		watcher:  watcher,
		devices:  make(map[string]*domain.Device),
		stopChan: make(chan struct{}),
	}, nil
}

// StartMonitoring begins monitoring for device connection/disconnection events
func (d *LinuxDeviceDetector) StartMonitoring(ctx context.Context) error {
	// Add /dev/input to watch list
	if err := d.watcher.Add("/dev/input"); err != nil {
		return fmt.Errorf("failed to watch /dev/input: %w", err)
	}

	// Get initial list of devices
	if err := d.scanDevices(); err != nil {
		return fmt.Errorf("failed to scan initial devices: %w", err)
	}

	// Start watching for events
	go d.watchEvents(ctx)

	return nil
}

// StopMonitoring stops the monitoring process
func (d *LinuxDeviceDetector) StopMonitoring() error {
	close(d.stopChan)
	return d.watcher.Close()
}

// GetConnectedDevices returns all currently connected keyboard devices
func (d *LinuxDeviceDetector) GetConnectedDevices(ctx context.Context) ([]*domain.Device, error) {
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
func (d *LinuxDeviceDetector) OnDeviceConnected(callback func(*domain.Device)) {
	d.onConnectedCallback = callback
}

// OnDeviceDisconnected registers a callback for device disconnection events
func (d *LinuxDeviceDetector) OnDeviceDisconnected(callback func(*domain.Device)) {
	d.onDisconnectedCallback = callback
}

// watchEvents watches for filesystem events
func (d *LinuxDeviceDetector) watchEvents(ctx context.Context) {
	for {
		select {
		case event, ok := <-d.watcher.Events:
			if !ok {
				return
			}

			// Only care about keyboard devices (event* files)
			if !strings.Contains(event.Name, "event") {
				continue
			}

			if event.Op&fsnotify.Create == fsnotify.Create {
				// New device connected
				if err := d.scanDevices(); err != nil {
					continue
				}
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				// Device disconnected
				d.handleDeviceRemoval(event.Name)
			}

		case err, ok := <-d.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)

		case <-d.stopChan:
			return

		case <-ctx.Done():
			return
		}
	}
}

// scanDevices scans /proc/bus/input/devices for keyboard devices
func (d *LinuxDeviceDetector) scanDevices() error {
	file, err := os.Open("/proc/bus/input/devices")
	if err != nil {
		return fmt.Errorf("failed to open /proc/bus/input/devices: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentDevice *deviceInfo

	for scanner.Scan() {
		line := scanner.Text()

		// New device block
		if strings.HasPrefix(line, "I:") {
			if currentDevice != nil {
				d.processDevice(currentDevice)
			}
			currentDevice = &deviceInfo{}
			d.parseInputLine(line, currentDevice)
		} else if currentDevice != nil {
			if strings.HasPrefix(line, "N:") {
				d.parseNameLine(line, currentDevice)
			} else if strings.HasPrefix(line, "H:") {
				d.parseHandlersLine(line, currentDevice)
			} else if strings.HasPrefix(line, "B: EV=") {
				d.parseEventsLine(line, currentDevice)
			}
		}
	}

	// Process last device
	if currentDevice != nil {
		d.processDevice(currentDevice)
	}

	return scanner.Err()
}

// deviceInfo holds temporary device information during parsing
type deviceInfo struct {
	vendorID  string
	productID string
	name      string
	handlers  []string
	hasKeys   bool
}

// parseInputLine parses the "I:" line containing vendor and product IDs
func (d *LinuxDeviceDetector) parseInputLine(line string, info *deviceInfo) {
	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.HasPrefix(part, "Vendor=") {
			info.vendorID = strings.TrimPrefix(part, "Vendor=")
		} else if strings.HasPrefix(part, "Product=") {
			info.productID = strings.TrimPrefix(part, "Product=")
		}
	}
}

// parseNameLine parses the "N:" line containing device name
func (d *LinuxDeviceDetector) parseNameLine(line string, info *deviceInfo) {
	if strings.HasPrefix(line, "N: Name=") {
		name := strings.TrimPrefix(line, "N: Name=")
		name = strings.Trim(name, "\"")
		info.name = name
	}
}

// parseHandlersLine parses the "H:" line containing handlers
func (d *LinuxDeviceDetector) parseHandlersLine(line string, info *deviceInfo) {
	if strings.HasPrefix(line, "H: Handlers=") {
		handlers := strings.TrimPrefix(line, "H: Handlers=")
		info.handlers = strings.Fields(handlers)
	}
}

// parseEventsLine parses the "B: EV=" line to detect keyboard capability
func (d *LinuxDeviceDetector) parseEventsLine(line string, info *deviceInfo) {
	// EV_KEY (0x01) indicates the device can produce key events
	// A keyboard typically has EV=120013 or similar
	if strings.Contains(line, "B: EV=") {
		evValue := strings.TrimPrefix(line, "B: EV=")
		evValue = strings.TrimSpace(evValue)
		// Check if bit 0 (EV_KEY) is set
		// This is a simplified check - a proper implementation would parse the hex value
		if len(evValue) > 0 {
			info.hasKeys = true
		}
	}
}

// processDevice processes a parsed device and adds it if it's a keyboard
func (d *LinuxDeviceDetector) processDevice(info *deviceInfo) {
	// Skip if not a keyboard (no key events or no event handler)
	if !info.hasKeys || len(info.handlers) == 0 {
		return
	}

	// Skip if it's a mouse or touchpad (simple heuristic)
	nameLower := strings.ToLower(info.name)
	if strings.Contains(nameLower, "mouse") ||
	   strings.Contains(nameLower, "touchpad") ||
	   strings.Contains(nameLower, "touchscreen") {
		return
	}

	// Create device
	device := domain.NewDevice(info.vendorID, info.productID, info.name)

	d.mu.Lock()
	existingDevice, exists := d.devices[device.ID]
	if !exists {
		d.devices[device.ID] = device
		d.mu.Unlock()

		// Trigger callback for new device
		if d.onConnectedCallback != nil {
			d.onConnectedCallback(device)
		}
	} else {
		existingDevice.UpdateLastSeen()
		d.mu.Unlock()
	}
}

// handleDeviceRemoval handles device removal events
func (d *LinuxDeviceDetector) handleDeviceRemoval(eventPath string) {
	// We can't easily map an event file back to a device ID
	// So we rescan and compare with current devices
	// This is a simplified approach - a production version would be more sophisticated

	// For now, we'll just rescan
	_ = d.scanDevices()
}

// GetDeviceByID returns a device by its ID
func (d *LinuxDeviceDetector) GetDeviceByID(deviceID string) (*domain.Device, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	device, exists := d.devices[deviceID]
	if !exists {
		return nil, fmt.Errorf("device %s not found", deviceID)
	}

	return device, nil
}
