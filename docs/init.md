# Init Command

The `anvil init` command bootstraps your Anvil CLI environment. This is the first command you should run after installing Anvil.

## Usage

```bash
anvil init [flags]
```

### Flags

- `--discover`: Automatically discover installed apps and add them to a "discovered-apps" group
- `--skip-tools`: Skip tool validation and installation

## What It Does

- **Tool Validation**: Validates and installs required system tools (Git, cURL, Homebrew)
- **Directory Creation**: Creates `~/.anvil/` directory with `settings.yaml` and `temp/` folder
- **Configuration Generation**: Generates default `settings.yaml` with tool preferences and groups
- **Environment Detection**: Detects OS, existing tools, Git configuration, and Homebrew status
- **App Discovery** (with `--discover`): Scans for installed Homebrew packages and macOS apps, adding untracked apps to a `discovered-apps` group
- **Recommendations**: Provides personalized setup recommendations based on system state

## Next Steps

After running `anvil init`:

```bash
anvil install dev          # Install development tools
anvil install essentials   # Install essential applications
anvil config show          # View your configuration
```

## Related Documentation

- [Install Command](install.md)
- [Config Command](config.md)
- [Doctor Command](doctor.md)
