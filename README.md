<div align="center">
  <img src="assets/anvil-2.0.png" alt="Anvil Logo" width="200" style="border-radius: 50%;">
  <h1>Anvil CLI</h1>
</div>

<div align="center">

[![Go Version](https://img.shields.io/badge/go-1.17+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/0xjuanma/anvil)](https://goreportcard.com/report/github.com/0xjuanma/anvil)
[![GitHub Release](https://img.shields.io/github/v/release/0xjuanma/anvil?style=flat&label=Release)](https://github.com/0xjuanma/anvil/releases/latest)
[![Platform](https://img.shields.io/badge/platform-macOS%20only-blue.svg)](#installation)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](#)


Save hours in your process — install the tools you need, sync your configs, and keep your environment consistent with a single command-line tool.
</div>

<div align="center">
  <img src="assets/anvil.gif" alt="Anvil Demo" width="600">
</div>

## What Anvil Does

- **🚀 Batch App Installation** - Install development tools in groups or individually via Homebrew
- **🔄 Configuration Sync** - Sync dotfiles across machines using simple commands and private GitHub repositories  
- **🩺 Health Checks** - Auto-diagnose and fix common setup issues

## Why Choose Anvil?
- **Fast Setup** - Get coding in minutes, not hours
- **Consistency** - Same configs and tools across all machines
- **Built-in Safety** - Dry-run mode, private repo enforcement and automatic backups

## Available Commands

| Command | Description |
|---------|-------------|
| `anvil init` | Initialize your Anvil environment & dependencies |
| `anvil doctor` | Check system health |
| `anvil install [group-name]` | Install tools by groups|
| `anvil config show [app-name]` | Show your anvil settings or app settings |
| `anvil config push [app-name]` | Push your app configurations to GitHub |
| `anvil config pull [app-name]` | Pull your app configurations from GitHub |
| `anvil config sync [app-name]` | Sync your pulled app configurations to your local machine |
| `anvil clean` | Clean your anvil environment |
| `anvil update` | Update your anvil installation |
| `anvil version` | Show the version of anvil |


## Quick Start

### Installation

**New installations:**
```bash
curl -sSL https://github.com/0xjuanma/anvil/releases/latest/download/install.sh | bash
```

**Update existing installation:**
```bash
anvil update
```


### Try It Out

```bash
# Initialize Anvil
anvil init

# Check environment health
anvil doctor

# Install development tools
anvil install essentials # sample essentials group
anvil install terraform  # Individual apps

# Import tool groups from shared configs
anvil config import https://example.com/team-groups.yaml

# Or start with example configurations
anvil config import https://raw.githubusercontent.com/0xjuanma/anvil/master/docs/import-examples/juanma-essentials.yaml

# Sync configurations (after setting up GitHub repo)
anvil config push neovim
anvil config pull neovim
anvil config sync neovim
```

## Key Features

- **Smart Installation** - Install individual apps or predefined groups(`dev`, `essentials`, etc) holding many apps
- **Group Import** - Import groups from local files or URLs with validation and conflict detection
- **Auto-tracking** - Automatically tracks installed apps and prevents duplicates
- **Secure Config Sync** - Uses private GitHub repositories with automatic backups
- **Health Diagnostics** - `anvil doctor` detects and auto-fixes common issues
- **Zero Configuration** - Works out of the box with sensible defaults

## Documentation

| Guide | Description |
|-------|-------------|
| **[Configuration Management](docs/config.md)** | Config sync setup and workflows |
| **[Install Command](docs/install.md)** | Tool installation guide |
| **[Import Groups](docs/import.md)** | Import tool groups from files/URLs |
| **[Doctor Command](docs/doctor.md)** | Health checks and validation |

**[View All Documentation →](docs/)**

---

<div align="center">

One CLI to rule them all.

**Author:** [@0xjuanma](https://github.com/0xjuanma)  
**[⭐ Star this project](https://github.com/0xjuanma/anvil)**

</div>
