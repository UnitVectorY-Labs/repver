---
layout: default
title: repver
nav_order: 1
permalink: /
---

# repver

Automate project updates and Git operations by updating simple strings—for example, version numbers—across many files at once.

## Example Usage

To use `repver` you need to create a configuration in your repository that defines which files you want updated.  An example of this is in the repver repository itself with the [.repver](https://github.com/UnitVectorY-Labs/repver/blob/main/.repver) which uses the following configuration to update the version of Go across multiple files.

```yaml
# .repver
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

The command to upgrade the version which creates a pull request for a new version is:

```bash
repver --command=goversion --param-version=1.24.3
```
