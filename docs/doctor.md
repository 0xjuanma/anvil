# Doctor Command

The `anvil doctor` command provides health checks to validate your development environment and troubleshoot configuration issues. This command also provides automatic ways of fixing the issues found.

## Usage

```bash
anvil doctor [category|check] [flags]
```

### Flags

- `--list`: Show all available categories and checks
- `--verbose`: Show detailed descriptions and step-by-step results
- `--fix`: Auto-fix fixable issues

### Categories

- **environment**: Verify anvil initialization and directory structure (3 checks)
- **dependencies**: Check required tools and Homebrew installation (2 checks)
- **configuration**: Validate git and GitHub settings (3 checks)
- **connectivity**: Test GitHub access and repository connections (3 checks)

### Examples

```bash
anvil doctor                    # Run all health checks
anvil doctor --list             # List available categories and checks
anvil doctor environment        # Run all checks in a category
anvil doctor git-config         # Run a specific individual check
anvil doctor --fix              # Auto-fix all fixable issues
anvil doctor dependencies --fix # Auto-fix issues in a specific category
```

## Health Checks

### Environment Checks

| Check | Description | Auto-Fix |
|-------|-------------|----------|
| `anvil-init` | Verify anvil initialization completed | No |
| `settings-valid` | Validate settings.yaml structure | No |
| `directory-structure` | Check ~/.anvil directory structure | No |

### Dependencies Checks

| Check | Description | Auto-Fix |
|-------|-------------|----------|
| `homebrew` | Verify Homebrew installation | Yes |
| `required-tools` | Check git and curl are installed | No |

### Configuration Checks

| Check | Description | Auto-Fix |
|-------|-------------|----------|
| `git-config` | Validate git user.name and user.email | Yes |
| `github-config` | Verify GitHub repository configuration | No |
| `sync-config` | Check config sync settings | No |

### Connectivity Checks

| Check | Description | Auto-Fix |
|-------|-------------|----------|
| `github-auth` | Test GitHub authentication | No |
| `github-repo` | Verify repository accessibility | No |
| `git-operations` | Test git clone and pull operations | No |

## Check Results

- **PASS**: Check completed successfully
- **WARN**: Check passed with warnings or recommendations
- **FAIL**: Check failed and requires attention
- **SKIP**: Check was skipped (usually due to missing configuration)

## Related Documentation

- [Init Command](init.md)
- [Config Command](config.md)
