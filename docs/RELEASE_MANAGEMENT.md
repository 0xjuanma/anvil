# Release Management Guide

This guide covers creating releases for Anvil and managing GitHub Actions workflows.

## Creating a New Release

### Standard Release Process

1. Ensure changes are merged to master:

```bash
git checkout master
git pull origin master
```

2. Create and push a version tag:

```bash
git tag v1.2.0
git push origin v1.2.0
```

3. Monitor the GitHub Actions workflow at `https://github.com/0xjuanma/anvil/actions`
4. Verify the release at `https://github.com/0xjuanma/anvil/releases`

### Version Numbering

Follow Semantic Versioning:

- **Major (v2.0.0)**: Breaking changes, major new features
- **Minor (v1.2.0)**: New features, backward compatible
- **Patch (v1.1.1)**: Bug fixes, small improvements

Pre-release versions: `v1.2.0-beta.1`, `v1.2.0-rc.1`, `v1.2.0-alpha.1`

## Managing Release Tags

### Fixing a Failed Release

```bash
# Delete tag locally and remotely
git tag -d v1.1.2
git push origin --delete v1.1.2

# Delete GitHub release via web UI if created

# Fix issue, commit, and recreate tag
git add .
git commit -m "fix: resolve release issue"
git push origin master
git tag v1.1.2
git push origin v1.1.2
```

### Useful Tag Commands

```bash
git tag -l              # List all tags
git tag -l "v1.1.*"     # List tags with pattern
git show v1.1.2         # Show tag details
```

## GitHub Actions Workflow

The release workflow triggers on tag push matching `v*.*.*` pattern or manual trigger.

### What the Workflow Does

- Builds binaries for macOS (Intel/Apple Silicon) and Linux (Intel/ARM)
- Generates security checksums
- Creates GitHub release with binaries and installation script

## Release Checklist

**Pre-release:**
- All changes merged to master
- Version number decided (semver)
- Local testing completed
- Documentation updated

**Release:**
- Create and push tag
- Monitor GitHub Actions
- Verify release created

**Post-release:**
- Test installation methods
- Verify binary downloads
- Update external documentation

## Troubleshooting

- **Workflow doesn't run**: Verify tag follows `v*.*.*` pattern
- **Build failures**: Check Go version (1.17+) and dependencies
- **Permission errors**: Verify GitHub Actions permissions
- **Download issues**: Ensure release is published (not draft)
