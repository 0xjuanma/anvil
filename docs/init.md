# Init Command

The `anvil init` command bootstraps your Anvil CLI environment by performing a complete initialization process. This is the first command you should run after installing Anvil.

## Overview

The init command establishes a solid foundation for all other Anvil commands by:

- **Validating and installing required system tools** (Git, cURL, Homebrew)
- **Creating necessary configuration directory** (`~/.anvil`)
- **Generating a default settings.yaml** configuration file with your system preferences
- **Checking your local environment** for common development configurations
- **Providing actionable recommendations** for completing your setup

## Usage

```bash
anvil init [flags]
```

### Flags

- `--discover` - Automatically discover installed apps and add them to a "discovered-apps" group
- `--skip-tools` - Skip tool validation and installation

## What It Does

### Stage 1: Tool Validation and Installation

The init command validates and installs required system tools:

**Required Tools:**

- **Git** - Essential for version control and configuration synchronization
- **cURL** - Required for downloading resources and API interactions
- **Homebrew** (macOS only) - Package manager installed automatically if missing

**Process:**

1. Checks if each tool is available in the system PATH
2. Attempts to install missing required tools using appropriate package managers
3. Provides clear feedback on installation success or failure
4. Continues with initialization even if some optional tools are unavailable

### Stage 2: Directory Structure Creation

Creates the necessary directory structure for Anvil:

```
~/.anvil/
├── settings.yaml    # Main configuration file
└── temp/           # Temporary directory for pulled configs
```

**Directory Purpose:**

- `~/.anvil/` - Main Anvil configuration directory
- `settings.yaml` - Stores your tool preferences, groups, and GitHub configuration
- `temp/` - Temporary storage for configuration files pulled from repositories

### Stage 3: Configuration File Generation

Generates a default `settings.yaml` configuration file with:

```yaml
tools:
  required_tools: [git, curl, brew]
  installed_apps: [] # Auto-populated by anvil install [app-name]

groups:
  dev: [git, zsh, iterm2, visual-studio-code]
  essentials: [slack, google-chrome, 1password]

git:
  username: ""
  email: ""

github:
  config_repo: "" # Configure this for config sync
  branch: "main"
  token_env_var: "GITHUB_TOKEN"
```

### Stage 4: Environment Detection

The init command automatically detects and reports:

- **Operating System** - Confirms macOS compatibility
- **Existing Tools** - Lists development tools already installed
- **Git Configuration** - Checks if Git is configured with user name and email
- **Homebrew Status** - Verifies Homebrew installation and functionality

### Stage 5: App Discovery (Optional)

If run with `--discover`, Anvil scans your system for installed applications:

- **Homebrew Discovery** - Finds all installed Homebrew packages
- **macOS App Discovery** - Scans `/Applications` for apps (filtering defaults like Safari)
- **Automatic Grouping** - Adds new, untracked apps to a `discovered-apps` group
- **Safe Execution** - Non-destructive; only adds apps not already in your config

### Stage 6: Recommendations

Based on your system state, init provides personalized recommendations:

- **Git Configuration** - Suggests setting up Git user name and email if missing
- **GitHub Setup** - Recommends configuring GitHub repository for config sync
- **Next Steps** - Provides specific commands to continue your setup

## What Happens

The init command validates tools, creates directories, generates configuration, and provides recommendations for next steps.

## Next Steps

After running `anvil init`, you can:

1. **Install tool groups**:

   ```bash
   anvil install dev          # Development tools
   anvil install essentials   # Essential applications
   ```

2. **Configure GitHub sync** (optional):

   ```bash
   # Edit ~/.anvil/settings.yaml to add your repository
   github:
     config_repo: "username/dotfiles"
   ```

3. **Install individual applications**:

   ```bash
   anvil install firefox
   anvil install slack
   ```

4. **Explore available options**:
   ```bash
   anvil install --list    # See available groups
   anvil config show       # View your configuration
   ```

## Troubleshooting

### Common Issues

**Permission Denied**

```bash
# Make sure you have admin privileges for tool installation
sudo anvil init
```

**Network Issues**

```bash
# Check internet connection for downloading tools
curl -I https://github.com
```

**Homebrew Installation Failed**

```bash
# Manually install Homebrew first
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

**Configuration Directory Issues**

```bash
# Check directory permissions
ls -la ~/.anvil
```

### Getting Help

For detailed setup guidance and examples, see:

- [Installation Guide](INSTALLATION.md)
