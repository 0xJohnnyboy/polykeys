package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/0xJohnnyboy/polykeys/internal/domain"
	lua "github.com/yuin/gopher-lua"
)

// LuaConfigLoader loads configuration from Lua files
type LuaConfigLoader struct {
	configPaths []string
}

// NewLuaConfigLoader creates a new LuaConfigLoader
func NewLuaConfigLoader() *LuaConfigLoader {
	return &LuaConfigLoader{
		configPaths: getDefaultConfigPaths(),
	}
}

// getDefaultConfigPaths returns the default configuration file paths in order of priority
func getDefaultConfigPaths() []string {
	paths := make([]string, 0, 4)

	// XDG_CONFIG_HOME/polykeys/polykeys.lua
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		paths = append(paths, filepath.Join(xdgConfig, "polykeys", "polykeys.lua"))
	}

	// HOME/.config/polykeys/polykeys.lua (fallback for XDG)
	if home := os.Getenv("HOME"); home != "" {
		paths = append(paths, filepath.Join(home, ".config", "polykeys", "polykeys.lua"))
	}

	// XDG_CONFIG_HOME/polykeys.lua
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		paths = append(paths, filepath.Join(xdgConfig, "polykeys.lua"))
	}

	// HOME/polykeys/polykeys.lua
	if home := os.Getenv("HOME"); home != "" {
		paths = append(paths, filepath.Join(home, "polykeys", "polykeys.lua"))
		// HOME/polykeys.lua
		paths = append(paths, filepath.Join(home, "polykeys.lua"))
	}

	return paths
}

// GetConfigPath returns the path to the configuration file
// Returns the first existing file, or the first path if none exist
func (l *LuaConfigLoader) GetConfigPath() (string, error) {
	for _, path := range l.configPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// If no config exists, return the first path (preferred location)
	if len(l.configPaths) > 0 {
		return l.configPaths[0], nil
	}

	return "", fmt.Errorf("no config paths available")
}

// Load loads the configuration from the Lua file
func (l *LuaConfigLoader) Load(ctx context.Context) (*domain.Config, error) {
	configPath, err := l.findExistingConfig()
	if err != nil {
		return nil, fmt.Errorf("no configuration file found: %w", err)
	}

	L := lua.NewState()
	defer L.Close()

	// Execute the Lua config file
	if err := L.DoFile(configPath); err != nil {
		return nil, fmt.Errorf("error executing Lua config: %w", err)
	}

	// Get the mappings table
	mappingsTable := L.GetGlobal("mappings")
	if mappingsTable.Type() != lua.LTTable {
		return nil, fmt.Errorf("'mappings' is not defined or is not a table in config")
	}

	// Parse mappings
	mappings, err := l.parseMappings(L, mappingsTable.(*lua.LTable))
	if err != nil {
		return nil, fmt.Errorf("error parsing mappings: %w", err)
	}

	// Check for enabled flag (default to true)
	enabled := true
	if enabledValue := L.GetGlobal("enabled"); enabledValue.Type() == lua.LTBool {
		enabled = bool(enabledValue.(lua.LBool))
	}

	return &domain.Config{
		Mappings: mappings,
		Enabled:  enabled,
	}, nil
}

// findExistingConfig finds the first existing configuration file
func (l *LuaConfigLoader) findExistingConfig() (string, error) {
	for _, path := range l.configPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no configuration file found in any of the default locations")
}

// parseMappings parses the mappings table from Lua
func (l *LuaConfigLoader) parseMappings(L *lua.LState, table *lua.LTable) ([]*domain.Mapping, error) {
	mappings := make([]*domain.Mapping, 0)

	// Get current OS
	currentOS := getCurrentOS()

	// Iterate over the mappings array
	table.ForEach(func(key, value lua.LValue) {
		if value.Type() != lua.LTTable {
			return
		}

		mappingTable := value.(*lua.LTable)

		// Get device name/ID (first element)
		deviceID := mappingTable.RawGetInt(1).String()
		if deviceID == "" {
			return
		}

		// Get layout name (second element)
		layoutName := mappingTable.RawGetInt(2).String()
		if layoutName == "" {
			return
		}

		// Create mapping
		mapping := domain.NewMapping(deviceID, deviceID, layoutName, currentOS)
		mappings = append(mappings, mapping)
	})

	return mappings, nil
}

// Save saves the configuration to the Lua file
func (l *LuaConfigLoader) Save(ctx context.Context, config *domain.Config) error {
	configPath, err := l.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Generate Lua content
	content := l.generateLuaConfig(config)

	// Write to file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// generateLuaConfig generates Lua configuration content
func (l *LuaConfigLoader) generateLuaConfig(config *domain.Config) string {
	content := "-- Polykeys configuration\n\n"
	content += "mappings = {\n"

	for _, mapping := range config.Mappings {
		deviceName := mapping.DeviceDisplayName
		if deviceName == "" {
			deviceName = mapping.DeviceID
		}
		content += fmt.Sprintf("    { \"%s\", \"%s\" },\n", deviceName, mapping.LayoutName)
	}

	content += "}\n\n"
	content += fmt.Sprintf("enabled = %v\n", config.Enabled)

	return content
}

// getCurrentOS returns the current operating system as a domain.OperatingSystem
func getCurrentOS() domain.OperatingSystem {
	switch runtime.GOOS {
	case "linux":
		return domain.OSLinux
	case "darwin":
		return domain.OSMacOS
	case "windows":
		return domain.OSWindows
	default:
		return domain.OSLinux
	}
}
