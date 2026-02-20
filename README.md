[![GitHub release](https://img.shields.io/github/release/UnitVectorY-Labs/repver.svg)](https://github.com/UnitVectorY-Labs/repver/releases/latest) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Active](https://img.shields.io/badge/Status-Active-green)](https://guide.unitvectorylabs.com/bestpractices/status/#active) [![codecov](https://codecov.io/gh/UnitVectorY-Labs/repver/graph/badge.svg?token=uW5nCgpegM)](https://codecov.io/gh/UnitVectorY-Labs/repver) [![Go Report Card](https://goreportcard.com/badge/github.com/UnitVectorY-Labs/repver)](https://goreportcard.com/report/github.com/UnitVectorY-Labs/repver)

# repver

Automate batch replacement of simple strings, like version numbers, across multiple files and handle all Git steps with one command.

## Why repver?

Sometimes, you need to update multiple files in a project with the same string, like a version number. This can be tedious and error-prone if done manually.  Using `repver`, short for "replace version", you can automate this process, making it easier to manage and reducing the risk of human error. This includes a flow that could look like:

- Creating a new branch for the change
- Updating multiple files by replacing a string with the updated version (using regex on a line-by-line basis)
- Committing the changes
- Pushing the branch to the remote repository
- Creating a pull request on GitHub
- Locally switching back to the original branch
- Deleting the local branch

While it doesn't take long to manually complete these tasks, if you need to do it across multiple repositories or multiple files, it can take a significant amount of time. The entire process can be automated with a single command to `repver` which will take care of all of those steps for you based on your repository specific configuration in the `.repver` file.

But wait, isn't Dependabot already doing this for you? Yes, but it will only work for the dependencies that it manages. If you have version numbers in places like documentation, Dockerfiles, or other files, you’ll need to update those manually. You can use `repver` to automate all of those updates at once, and even create a pull request for you to review the changes before merging.

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
   params:
   - name: "version"
     pattern: "^(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)$"
   targets:
   - path: "go.mod"
     pattern: "^go (?P<version>.*) // GOVERSION$"
     transform: "{{major}}.{{minor}}"
   - path: ".github/workflows/build-go.yml"
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

## Command

To run `repver`, you must provide the `--command={name}` argument to specify which command to run. The command name must match one of the command names defined in the `.repver` file. This allows you to have different sets of values to update independently with different commands. The pattern uses a regular expression to match individual lines in the file. The capture groups must be named, and their names are used as parameters in the format `--param-{name}={value}`.

For example, to run the `goversion` command shown above, use:

```bash
repver --command=goversion --param-version=1.26.0
```

## Dependencies

- Git commands utilize the git command line commands which must be installed and accessible with appropriate permissions to the repository
- Creating pull requests with GitHub utilizes the `gh` command line tool, which must be installed and authenticated with your GitHub account
