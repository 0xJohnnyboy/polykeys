# Polykeys

> **Status:** Alpha v1.0.0-alpha.2 - Fully functional on Windows, Linux support in progress

Switch your keyboard layout automatically based on which keyboard you plug in.

## What is this?

You have multiple keyboards (Corne, Lily58, laptop keyboard...) and each one uses a different layout? Polykeys detects which keyboard you just connected and switches to the right layout automatically.

**Current features:**
- Automatic layout switching on device connection/disconnection
- Interactive device detection with `polykeys add --detect`
- System default fallback when no device is connected
- Real-time device monitoring
- Lua-based configuration

## Installation

### From source

```bash
# Clone the repository
git clone https://github.com/0xJohnnyboy/polykeys.git
cd polykeys

# Build for your platform
make build

# Or build for all platforms
make build-all

# Install to system
make install
```

### From Go
```bash
go install github.com/0xJohnnyboy/polykeys/cmd/polykeysd@latest
go install github.com/0xJohnnyboy/polykeys/cmd/polykeys@latest
```

## Configuration

Create a config file with device-to-layout mappings:

```lua
-- Format: { "alias", "deviceID", "layout" }
mappings = {
    { "Corne", "4653:0004", "US International" },
    { "Lily58", "1209:bb58", "US" },
    { "Logitech K380", "046d:c52b", "US" },

    -- Fallback when no device matches
    { "System Default", "system_default", "French AZERTY" },
}
```

**Device ID format:** `VID:PID` (Vendor ID:Product ID in hex, lowercase)
**Tip:** Use `polykeys add --detect` to automatically detect and add keyboards

### Config file locations

**Linux/macOS:**
- `$XDG_CONFIG_HOME/polykeys/polykeys.lua` (preferred)
- `~/.config/polykeys/polykeys.lua`
- `$XDG_CONFIG_HOME/polykeys.lua`
- `~/polykeys/polykeys.lua`
- `~/polykeys.lua`

**Windows:**
- `%APPDATA%\polykeys\polykeys.lua` (preferred)
- `%USERPROFILE%\.config\polykeys\polykeys.lua`
- `%LOCALAPPDATA%\polykeys\polykeys.lua`
- `%USERPROFILE%\polykeys\polykeys.lua`
- `%USERPROFILE%\polykeys.lua`

## Usage

Start the daemon:
```bash
polykeysd
```

Add a new keyboard (interactive):
```bash
polykeys add --detect
# Then plug in your keyboard
```

See what's happening (useful for debugging):
```bash
polykeys logs -f
```

List your mappings:
```bash
polykeys list
```

## Supported platforms

- ✅ **Windows** - Fully functional (alpha)
- ✅ **macOS** - Implemented (requires testing)
- 🚧 **Linux** - In progress

### Platform implementation details

**Windows:**
- Device detection via WMI queries
- Layout switching using Windows Keyboard Layout API
- Polling-based detection (every 2 seconds)
- Automatic switch to default layout on device disconnection

**macOS:**
- Device detection via `system_profiler` USB enumeration
- Layout switching using Carbon Text Input Sources API (CGO)
- Polling-based detection (every 2 seconds)
- Requires native compilation on macOS with CGO enabled
- Supports standard macOS keyboard layouts and input methods

## Important notes

### WSL (Windows Subsystem for Linux)

Polykeys **does not work on WSL** due to limitations in device access:
- WSL does not expose `/dev/input` for USB device monitoring
- USB events are not propagated to the WSL environment
- Layout switching commands may not affect the Windows host

**Workarounds:**
- Run polykeys natively on Windows (use the Windows build)
- Use a native Linux installation (dual boot or VM)
- Use WSL2 with usbipd (complex setup, not recommended)

### Permissions

On Linux, you may need appropriate permissions to access `/dev/input`. If the daemon fails to start, try:
- Adding your user to the `input` group: `sudo usermod -a -G input $USER`
- Running the daemon with `sudo` (not recommended for production)

## Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build for specific platform
make build-linux
make build-windows
make build-darwin

# Run tests
make test

# Show all available targets
make help
```

## Troubleshooting

Having issues? Check the [troubleshooting guide](TROUBLESHOOTING.md) for common error codes and solutions.

All errors include a code (e.g., `PK_100`) to help identify and resolve issues quickly.

## License

AGPL-3.0
