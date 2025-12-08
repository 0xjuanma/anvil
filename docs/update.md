# Update Command

The `anvil update` command updates your Anvil installation to the latest version.

## Usage

```bash
anvil update [flags]
```

### Flags

- `--dry-run`: Preview the update process without actually updating
- `--help`: Show help information

## What It Does

1. Verifies macOS platform and curl availability
2. Downloads the latest installation script from GitHub releases
3. Executes the script to install the latest binary
4. Validates the update was successful

The update process automatically detects your system architecture (Intel or Apple Silicon) and downloads the appropriate binary.

## Examples

```bash
anvil update              # Update to latest version
anvil update --dry-run    # Preview update without changes
```

## Notes

- **Terminal Restart**: May need to restart terminal for changes to take effect
- **Admin Permissions**: Script may request admin permissions for `/usr/local/bin/`
- **Internet Required**: Active internet connection required
- **Configuration Preserved**: `~/.anvil/settings.yaml` remains unchanged

## Troubleshooting

- **curl not available**: Install via `brew install curl`
- **Update script failed**: Check internet connection or try again (GitHub rate limits)
- **Permission denied**: Script will request admin permissions automatically
- **Update doesn't take effect**: Restart terminal or run `hash -r`

## Related Documentation

- [Doctor Command](doctor.md)
