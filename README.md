[![License](https://img.shields.io/badge/license-MIT-blue)](https://opensource.org/licenses/MIT) [![Concept](https://img.shields.io/badge/Status-Concept-white)](https://guide.unitvectorylabs.com/bestpractices/status/#concept)

# repver

Automate repetitive project updates and Git operations when updating simple strings like version numbers across multiple files concurrently.

## Why repver?

Sometimes, you need to update multiple files in a project with the same string, like a version number. This can be tedious and error-prone if done manually.

- Create a new branch
- Update the first file with the new version number, located in one place in the file
- Update the next file with the new version number, located in a different place in that file
- Commit the changes
- Push the branch
- Create a pull request

But wait, isn't Dependabot already doing this for you? Yes, but it will only work for the dependencies that it manages. If you have version numbers in places like documentation, Dockerfiles, or other files, you’ll need to update those manually.

This is where `repver` comes in. It automates updating multiple files with the same string—creating a new branch, committing the changes, and pushing them to your remote repository.

## Configuration

The `repver` command is run from the commandline inside of a git repository. It relies on a `.repver` file in the root of the repository containing a YAML configuration that defines the desired actions. You can define multiple commands, each of which can operate on multiple files.

Let's take a look at an example configuration file, the one used by `repver` itself to manage its own version of Go that it uses:

```yaml
commands:
 - name: "goversion"
   targets:
   - path: "go.mod"
     pattern: "^go (?P<version>.*) // GOVERSION$"
   - path: ".github/workflows/build-go.yml"
     pattern: "^          go-version: '(?P<version>.*)' # GOVERSION$"

```

The `path` attribute will specify the file to update within the repository.

The `pattern` attribute specifies a regex pattern that is used to identify a single line in a file. It must contain at least one capture group, and all capture groups must be named. These capture group names can then be specified as values in the command to substitute these values in the file.

## Command

To run `repver`, you must provide the `--command={name}` argument to specify which command to run. The command name must match one of the command names defined in the `.repver` file. Additionally, the pattern uses a regular expression to match individual lines in the file. The capture groups must be named, and their names are used as parameters in the format `--param.{name}={value}`.

For example, to run the `goversion` command shown above, use:

```bash
repver --command=goversion --param.version=1.24.1
```
