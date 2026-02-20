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

## Updating Go Version with Parameter Validation

This example demonstrates using `params` to validate the version format and `transform` to handle the Go version replacement. The version is validated as a semantic version, and the full version is used in all files.

### Configuration

Create a `.repver` file in your repository root:

```yaml
commands:
  - name: "goversion"
    params:
    - name: "version"
      pattern: "^(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)$"
    targets:
    - path: "go.mod"
      pattern: "^go (?P<version>.*) // GOVERSION$"
      transform: "{{major}}.{{minor}}"
    - path: ".github/workflows/build.yml"
      pattern: "^          go-version: '(?P<version>.*)' # GOVERSION$"
      transform: "{{major}}.{{minor}}.{{patch}}"
    git:
      create_branch: true
      branch_name: "repver/go-v{{version}}"
      commit: true
      commit_message: "Update Go version to {{version}}"
      push: true
      remote: "origin"
      pull_request: "GITHUB_CLI"
      return_to_original_branch: true
      delete_branch: true
```

### Running the Command

Update Go to version 1.26.0 and create a PR:

```bash
repver --command=goversion --param-version=1.26.0
```

The `params` section validates that the version follows semantic versioning format. If you provide an invalid version like `1.26` or `latest`, repver will report an error before making any changes.

### Preview Changes First

Use `--dry-run` to see what would be changed without making modifications:

```bash
repver --command=goversion --param-version=1.26.0 --dry-run
```

This displays the files that would be modified and the Git operations that would be performed.

## Basic Version Update Without Parameter Validation

For simpler use cases where you don't need parameter validation, you can omit the `params` section:

### Configuration

```yaml
commands:
  - name: "appversion"
    targets:
    - path: "version.txt"
      pattern: "^version: (?P<version>.*)$"
    - path: "package.json"
      pattern: "^  \"version\": \"(?P<version>.*)\",?$"
```

### Running the Command

```bash
repver --command=appversion --param-version=2.0.0
```

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
  repver --command=goversion --param-version=1.26.0
else
  echo "Repository does not support goversion command"
fi
```

### Multi-Repository Script

Loop over multiple repositories and apply updates only where the command is defined:

```bash
#!/usr/bin/env bash
set -euo pipefail

GO_VERSION="1.26.0"

for repo in */; do
  [ -d "$repo/.git" ] || continue

  echo "==> $repo"
  (
    cd "$repo"

    # check for repver + goversion support
    if repver --command=goversion --exists >/dev/null 2>&1; then
      repver --command=goversion --param-version="$GO_VERSION"
    else
      echo "    skipping (no .repver goversion)"
    fi
  )
done
```
