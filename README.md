[![GitHub release](https://img.shields.io/github/release/UnitVectorY-Labs/repver.svg)](https://github.com/UnitVectorY-Labs/repver/releases/latest) [![License](https://img.shields.io/badge/license-MIT-blue)](https://opensource.org/licenses/MIT) [![Active](https://img.shields.io/badge/Status-Active-green)](https://guide.unitvectorylabs.com/bestpractices/status/#active) [![codecov](https://codecov.io/gh/UnitVectorY-Labs/repver/graph/badge.svg?token=uW5nCgpegM)](https://codecov.io/gh/UnitVectorY-Labs/repver) [![Go Report Card](https://goreportcard.com/badge/github.com/UnitVectorY-Labs/repver)](https://goreportcard.com/report/github.com/UnitVectorY-Labs/repver)

# repver

Automate batch replacement of simple strings, like version numbers, across multiple files and handle all Git steps with one command.

## Why repver?

Sometimes, you need to update multiple files in a project with the same string, like a version number. This can be tedious and error-prone if done manually.  Using `repver`, short for "replace version", you can automate this process, making it easier to manage and reducing the risk of human error. This includes a flow that could look like:

- Creating a new branch for the change
- Updating multiple files by replacing a string with the updated version (using regex on a line-by-line basis)
- Committing the changes
- Pushing the branch to the remote repository
- Creating a pull request on GitHub
- Switching back to the original branch
- Deleting the local branch

While it doesn't take long to manually complete these tasks, if you need to do it across multiple repositories or multiple files, it can become tedious. The entire process can be automated with a single command to `repver` which will take care of all of those steps for you based on your pre-configured `.repver` file included in your repository.

But wait, isn't Dependabot already doing this for you? Yes, but it will only work for the dependencies that it manages. If you have version numbers in places like documentation, Dockerfiles, or other files, you’ll need to update those manually.

## Releases

All official versions of **repver** are published on [GitHub Releases](https://github.com/UnitVectorY-Labs/repver/releases). Since this application is written in Go, each release provides pre-compiled executables for macOS, Linux, and Windows—ready to download and run.

Alternatively, if you have Go installed, you can install **repver** directly from source using the following command:

```bash
go install github.com/UnitVectorY-Labs/repver@latest
```

## Configuration

The `repver` command is run from the command line inside of a Git repository. It relies on a `.repver` file in the root of the repository containing a YAML configuration that defines the desired actions. You can define multiple commands, each of which can operate on multiple files.

Let's take a look at an example configuration file, the one used by `repver` itself to manage its own version of Go that it uses:

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

The `path` attribute will specify the file to update within the repository.

The `pattern` attribute specifies a regex pattern that is used to identify a single line in a file. It must contain at least one capture group, and all capture groups must be named. These capture group names can then be specified as values in the command to substitute these values in the file.

The `git` configuration allows for the application to run local Git commands to further automate the process.  This includes:
  - Creating a new branch with `create_branch` whose name can be specified with `branch_name`.
  - Automatically committing the changes with `commit` and a commit message specified with `commit_message`.
  - Pushing the branch to the remote repository with `push` and specifying the remote with `remote`.
  - Returning to the original branch with `return_to_original_branch`.
  - Deleting the branch with `delete_branch` after the changes have been pushed.

The `branch_name` and `commit_message` attributes can use the capture group names from the `pattern` attribute to create a dynamic branch name and commit message. For example, if the capture group name is `version`, you can use `{{version}}` in the branch name or commit message to substitute the value of that capture group.

## Command

To run `repver`, you must provide the `--command={name}` argument to specify which command to run. The command name must match one of the command names defined in the `.repver` file. Additionally, the pattern uses a regular expression to match individual lines in the file. The capture groups must be named, and their names are used as parameters in the format `--param-{name}={value}`.

For example, to run the `goversion` command shown above, use:

```bash
repver --command=goversion --param-version=1.24.1
```

## Dependencies

- Git commands utilize the git command line commands which must be installed and accessible with appropriate permissions to the repository
- Creating pull requests with GitHub utilizes the `gh` command line tool, which must be installed and authenticated with your GitHub account
