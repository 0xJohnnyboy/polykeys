package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
)

func TestLuaConfigLoader_Load(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "polykeys.lua")

	configContent := `
mappings = {
    { "Corne", "US International" },
    { "Lily58", "US" },
    { "system_default", "French" },
}

enabled = true
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create loader with custom path
	loader := &LuaConfigLoader{
		configPaths: []string{configPath},
	}

	// Load config
	ctx := context.Background()
	config, err := loader.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config
	if !config.Enabled {
		t.Error("Expected enabled to be true")
	}

	if len(config.Mappings) != 3 {
		t.Errorf("Expected 3 mappings, got %d", len(config.Mappings))
	}

	// Verify first mapping
	if config.Mappings[0].DeviceID != "Corne" {
		t.Errorf("Expected first device to be 'Corne', got '%s'", config.Mappings[0].DeviceID)
	}

	if config.Mappings[0].LayoutName != "US International" {
		t.Errorf("Expected first layout to be 'US International', got '%s'", config.Mappings[0].LayoutName)
	}
}

func TestLuaConfigLoader_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "polykeys.lua")

	loader := &LuaConfigLoader{
		configPaths: []string{configPath},
	}

	// Create config
	config := &domain.Config{
		Mappings: []*domain.Mapping{
			domain.NewMapping("046d:c52b", "Logitech Keyboard", "US International", domain.OSLinux),
			domain.NewMapping("system_default", "System Default", "French", domain.OSLinux),
		},
		Enabled: true,
	}

	// Save config
	ctx := context.Background()
	if err := loader.Save(ctx, config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load and verify
	loadedConfig, err := loader.Load(ctx)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if len(loadedConfig.Mappings) != 2 {
		t.Errorf("Expected 2 mappings, got %d", len(loadedConfig.Mappings))
	}
}

func TestLuaConfigLoader_GetConfigPath(t *testing.T) {
	loader := NewLuaConfigLoader()

	path, err := loader.GetConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	if path == "" {
		t.Error("Expected non-empty config path")
	}
}

func TestGetDefaultConfigPaths(t *testing.T) {
	paths := getDefaultConfigPaths()

	if len(paths) == 0 {
		t.Error("Expected at least one default config path")
	}

	// All paths should be absolute
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			t.Errorf("Expected absolute path, got '%s'", path)
		}
	}
}

func TestGetCurrentOS(t *testing.T) {
	os := getCurrentOS()

	validOS := []domain.OperatingSystem{domain.OSLinux, domain.OSMacOS, domain.OSWindows}
	found := false
	for _, validOs := range validOS {
		if os == validOs {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("getCurrentOS returned invalid OS: %s", os)
	}
}
