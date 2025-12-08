# Clean Command

The `anvil clean` command provides safe cleanup of temporary files, archives, and cached configurations while preserving your essential settings.

## Usage

```bash
anvil clean [flags]
```

### Flags

- `--force`: Skip confirmation prompts
- `--dry-run`: Preview what would be cleaned without deletion

## What Gets Cleaned

### Cleaned Content

- **temp/ directory contents**: All pulled configurations waiting to be synced
- **archive/ directory contents**: Old archived configurations and backups
- **dotfiles/ directory**: Completely removed for clean git repository state
- **Other root files/directories**: Any additional files in ~/.anvil (except settings.yaml)

### Preserved Content

- **settings.yaml**: Your main configuration file with all settings
- **Directory structure**: Essential directories (temp/, archive/) preserved for tool functionality

## Examples

```bash
anvil clean            # Interactive cleanup with confirmation
anvil clean --force    # Force cleanup without confirmation
anvil clean --dry-run  # Preview what would be cleaned
```

## Safety Features

- **Interactive Confirmation**: Shows exactly what will be deleted before proceeding
- **Dry-Run Mode**: Safe preview using identical detection logic as real cleanup
- **Settings Preservation**: settings.yaml is never touched
- **Directory Maintenance**: Essential directories remain empty but ready for use

## Related Documentation

- [Config Command](config.md)
- [Doctor Command](doctor.md)
