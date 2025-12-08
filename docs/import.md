# Import Groups

The `anvil config import` command imports tool group definitions from local files or remote URLs into your anvil configuration.

## Usage

```bash
anvil config import [file-or-url]
```

### Examples

```bash
# Import from local file
anvil config import ./team-groups.yaml

# Import from remote URL
anvil config import https://raw.githubusercontent.com/company/configs/main/groups.yaml

# Import from example configurations
anvil config import import-examples/frontend-developer.yaml
anvil config import import-examples/backend-developer.yaml
```

## Features

- **Flexible Sources**: Import from local files or publicly accessible URLs
- **Comprehensive Validation**: Validates group names, application names, and structure
- **Conflict Detection**: Prevents overwriting existing groups
- **Tree Display**: Shows visual preview of groups before import
- **Interactive Confirmation**: Requires user approval before making changes
- **Security-First**: Only imports group definitions, ignoring sensitive data

## File Format

Import files must be valid YAML with a `groups` section:

```yaml
groups:
  group-name:
    - tool1
    - tool2
  another-group:
    - tool3
    - tool4
```

## Available Example Configurations

| Persona | File | Description |
|---------|------|-------------|
| Frontend Developer | `import-examples/frontend-developer.yaml` | Modern web development tools |
| Backend Developer | `import-examples/backend-developer.yaml` | Server-side technologies |
| Data Scientist | `import-examples/data-scientist.yaml` | Data analysis and ML tools |
| DevOps Engineer | `import-examples/devops-engineer.yaml` | Infrastructure and deployment |
| Designer | `import-examples/designer.yaml` | UI/UX design tools |
| Startup Founder | `import-examples/startup-founder.yaml` | Technical founder setup |
| Team Startup | `import-examples/team-startup.yaml` | Multi-role team configuration |

## Import Process

1. **File Fetching**: Validates file existence or downloads from URL (30s timeout)
2. **Parsing and Validation**: Validates YAML syntax and extracts groups section
3. **Conflict Detection**: Checks for conflicts with existing groups
4. **Preview and Confirmation**: Displays tree structure and requires approval
5. **Import Execution**: Adds new groups and persists to settings.yaml

## Security

- Only imports group names and tool names
- No API keys, tokens, personal information, or file paths are imported
- Remote imports should use HTTPS URLs
- Temporary files are securely cleaned up

## Related Documentation

- [Config Command](config.md)
- [Install Command](install.md)
- [Import Examples](import-examples/README.md)
