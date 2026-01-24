---
layout: default
title: Command
nav_order: 2
permalink: /command
---

# Command Line Reference

The `repver` tool automates file updates and Git operations using commands defined in your configuration file.

To use the command, you must run it from the root of your repository where the `.repver` file is located.  If you configure Git operations, this is required to be the root of the Git repository as well.

## Usage

```bash
repver --command=<command_name> [--param-<name>=<value> ...] [--debug] [--dry-run] [--exists]
```

## Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `--command=<command_name>` | The command to execute (as defined in your .repver file) | Yes |
| `--param-<name>=<value>` | Values for the named parameters (matching regex capture groups) | Yes (if defined by the command) |
| `--debug` | Enable detailed debug output | No |
| `--dry-run` | Show what would be changed without modifying files or performing git operations | No |
| `--exists` | Check whether .repver exists and contains the specified command; exits 0 if yes, non-zero otherwise | No |

## Parameters

Parameters provided via the `--param` flag must correspond to the named capture groups in your regex patterns. For example, if your regex includes `(?P<version>.*)`, you supply:

```bash
--param-version=1.2.3
```

Each named capture group you define in your regex patterns will result in a required parameter.

## Dry Run Mode

When you use the `--dry-run` flag, the tool will:

1. Display what files would be modified, showing the specific line numbers with current and updated content
2. Skip all git operations (creating branches, committing, pushing, etc.)
3. Print information about the git operations that would have been performed

This is useful for verifying what changes would be made before actually applying them.

## Exists Mode

The `--exists` flag is designed for scripting and CI workflows. It checks whether a repository has a valid `.repver` configuration file and whether the specified command is defined.

When you use the `--exists` flag:

1. The `--command` flag is required
2. All `--param-*` flags are ignored
3. No file modifications are performed
4. No Git operations are performed
5. Output is minimal (errors only, to stderr)

**Exit Codes:**

| Exit Code | Meaning |
|-----------|---------|
| 0 | `.repver` exists, is valid, and contains the specified command |
| 1 | `.repver` is missing, invalid, or the command is not found |

**Example:**

```bash
# Check if a repository supports the 'goversion' command
if repver --command=goversion --exists; then
  echo "Repository supports goversion command"
else
  echo "Repository does not support goversion command"
fi
```

This mode is ideal for scripts that need to conditionally run repver across multiple repositories.
