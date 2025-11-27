# Polykeys

Switch your keyboard layout automatically based on which keyboard you plug in.

## What is this?

You have multiple keyboards (Corne, Lily58, laptop keyboard...) and each one uses a different layout? Polykeys detects which keyboard you just connected and switches to the right layout automatically.

## Installation

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

- Linux
- macOS
- Windows

## License

AGPL-3.0
