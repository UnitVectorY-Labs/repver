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
repver --command=<command_name> [--param.<name>=<value> ...] [--debug] [--dryRun]
```

## Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `--command=<command_name>` | The command to execute (as defined in your .repver file) | Yes |
| `--param.<name>=<value>` | Values for the named parameters (matching regex capture groups) | Yes (if defined by the command) |
| `--debug` | Enable detailed debug output | No |
| `--dryRun` | Show what would be changed without modifying files or performing git operations | No |

## Parameters

Parameters provided via the `--param` flag must correspond to the named capture groups in your regex patterns. For example, if your regex includes `(?P<version>.*)`, you supply:

```bash
--param.version=1.2.3
```

Each named capture group you define in your regex patterns will result in a required parameter.

## Dry Run Mode

When you use the `--dryRun` flag, the tool will:

1. Display what files would be modified, showing the specific line numbers with current and updated content
2. Skip all git operations (creating branches, committing, pushing, etc.)
3. Print information about the git operations that would have been performed

This is useful for verifying what changes would be made before actually applying them.
