# Troubleshooting

This document lists all error codes used by Polykeys and their meanings.

## Error Code Ranges

- **PK_000-099**: General errors
- **PK_100-199**: Layout switching errors
- **PK_200-299**: Device detection errors
- **PK_300-399**: Configuration errors
- **PK_400-499**: Use case / mapping errors
- **PK_500-599**: Repository errors

## Error Codes Reference

### General Errors (000-099)

| Code | Description | Common Causes | Solution |
|------|-------------|---------------|----------|
| `PK_000` | Unknown error | Unexpected error condition | Check logs for details |

### Layout Errors (100-199)

| Code | Description | Common Causes | Solution |
|------|-------------|---------------|----------|
| `PK_100` | Layout not found | Layout not installed on system | Install the keyboard layout in system settings |
| `PK_101` | Failed to enable layout | Layout exists but can't be enabled | Check system permissions, try enabling layout manually |
| `PK_102` | Failed to select layout | System API call failed | Restart polykeys daemon, check system logs |
| `PK_103` | Invalid OS for layout | Layout config specifies wrong OS | Verify layout configuration matches your OS |
| `PK_104` | String conversion failed | Invalid characters in layout name | Check layout identifier in config |
| `PK_105` | Invalid layout identifier | Malformed layout identifier | Verify layout identifier format in config |

### Device Errors (200-299)

| Code | Description | Common Causes | Solution |
|------|-------------|---------------|----------|
| `PK_200` | Device not found | Device disconnected or not recognized | Check USB connection, verify device ID |
| `PK_201` | Device detection failed | System API error | Check permissions, restart daemon |
| `PK_202` | Device scan failed | Unable to enumerate devices | Check system device access permissions |

### Configuration Errors (300-399)

| Code | Description | Common Causes | Solution |
|------|-------------|---------------|----------|
| `PK_300` | Config load failed | Config file missing or unreadable | Verify config file exists at expected path |
| `PK_301` | Config parse failed | Invalid Lua syntax | Check config file syntax, see example config |
| `PK_302` | Config save failed | Permission denied or disk full | Check file permissions and disk space |
| `PK_303` | Config not found | No config file exists | Run `polykeys add --detect` to create initial config |

### Mapping Errors (400-499)

| Code | Description | Common Causes | Solution |
|------|-------------|---------------|----------|
| `PK_400` | Mapping not found | No mapping exists for device | Run `polykeys add` to create mapping |
| `PK_401` | Mapping already exists | Duplicate mapping for device | Use `polykeys remove` first or edit config directly |
| `PK_402` | Invalid mapping | Malformed mapping entry | Check mapping format in config file |

### Repository Errors (500-599)

| Code | Description | Common Causes | Solution |
|------|-------------|---------------|----------|
| `PK_500` | Repository operation failed | Internal database error | Report as bug with error details |
| `PK_501` | Repository entry not found | Requested item doesn't exist | Verify the item exists before accessing |

## Getting Help

If you encounter an error not listed here or need additional help:

1. Check the logs with `--debug` flag: `polykeys --debug` or `polykeysd --debug`
2. Search existing issues: https://github.com/0xJohnnyboy/polykeys/issues
3. Open a new issue with:
   - Error code
   - Full error message
   - Debug logs
   - Your OS and configuration
