package infrastructure

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
)

// InMemoryDeviceRepository is a simple in-memory implementation
type InMemoryDeviceRepository struct {
	devices map[string]*domain.Device
	mu      sync.RWMutex
}

func NewInMemoryDeviceRepository() *InMemoryDeviceRepository {
	return &InMemoryDeviceRepository{
		devices: make(map[string]*domain.Device),
	}
}

func (r *InMemoryDeviceRepository) Save(ctx context.Context, device *domain.Device) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices[device.ID] = device
	return nil
}

func (r *InMemoryDeviceRepository) FindByID(ctx context.Context, id string) (*domain.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	device, exists := r.devices[id]
	if !exists {
		return nil, fmt.Errorf("device not found")
	}
	return device, nil
}

func (r *InMemoryDeviceRepository) FindAll(ctx context.Context) ([]*domain.Device, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	devices := make([]*domain.Device, 0, len(r.devices))
	for _, device := range r.devices {
		devices = append(devices, device)
	}
	return devices, nil
}

func (r *InMemoryDeviceRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.devices, id)
	return nil
}

// InMemoryMappingRepository is a simple in-memory implementation
type InMemoryMappingRepository struct {
	mappings map[string]*domain.Mapping
	mu       sync.RWMutex
}

func NewInMemoryMappingRepository() *InMemoryMappingRepository {
	return &InMemoryMappingRepository{
		mappings: make(map[string]*domain.Mapping),
	}
}

func (r *InMemoryMappingRepository) Save(ctx context.Context, mapping *domain.Mapping) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mappings[mapping.DeviceID] = mapping
	return nil
}

func (r *InMemoryMappingRepository) FindByDeviceID(ctx context.Context, deviceID string) (*domain.Mapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mapping, exists := r.mappings[deviceID]
	if !exists {
		return nil, fmt.Errorf("mapping not found")
	}
	return mapping, nil
}

func (r *InMemoryMappingRepository) FindAll(ctx context.Context) ([]*domain.Mapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mappings := make([]*domain.Mapping, 0, len(r.mappings))
	for _, mapping := range r.mappings {
		mappings = append(mappings, mapping)
	}
	return mappings, nil
}

func (r *InMemoryMappingRepository) Delete(ctx context.Context, deviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.mappings, deviceID)
	return nil
}

func (r *InMemoryMappingRepository) GetSystemDefault(ctx context.Context) (*domain.Mapping, error) {
	return r.FindByDeviceID(ctx, "system_default")
}

// InMemoryLayoutRepository is a simple in-memory implementation
type InMemoryLayoutRepository struct {
	layouts map[string]*domain.KeyboardLayout
	mu      sync.RWMutex
}

func NewInMemoryLayoutRepository() *InMemoryLayoutRepository {
	repo := &InMemoryLayoutRepository{
		layouts: make(map[string]*domain.KeyboardLayout),
	}
	// Pre-populate with common layouts
	repo.populateDefaultLayouts()
	return repo
}

func (r *InMemoryLayoutRepository) populateDefaultLayouts() {
	// Determine current OS
	var os domain.OperatingSystem
	switch getCurrentOS() {
	case "linux":
		os = domain.OSLinux
	case "darwin":
		os = domain.OSMacOS
	case "windows":
		os = domain.OSWindows
	default:
		os = domain.OSLinux
	}

	// Add common layouts based on OS
	layouts := getDefaultLayoutsForOS(os)
	for _, layout := range layouts {
		r.layouts[layout.ID] = layout
	}
}

func getCurrentOS() string {
	// Use runtime.GOOS to detect the actual OS
	switch runtime.GOOS {
	case "linux":
		return "linux"
	case "darwin":
		return "darwin"
	case "windows":
		return "windows"
	default:
		return "linux"
	}
}

func getDefaultLayoutsForOS(os domain.OperatingSystem) []*domain.KeyboardLayout {
	switch os {
	case domain.OSLinux:
		return []*domain.KeyboardLayout{
			domain.NewKeyboardLayout(domain.LayoutUSQwerty, os, "us"),
			domain.NewKeyboardLayout(domain.LayoutUSInternational, os, "us -variant intl"),
			domain.NewKeyboardLayout(domain.LayoutUSInternationalDeadKeys, os, "us -variant altgr-intl"),
			domain.NewKeyboardLayout(domain.LayoutFrenchAzerty, os, "fr"),
			domain.NewKeyboardLayout(domain.LayoutUKQwerty, os, "gb"),
			domain.NewKeyboardLayout(domain.LayoutColemak, os, "us -variant colemak"),
			domain.NewKeyboardLayout(domain.LayoutDvorak, os, "us -variant dvorak"),
		}
	case domain.OSWindows:
		return []*domain.KeyboardLayout{
			domain.NewKeyboardLayout(domain.LayoutUSQwerty, os, "00000409"),
			domain.NewKeyboardLayout(domain.LayoutUSInternational, os, "00020409"),
			domain.NewKeyboardLayout(domain.LayoutUSInternationalDeadKeys, os, "00020409"),
			domain.NewKeyboardLayout(domain.LayoutFrenchAzerty, os, "0000040c"),
			domain.NewKeyboardLayout(domain.LayoutUKQwerty, os, "00000809"),
			domain.NewKeyboardLayout(domain.LayoutColemak, os, "00000409"),
			domain.NewKeyboardLayout(domain.LayoutDvorak, os, "00010409"),
			domain.NewKeyboardLayout(domain.LayoutGerman, os, "00000407"),
			domain.NewKeyboardLayout(domain.LayoutSpanish, os, "0000040a"),
			domain.NewKeyboardLayout(domain.LayoutItalian, os, "00000410"),
			domain.NewKeyboardLayout(domain.LayoutPortuguese, os, "00000816"),
			domain.NewKeyboardLayout(domain.LayoutRussian, os, "00000419"),
			domain.NewKeyboardLayout(domain.LayoutJapanese, os, "00000411"),
		}
	case domain.OSMacOS:
		return []*domain.KeyboardLayout{
			domain.NewKeyboardLayout(domain.LayoutUSQwerty, os, "com.apple.keylayout.US"),
			domain.NewKeyboardLayout(domain.LayoutUSInternational, os, "com.apple.keylayout.USInternational-PC"),
			domain.NewKeyboardLayout(domain.LayoutUSInternationalDeadKeys, os, "com.apple.keylayout.USInternational-PC"),
			domain.NewKeyboardLayout(domain.LayoutFrenchAzerty, os, "com.apple.keylayout.French"),
			domain.NewKeyboardLayout(domain.LayoutUKQwerty, os, "com.apple.keylayout.British"),
			domain.NewKeyboardLayout(domain.LayoutColemak, os, "com.apple.keylayout.Colemak"),
			domain.NewKeyboardLayout(domain.LayoutDvorak, os, "com.apple.keylayout.Dvorak"),
			domain.NewKeyboardLayout(domain.LayoutGerman, os, "com.apple.keylayout.German"),
			domain.NewKeyboardLayout(domain.LayoutSpanish, os, "com.apple.keylayout.Spanish"),
			domain.NewKeyboardLayout(domain.LayoutItalian, os, "com.apple.keylayout.Italian"),
			domain.NewKeyboardLayout(domain.LayoutPortuguese, os, "com.apple.keylayout.Portuguese"),
			domain.NewKeyboardLayout(domain.LayoutRussian, os, "com.apple.keylayout.Russian"),
			domain.NewKeyboardLayout(domain.LayoutJapanese, os, "com.apple.inputmethod.Kotoeri.Japanese"),
		}
	default:
		return []*domain.KeyboardLayout{}
	}
}

func (r *InMemoryLayoutRepository) Save(ctx context.Context, layout *domain.KeyboardLayout) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.layouts[layout.ID] = layout
	return nil
}

func (r *InMemoryLayoutRepository) FindByName(ctx context.Context, name string, os domain.OperatingSystem) (*domain.KeyboardLayout, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Search by name and OS
	for _, layout := range r.layouts {
		if layout.Name == name && layout.OS == os {
			return layout, nil
		}
	}

	return nil, fmt.Errorf("layout not found: %s for %s", name, os)
}

func (r *InMemoryLayoutRepository) FindByOS(ctx context.Context, os domain.OperatingSystem) ([]*domain.KeyboardLayout, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	layouts := make([]*domain.KeyboardLayout, 0)
	for _, layout := range r.layouts {
		if layout.OS == os {
			layouts = append(layouts, layout)
		}
	}

	return layouts, nil
}

func (r *InMemoryLayoutRepository) FindAll(ctx context.Context) ([]*domain.KeyboardLayout, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	layouts := make([]*domain.KeyboardLayout, 0, len(r.layouts))
	for _, layout := range r.layouts {
		layouts = append(layouts, layout)
	}

	return layouts, nil
}
