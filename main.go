package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/UnitVectorY-Labs/repver/internal/repver"
)

// Main function returns hello world
func main() {

	// Check for a file named ".repver" in the current directory
	// If it doesn't exist exit with an error
	if _, err := os.Stat(".repver"); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "Error: .repver file not found")
		os.Exit(1)
	}

	config, err := repver.Load(".repver")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading configuration:", err)
		os.Exit(1)
	}

	err = config.Validate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error validating configuration:", err)
		os.Exit(1)
	}

	// Add all parameters to the command line using config.GetParameterNames setting them as string flag variables prefixed with val.
	allParameters, err := config.GetParameterNames()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error getting parameter names:", err)
		os.Exit(1)
	}

	// Map of the flag variables
	commandParams := make(map[string]*string)
	for _, command := range allParameters {
		commandParams[command] = flag.String("param."+command, "", "Value for "+command)
	}

	// Parse the params
	repver.ParseParams()

	if repver.Command != "" {
		command, err := config.GetCommand(repver.Command)
		if err != nil {
			// Command not found, print error and exit
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		parameters, err := command.GetParameterNames()
		if err != nil {
			// Error getting parameter names, print error and exit
			fmt.Fprintln(os.Stderr, "Error getting parameter names:", err)
			os.Exit(1)
		}

		// Create a map for named capture group replacements
		values := make(map[string]string)
		for _, param := range parameters {
			// Check if the parameter is set
			if val, ok := commandParams[param]; ok && *val != "" {
				values[param] = *val
			} else {
				// If the parameter is not set, return an error
				fmt.Fprintf(os.Stderr, "Error: Parameter %s is required\n", param)
				os.Exit(1)
			}
		}

		// Get the original branch name
		originalBranch, err := repver.GetCurrentBranch()
		if err != nil {
			// Error getting current branch, print error and exit
			fmt.Fprintln(os.Stderr, "Error getting current branch:", err)
			os.Exit(1)
		}

		// Check if we are switching branches
		branchName := originalBranch
		if command.GitOptions.CreateBranch {
			branchName = command.GitOptions.BuildBranchName(values)
			err = repver.CreateAndSwitchBranch(branchName)
			if err != nil {
				// Error creating and switching branch, print error and exit
				fmt.Fprintln(os.Stderr, "Error creating and switching branch:", err)
				os.Exit(1)
			}
		}

		commitFiles := []string{}

		// Loop through the targets and execute
		for _, target := range command.Targets {
			// execute the command
			err = target.Execute(values)
			if err != nil {
				// Error executing command, print error and exit
				fmt.Fprintln(os.Stderr, "Error executing command:", err)
				os.Exit(1)
			}
			fmt.Printf("Command executed successfully for target: %s\n", target.Path)

			// Add the target path to the commit files
			commitFiles = append(commitFiles, target.Path)
		}

		// Check if we need to commit the changes
		if command.GitOptions.Commit {
			commitMessage := command.GitOptions.BuildCommitMessage(values)
			err = repver.AddAndCommitFiles(commitFiles, commitMessage)
			if err != nil {
				// Error committing changes, print error and exit
				fmt.Fprintln(os.Stderr, "Error committing changes:", err)
				os.Exit(1)
			}
			fmt.Println("Changes committed successfully")

			// Check if we need to push the changes
			if command.GitOptions.Push && branchName != "" {
				remote := command.GitOptions.Remote
				if remote == "" {
					remote = "origin"
				}

				err = repver.PushChanges(remote, branchName)
				if err != nil {
					// Error pushing changes, print error and exit
					fmt.Fprintln(os.Stderr, "Error pushing changes:", err)
					os.Exit(1)
				}
				fmt.Println("Changes pushed successfully")
			}
		}

		// Check if we need to return to the original branch
		if command.GitOptions.ReturnToOriginalBranch {
			err = repver.SwitchBranch(originalBranch)
			if err != nil {
				// Error switching back to original branch, print error and exit
				fmt.Fprintln(os.Stderr, "Error switching back to original branch:", err)
				os.Exit(1)
			}
			fmt.Println("Returned to original branch successfully")

			// Check if we need to delete the branch
			if command.GitOptions.DeleteBranch && command.GitOptions.CreateBranch {
				err = repver.DeleteLocalBranch(branchName)
				if err != nil {
					// Error deleting branch, print error and exit
					fmt.Fprintln(os.Stderr, "Error deleting branch:", err)
					os.Exit(1)
				}
				fmt.Println("Deleted branch successfully")
			}
		}

	} else {
		// If no command is specified, enumerate all commands to the user and explain how to use them
		fmt.Fprintln(os.Stderr, "Available commands:")
		for _, command := range config.Commands {
			fmt.Fprintf(os.Stderr, "  - %s\n", command.Name)
		}
		fmt.Fprintln(os.Stderr, "Usage: repver -command <command_name>")
		os.Exit(1)
	}
}
