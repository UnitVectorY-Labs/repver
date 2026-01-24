---
layout: default
title: Examples
nav_order: 4
permalink: /examples
---

# Examples
{: .no_toc }

## Table of contents
{: .no_toc .text-delta }

1. TOC
{:toc}

## Updating Versions Across Files

The most common use case for repver is updating version numbers or other strings consistently across multiple files.

### Configuration

Create a `.repver` file in your repository root:

```yaml
commands:
  - name: "goversion"
    targets:
    - path: "go.mod"
      pattern: "^go (?P<version>.*) // GOVERSION$"
    - path: ".github/workflows/build.yml"
      pattern: "^          go-version: '(?P<version>.*)' # GOVERSION$"
    git:
      create_branch: true
      branch_name: "go-v{{version}}"
      commit: true
      commit_message: "Update Go version to {{version}}"
      push: true
      remote: "origin"
      pull_request: "GITHUB_CLI"
      return_to_original_branch: true
      delete_branch: true
```

### Running the Command

Update Go to version 1.24.0 and create a PR:

```bash
repver --command=goversion --param-version=1.24.0
```

### Preview Changes First

Use `--dry-run` to see what would be changed without making modifications:

```bash
repver --command=goversion --param-version=1.24.0 --dry-run
```

This displays the files that would be modified and the Git operations that would be performed.

## Checking Command Availability

Use `--exists` to check whether a repository supports a specific repver command. This is useful for scripting across multiple repositories.

### Basic Check

```bash
repver --command=goversion --exists
echo $?  # 0 if command exists, 1 otherwise
```

### Conditional Execution

```bash
if repver --command=goversion --exists; then
  repver --command=goversion --param-version=1.24.0
else
  echo "Repository does not support goversion command"
fi
```

### Multi-Repository Script

Loop over multiple repositories and apply updates only where the command is defined:

```bash
#!/bin/bash
VERSION="1.24.0"

for repo in repo1 repo2 repo3; do
  cd "$repo"
  
  if repver --command=goversion --exists 2>/dev/null; then
    echo "Updating $repo..."
    repver --command=goversion --param-version="$VERSION"
  else
    echo "Skipping $repo (no goversion command)"
  fi
  
  cd ..
done
```

## Unix-Style Scripting

Repver follows Unix conventions for composability:

- **Exit code 0** indicates success
- **Non-zero exit codes** indicate failure
- **Minimal output** in `--exists` mode (errors to stderr only)
- **Works with standard shell constructs** like `if`, `&&`, `||`

### Examples

```bash
# Run only if command exists
repver --command=update --exists && repver --command=update --param-version=2.0

# Handle errors
repver --command=update --param-version=2.0 || echo "Update failed"

# Silent check
if repver --command=update --exists 2>/dev/null; then
  # Command is available
fi
```

## CI/CD Integration

Repver works well in CI pipelines. Use `--dry-run` in PR checks to validate the configuration without making changes:

```yaml
# GitHub Actions example
- name: Validate repver configuration
  run: repver --command=goversion --param-version=1.24.0 --dry-run
```

For automated updates, use the full command with Git automation:

```yaml
- name: Update Go version
  run: repver --command=goversion --param-version=${{ inputs.version }}
```
