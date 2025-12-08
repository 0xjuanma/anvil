# Install Command

The `anvil install` command provides automated installation of development tools and applications using Homebrew on macOS.

## Usage

```bash
anvil install [app-name|group-name] [flags]
```

### Flags

- `--list`: Show available groups and tracked apps
- `--tree`: View applications in hierarchical tree format
- `--dry-run`: Preview installations before execution
- `--group-name`: Add installed app to a specific group(new or existing)

## Installation Modes

### Individual Application

```bash
anvil install firefox
anvil install slack
anvil install visual-studio-code
```

Apps are automatically tracked in `tools.installed_apps` unless already in a group or required_tools.

### With Group Assignment

```bash
anvil install firefox --group-name browsers
anvil install terraform --group-name devops
```

Creates the group if it doesn't exist.

### Group Installation

```bash
anvil install dev          # Development tools
anvil install essentials   # Essential applications
```

## Default Groups

- **dev**: git, zsh, iterm2, visual-studio-code
- **essentials**: slack, google-chrome, 1password

Custom groups can be defined in `~/.anvil/settings.yaml`:

```yaml
groups:
  frontend: [git, node, visual-studio-code, figma]
  devops: [docker, kubectl, terraform]
```

## Source-Based Installation

For apps not in Homebrew, configure custom sources in settings.yaml:

```yaml
sources:
  moom: https://manytricks.com/download/moom
  oh-my-zsh: 'sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"'
```

Supported formats: URLs (.dmg, .pkg, .zip, .deb, .rpm, .AppImage) and shell commands.

## App Detection

Anvil uses intelligent detection to identify already-installed applications:
- Homebrew package check
- Installed cask detection
- Dynamic cask search
- /Applications directory search
- System-wide Spotlight search
- PATH-based detection for CLI tools

## Related Documentation

- [Init Command](init.md)
- [Config Command](config.md)
