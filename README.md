# Polykeys

Switch your keyboard layout automatically based on which keyboard you plug in.

## What is this?

You have multiple keyboards (Corne, Lily58, laptop keyboard...) and each one uses a different layout? Polykeys detects which keyboard you just connected and switches to the right layout automatically.

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

Create a config file at `~/.config/polykeys/polykeys.lua`:

```lua
mappings = {
    { "Corne", "US International" },
    { "Lily58", "US" },
    { "system_default", "French" },
}
```

The config can also be at:
- `$XDG_CONFIG_HOME/polykeys/polykeys.lua`
- `$XDG_CONFIG_HOME/polykeys.lua`
- `~/polykeys/polykeys.lua`
- `~/polykeys.lua`

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

- Linux (native)
- macOS (in progress)
- Windows (in progress)

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

## License

AGPL-3.0
