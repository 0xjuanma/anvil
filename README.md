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

- **Batch App Installation**: Install development tools in groups or individually via Homebrew
- **Configuration Sync**: Sync dotfiles across machines using simple commands and private GitHub repositories  
- **Health Checks**: Auto-diagnose and fix common setup issues

## Why Choose Anvil?

- **Fast Setup**: Get coding in minutes, not hours
- **Consistency**: Same configs and tools across all machines
- **Built-in Safety**: Dry-run mode, private repo enforcement and automatic backups

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

### Available Commands

<table>
<thead>
<tr>
<th style="padding: 8px 16px; text-align: center;">Command</th>
<th style="padding: 8px 16px; text-align: left;">Description</th>
</tr>
</thead>
<tbody>
<tr>
<td style="padding: 8px 16px; white-space: nowrap; text-align: center;"><code>anvil init [--discover]</code></td>
<td style="padding: 8px 16px; text-align: left;">Initialize your Anvil environment, dependencies & optionally discovers apps in your system</td>
</tr>
<tr>
<td style="padding: 8px 16px; white-space: nowrap; text-align: center;"><code>anvil doctor</code></td>
<td style="padding: 8px 16px; text-align: left;">Check system health</td>
</tr>
<tr>
<td style="padding: 8px 16px; white-space: nowrap; text-align: center;"><code>anvil install [group-name]</code></td>
<td style="padding: 8px 16px; text-align: left;">Install tools by groups</td>
</tr>
<tr>
<td style="padding: 8px 16px; white-space: nowrap; text-align: center;"><code>anvil config show [app-name]</code></td>
<td style="padding: 8px 16px; text-align: left;">Show your anvil settings or app settings</td>
</tr>
<tr>
<td style="padding: 8px 16px; white-space: nowrap; text-align: center;"><code>anvil config push [app-name]</code></td>
<td style="padding: 8px 16px; text-align: left;">Push your app configurations to GitHub</td>
</tr>
<tr>
<td style="padding: 8px 16px; white-space: nowrap; text-align: center;"><code>anvil config pull [app-name]</code></td>
<td style="padding: 8px 16px; text-align: left;">Pull your app configurations from GitHub</td>
</tr>
<tr>
<td style="padding: 8px 16px; white-space: nowrap; text-align: center;"><code>anvil config sync [app-name]</code></td>
<td style="padding: 8px 16px; text-align: left;">Sync your pulled app configurations to your local machine</td>
</tr>
<tr>
<td style="padding: 8px 16px; white-space: nowrap; text-align: center;"><code>anvil clean</code></td>
<td style="padding: 8px 16px; text-align: left;">Clean your anvil environment</td>
</tr>
<tr>
<td style="padding: 8px 16px; white-space: nowrap; text-align: center;"><code>anvil update</code></td>
<td style="padding: 8px 16px; text-align: left;">Update your anvil installation</td>
</tr>
<tr>
<td style="padding: 8px 16px; white-space: nowrap; text-align: center;"><code>anvil version</code></td>
<td style="padding: 8px 16px; text-align: left;">Show the version of anvil</td>
</tr>
</tbody>
</table>


### Try It Out

```bash
# Initialize Anvil (optionally discover existing apps)
anvil init --discover

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

## Features

- **Smart Installation**: Install individual apps or user-defined groups (`dev`, `essentials`, etc) holding many apps
- **Group Import**: Import groups from local files or URLs with validation and conflict detection
- **Auto-tracking**: Automatically tracks installed apps and prevents duplicates
- **Secure Config Sync**: Uses private GitHub repositories with automatic backups
- **Health Diagnostics**: `anvil doctor` detects and auto-fixes common issues
- **Zero Configuration**: Works out of the box with sensible defaults

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
**[Star this project](https://github.com/0xjuanma/anvil)**

</div>
