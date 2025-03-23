package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/UnitVectorY-Labs/repver/internal/repver"
)

// Main function returns hello world
func main() {

	// Decision: .repver exists?
	if _, err := os.Stat(".repver"); os.IsNotExist(err) {
		printErrorAndExit(100, ".repver file not found")
	}

	// Process: Load .repver
	config, err := repver.Load(".repver")

	// Decision: Load successful?
	if err != nil {
		printErrorAndExit(101, ".repver failed to load")
	}

	// Process: Validate .repver
	err = config.Validate()

	// Decision: Validation Successful?
	if err != nil {
		printErrorAndExit(102, fmt.Sprintf(".repver validation failed\n%v", err))
	}

	// Process: Enumerate possible command line arguments from .repver
	argumentNames, err := config.GetParameterNames()
	if err != nil {
		// This error is not on the flowchart because the previous validate step
		// should prevent this from ever happening
		printErrorAndExit(501, "Internal error compiling prevalidated parameters")
	}

	argumentValues := make(map[string]*string)
	for _, argumentName := range argumentNames {
		argumentValues[argumentName] = flag.String("param."+argumentName, "", "Value for "+argumentName)
	}

	// Process: Parse command line arguments
	repver.ParseParams()

	// Decision: Command specified?
	if repver.UserCommand == "" {
		// TODO: Print out additional help
		printErrorAndExit(103, "No command specified")
	}

	// Process: Retrieve command configuration
	command, err := config.GetCommand(repver.UserCommand)
	if err != nil {
		// TODO: Print out additional help
		printErrorAndExit(104, "No command found")
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
		if val, ok := argumentValues[param]; ok && *val != "" {
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
}

func printErrorAndExit(errNum int, errMsg string) {
	fmt.Fprintf(os.Stderr, "Error (%d): %s\n", errNum, errMsg)
	os.Exit(errNum)
}
