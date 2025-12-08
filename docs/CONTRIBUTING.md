# Contributing to Anvil CLI

## Ways to Contribute

- **Bug Reports**: Help identify and fix issues
- **Feature Requests**: Suggest new functionality
- **Documentation**: Improve guides and examples
- **Code Contributions**: Fix bugs, implement features
- **Testing**: Add test cases and improve coverage

## Getting Started

1. Check existing issues to avoid duplicate work
2. Start small with documentation or minor bug fixes
3. Ask questions if anything is unclear

## Development Setup

### Prerequisites

- Go 1.17+
- Git

### Setup

```bash
git clone https://github.com/yourusername/anvil.git
cd anvil
git remote add upstream https://github.com/0xjuanma/anvil.git
go build -o anvil main.go
```

## Contributing Guidelines

### Bug Reports

Include clear description, steps to reproduce, expected vs actual behavior, and environment details (OS, Go version, Anvil version).

### Feature Requests

Include clear description, use case, and proposed solution.

### Pull Requests

Before submitting: search existing PRs, create an issue for significant changes, write tests for new functionality, update documentation, and test thoroughly.

## Development Workflow

1. Create feature branch: `git checkout -b feature/your-feature-name`
2. Make changes following code style guidelines
3. Test changes: `go test ./...`
4. Commit with conventional messages: `git commit -m "feat: add new feature"`
5. Submit pull request

## Code Style

Follow standard Go conventions and use existing code patterns in the project.
