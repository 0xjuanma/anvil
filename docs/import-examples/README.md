# Import Examples

This directory contains example group configurations for different developer personas. Use with `anvil config import` to quickly set up tool groups.

## Available Personas

| Persona | File | Description |
|---------|------|-------------|
| Frontend Developer | `frontend-developer.yaml` | JavaScript/TypeScript, design tools, productivity |
| Backend Developer | `backend-developer.yaml` | Server technologies, databases, DevOps tools |
| Data Scientist | `data-scientist.yaml` | Data analysis, ML tools, visualization |
| DevOps Engineer | `devops-engineer.yaml` | Infrastructure, cloud tools, monitoring |
| Designer | `designer.yaml` | UI/UX design, prototyping applications |
| Startup Founder | `startup-founder.yaml` | Full-stack development, business tools |
| Team Startup | `team-startup.yaml` | Multi-role team configuration |

## Usage

```bash
anvil config import import-examples/frontend-developer.yaml
anvil config import import-examples/backend-developer.yaml
```

## Group Structure

```yaml
groups:
  group-name:
    - tool1
    - tool2
```

- **Group names**: Use kebab-case (e.g., `frontend-core`)
- **Tool names**: Use lowercase with hyphens (e.g., `visual-studio-code`)
- **Size**: Keep groups focused with 3-8 tools each

## Security Note

These files contain only tool group definitions. No sensitive data, API keys, or personal information is included.

## Related Documentation

- [Import Groups](../import.md)
- [Config Command](../config.md)
