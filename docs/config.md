# Configuration Management

Anvil's configuration management system syncs dotfiles and configuration files across machines using GitHub repositories.

## Security Requirement

> **Note:** Anvil *requires* private repositories for configuration management.

Configuration files contain sensitive data (API keys, tokens, personal paths, system information, authentication data) that must never be exposed publicly. Anvil blocks all pushes to public repositories and verifies privacy before every push.

## Commands

### anvil config show [app-name]

Display configuration files and settings. when no [app-name] is provided, all commands execute in the context of Anvil CLI

```bash
anvil config show              # Show all Anvil settings
anvil config show cursor       # Show specific app configs directory
anvil config show --groups     # Show only Anvil groups (-g)
anvil config show --configs    # Show only Anvil config sources (-c)
anvil config show --sources    # Show only Anvil Installation Sources (-s)
anvil config show --git        # Show only Anvil git configuration
anvil config show --github     # Show only Anvil GitHub configuration
```

### anvil config pull [app-name]

Pull configuration files from a specific directory in your GitHub repository.

```bash
anvil config pull cursor
anvil config pull vscode
```

### anvil config push [app-name]

Push configuration files to your GitHub repository with automated branch creation.

```bash
anvil config push
anvil config push cursor
```

### anvil config sync [app-name]

Move pulled configuration files from temp directory to local destinations with automatic archiving.

```bash
anvil config sync
anvil config sync cursor
anvil config sync --dry-run    # Preview changes
```

### anvil config import [file-or-url]

Import group definitions from local files or URLs. This command allows to share/import existing Anvil groups and automatically updates local Anvil settings.

```bash
anvil config import ./team-groups.yaml
anvil config import https://example.com/groups.yaml
```

See [Import Groups](import.md) for detailed documentation.

## Related Documentation

- [Import Groups](import.md)
- [Doctor Command](doctor.md)
- [Clean Command](clean.md)
