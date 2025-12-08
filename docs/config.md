# Configuration Management

Anvil's configuration management system syncs dotfiles and configuration files across machines using GitHub repositories.

## Security Requirement

**Anvil REQUIRES private repositories for configuration management.**

Configuration files contain sensitive data (API keys, tokens, personal paths, system information, authentication data) that must never be exposed publicly. Anvil blocks all pushes to public repositories and verifies privacy before every push.

## Commands

### anvil config pull [directory]

Pull configuration files from a specific directory in your GitHub repository.

```bash
anvil config pull cursor
anvil config pull vscode
```

### anvil config show [directory]

Display configuration files and settings.

```bash
anvil config show              # Show all settings
anvil config show cursor       # Show specific directory
anvil config show --groups     # Show only groups (-g)
anvil config show --configs    # Show only config sources (-c)
anvil config show --git        # Show only git configuration
anvil config show --github     # Show only GitHub configuration
```

### anvil config sync [app-name]

Move pulled configuration files from temp directory to local destinations with automatic archiving.

```bash
anvil config sync
anvil config sync cursor
anvil config sync --dry-run    # Preview changes
```

### anvil config push [app-name]

Push configuration files to your GitHub repository with automated branch creation.

```bash
anvil config push
anvil config push cursor
```

### anvil config import [file-or-url]

Import group definitions from local files or URLs.

```bash
anvil config import ./team-groups.yaml
anvil config import https://example.com/groups.yaml
```

See [Import Groups](import.md) for detailed documentation.

## Setup

1. Run `anvil init`
2. Create a **private** GitHub repository
3. Configure repository in `~/.anvil/settings.yaml`:

```yaml
github:
  config_repo: "username/repo-name"
  branch: "main"
  token_env_var: "GITHUB_TOKEN"
```

4. Set up authentication via GitHub token (`export GITHUB_TOKEN="your_token"`) or SSH keys

## Related Documentation

- [Import Groups](import.md)
- [Doctor Command](doctor.md)
- [Clean Command](clean.md)
