---
layout: default
title: Configuration
nav_order: 3
permalink: /configuration
---

# Configuration

The `repver` tool uses a configuration file to define commands and their parameters. This file is typically named `.repver` and should be located in the root of your repository. This file is in YAML format and contains `commands` that are invoked when running the `repver` command with the `--command=<name>` argument.

Each command is defined by a `name` which must be unique. It is an alphanumeric string that can be up to 30 characters long.

A command can define multiple `targets`, each specifying a `path` to the file to modify and a `pattern` that defines the regex.  The `pattern` regex must start with "^" and end with "$" to ensure that it matches an entire line in the file as it is applied line-by-line.  The pattern must also contain at least one capture group, and all capture groups must be named using the syntax "(?P<name>.*)" as an example.  The names of the capture groups are used as parameters in the command, to substitute these values in the file using the `--param-<name>=<value>` argument.

The optional `git` section of the command allows for the application to run local Git commands to further automate the process.  This includes:

- Creating a new branch with `create_branch` (boolean) whose name must then be specified with `branch_name`. 
- Automatically committing the changes with `commit` (boolean) and a commit message must then be specified with `commit_message`.
- Pushing the branch to the remote repository with `push` (boolean) and specifying the remote with `remote` (optional, default to 'origin').
- A pull request can be created automatically if `pull_request` (enum: `NO`, `GITHUB_CLI`) enum is set.  The `GITHUB_CLI` option will create a pull request using the GitHub `gh` CLI command. This requires `create_branch` and `push` to be set to true.
- Returning to the original branch with `return_to_original_branch`. This can only be set to true if `create_branch` is set to true.
- Deleting the branch with `delete_branch` after the changes have been pushed. This can only be set to true if `return_to_original_branch` is also set to true.

The `branch_name` and `commit_message` attributes can use the capture group names from the `pattern` attribute to create a dynamic branch name and commit message. For example, if the capture group name is `version`, you can use `{{version}}` in the branch name or commit message to substitute the value of that capture group.

## Example Configuration

```yaml
commands:
 - name: "goversion"
   targets:
   - path: "go.mod"
     pattern: "^go (?P<version>.*) // GOVERSION$"
   - path: ".github/workflows/build-go.yml"
     pattern: "^          go-version: '(?P<version>.*)' # GOVERSION$"
   git:
     create_branch: true
     branch_name : "go-v{{version}}"
     commit: true
     commit_message: "Update Go version to {{version}}"
     push: true
     remote: "origin"
     pull_request: "GITHUB_CLI"
     return_to_original_branch: true
     delete_branch: true
```
