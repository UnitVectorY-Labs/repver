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
repver --command=<command_name> [--param.<name>=<value> ...] [--debug]
```

## Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `--command=<command_name>` | The command to execute (as defined in your .repver file) | Yes |
| `--param.<name>=<value>` | Values for the named parameters (matching regex capture groups) | Yes (if defined by the command) |
| `--debug` | Enable detailed debug output | No |

## Parameters

Parameters provided via the `--param` flag must correspond to the named capture groups in your regex patterns. For example, if your regex includes `(?P<version>.*)`, you supply:

```bash
--param.version=1.2.3
```

Each named capture group you define in your regex patterns will result in a required parameter.
