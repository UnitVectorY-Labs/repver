package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/UnitVectorY-Labs/repver/internal/repver"
)

// Main function returns hello world
func main() {

	// Initilization Phase

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

	argumentFlags := make(map[string]*string)
	for _, argumentName := range argumentNames {
		argumentFlags[argumentName] = flag.String("param."+argumentName, "", "Value for "+argumentName)
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

	// Decision: Command found?
	if err != nil {
		// TODO: Print out additional help
		printErrorAndExit(104, "No command found")
	}

	// Process: Identify required arguments for command]
	parameters, err := command.GetParameterNames()
	if err != nil {
		// This error is not on the flowchart because the previous validate step
		// should prevent this from ever happening
		printErrorAndExit(502, "Internal error compiling prevalidated parameters")
	}

	// Decision: All params provided?
	argumentValues := make(map[string]string)
	for _, parameter := range parameters {
		// Check if the parameter is set
		if val, ok := argumentFlags[parameter]; ok && *val != "" {
			argumentValues[parameter] = *val
		} else {
			// TODO: Print out additional help listing the missing parameters
			printErrorAndExit(105, "Missing required parameters")
		}
	}

	// Decision: Git options specified?
	useGit := command.GitOptions.GitOptionsSpecified()
	if useGit {
		// Decision: In git root?
		isGitRoot, err := repver.IsGitRoot()
		if err != nil {
			// This error isn't in the flowchart because the failure here is
			printErrorAndExit(503, "Internal error determining git root")
		}
		if !isGitRoot {
			printErrorAndExit(106, "Not in git repository")
		}

		// Decision: Git workspace clean?
		err = repver.CheckGitClean()
		if err != nil {
			printErrorAndExit(107, "Git workspace not clean")
		}
	}

	// Execution Phase

	// Decision: Git options specified?
	originalBranchName := ""
	newBranchName := ""
	if useGit {
		// Process: Get the current branch name
		originalBranchName, err := repver.GetCurrentBranch()
		if err != nil {
			// This error isn't in the flowchart because we previously checked we are in a git repo
			printErrorAndExit(504, "Internal error could not get current branch name")
		}

		// Decision: Create new branch?
		newBranchName := originalBranchName
		if command.GitOptions.CreateBranch {
			newBranchName = command.GitOptions.BuildBranchName(argumentValues)
			err = repver.CreateAndSwitchBranch(newBranchName)
			// Decision: Branch creation successful?
			if err != nil {
				printErrorAndExit(200, "Failed to create new branch")
			}
		}
	}

	commitFiles := []string{}

	// Decision: Has target to update?
	for _, target := range command.Targets {

		// Process: Execute update to target
		err = target.Execute(argumentValues)

		// Decision: Execution successful?
		if err != nil {
			printErrorAndExit(201, "Failed to execute command on target")
		}

		repver.Debugln("Command executed successfully for target: %s", target.Path)
		commitFiles = append(commitFiles, target.Path)
	}

	// Decision: Commit changes to git?
	if command.GitOptions.Commit {
		// Process: Construct commit message
		commitMessage := command.GitOptions.BuildCommitMessage(argumentValues)

		// Process: Commit changes to git
		err = repver.AddAndCommitFiles(commitFiles, commitMessage)
		if err != nil {
			// This error isn't in the flowchart because we previously checked we are in a git repo
			printErrorAndExit(505, "Internal error could not add and commit files")
		}

		repver.Debugln("Changes committed successfully")

		// Decision: Push changes to remote?
		if command.GitOptions.Push && newBranchName != "" {
			remote := command.GitOptions.Remote
			if remote == "" {
				remote = "origin"
			}

			// Process: Push changes to remote
			err = repver.PushChanges(remote, newBranchName)
			if err != nil {
				// This error isn't in the flowchart because we previously checked we are in a git repo
				printErrorAndExit(506, "Internal error failed to push changes")
			}

			repver.Debugln("Changes pushed successfully")
		}
	}

	// Decision: Return to original branch?
	if command.GitOptions.ReturnToOriginalBranch {
		// Process: Switch back to original branch
		err = repver.SwitchBranch(originalBranchName)
		if err != nil {
			// This error isn't in the flowchart because we previously checked we are in a git repo
			printErrorAndExit(507, "Internal error failed to switch back to original branch")
		}

		repver.Debugln("Returned to original branch successfully")

		// Decision: Delete new branch?
		if command.GitOptions.DeleteBranch && command.GitOptions.CreateBranch {

			// Process: Delete new branch
			err = repver.DeleteLocalBranch(newBranchName)
			if err != nil {
				// This error isn't in the flowchart because we previously checked we are in a git repo
				printErrorAndExit(508, "Internal error failed to delete new branch")
			}

			repver.Debugln("Deleted branch successfully")
		}
	}
}

func printErrorAndExit(errNum int, errMsg string) {
	fmt.Fprintf(os.Stderr, "Error (%d): %s\n", errNum, errMsg)
	os.Exit(errNum)
}
