package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/UnitVectorY-Labs/repver/internal/git"
	"github.com/UnitVectorY-Labs/repver/internal/repver"
)

var Version = "dev" // This will be set by the build systems to the release version

// main is the entry point for the repver command-line tool.
func main() {

	// Initialization Phase

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
		argumentFlags[argumentName] = flag.String("param-"+argumentName, "", "Value for "+argumentName)
	}

	showVersion := flag.Bool("version", false, "Print version")

	// Process: Parse command line arguments
	repver.ParseParams()

	if *showVersion {
		fmt.Println("Version:", Version)
		return
	}

	// If dry run mode is enabled, output that information
	if repver.DryRun {
		fmt.Println("DRY RUN MODE ENABLED")
	}

	// Decision: Command specified?
	if repver.UserCommand == "" {
		// Generate help message listing all available commands with their parameters
		helpMessage := generateHelpMessage(config)
		printErrorAndExit(103, "No command specified", helpMessage)
	}

	// Process: Retrieve command configuration
	command, err := config.GetCommand(repver.UserCommand)

	// Decision: Command found?
	if err != nil {
		helpMessage := generateHelpMessage(config)
		printErrorAndExit(104, "Command not found", helpMessage)
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
	missingParams := []string{}
	for _, parameter := range parameters {
		// Check if the parameter is set
		if val, ok := argumentFlags[parameter]; ok && *val != "" {
			argumentValues[parameter] = *val
		} else {
			missingParams = append(missingParams, parameter)
		}
	}

	if len(missingParams) > 0 {
		// Create a targeted help message for the specific command
		var helpBuilder strings.Builder
		helpBuilder.WriteString(fmt.Sprintf("Command '%s' requires the following parameters:\n", repver.UserCommand))

		for _, param := range missingParams {
			helpBuilder.WriteString(fmt.Sprintf("  --param-%s=<value>\n", param))
		}

		helpBuilder.WriteString("\nComplete usage example:\n")
		helpBuilder.WriteString(fmt.Sprintf("  repver --command=%s", repver.UserCommand))
		for _, param := range parameters {
			helpBuilder.WriteString(fmt.Sprintf(" --param-%s=<value>", param))
		}

		printErrorAndExit(105, "Missing required parameters", helpBuilder.String())
	}

	// Decision: Git options specified?
	useGit := command.GitOptions.GitOptionsSpecified()
	if useGit && !repver.DryRun {
		// Decision: In git root?
		isGitRoot, err := git.IsGitRoot()
		if err != nil {
			// This error isn't in the flowchart because the failure here is
			printErrorAndExit(503, "Internal error determining git root")
		}
		if !isGitRoot {
			printErrorAndExit(106, "Not in git repository")
		}

		// Decision: Git workspace clean?
		err = git.CheckGitClean()
		if err != nil {
			printErrorAndExit(107, "Git workspace not clean")
		}
	} else if useGit && repver.DryRun {
		fmt.Println("[DRYRUN] Git operations would be performed but are disabled in dry run mode")
	}

	// Execution Phase

	// Decision: Git options specified?
	originalBranchName := ""
	newBranchName := ""
	if useGit && !repver.DryRun {
		// Process: Get the current branch name
		originalBranchName, err = git.GetCurrentBranch()
		if err != nil {
			// This error isn't in the flowchart because we previously checked we are in a git repo
			printErrorAndExit(504, "Internal error could not get current branch name")
		}

		// Decision: Create new branch?
		newBranchName = originalBranchName
		if command.GitOptions.CreateBranch {
			newBranchName = command.GitOptions.BuildBranchName(argumentValues)

			// Decision: Branch already exists?
			branchExists, err := git.BranchExists(newBranchName)
			if err != nil {
				printErrorAndExit(503, "Internal error checking if branch exists")
			}
			if branchExists {
				printErrorAndExit(200, fmt.Sprintf("Branch '%s' already exists", newBranchName))
			}

			// Process: Create new branch
			output, err := git.CreateAndSwitchBranch(newBranchName)
			// Decision: Branch creation successful?
			if err != nil {
				printErrorAndExit(201, "Failed to create new branch")
			}
			repver.Debugln("Created and switched to new branch\n%s", output)
		}
	} else if useGit && repver.DryRun && command.GitOptions.CreateBranch {
		// Process: Get the current branch name
		originalBranchName, err = git.GetCurrentBranch()
		if err != nil {
			// This error isn't in the flowchart because we previously checked we are in a git repo
			printErrorAndExit(504, "Internal error could not get current branch name")
		}

		// In dry run mode, just show what branch would be created
		newBranchName = command.GitOptions.BuildBranchName(argumentValues)
		fmt.Printf("[DRYRUN] Would create and switch to branch: %s\n", newBranchName)
	}

	// Decision: Has targets to update?
	anyFileModified := false
	commitFiles := []string{}

	for _, target := range command.Targets {
		// Process: Execute update to target
		modified, err := target.Execute(argumentValues)

		// Decision: Execution successful?
		if err != nil {
			printErrorAndExit(202, "Failed to execute command on target")
		}

		if modified {
			anyFileModified = true
			commitFiles = append(commitFiles, target.Path)
		}
	}

	// Decision: Commit changes to git?
	if !anyFileModified {
		repver.Debugln("No files modified, skipping commit")
	} else if command.GitOptions.Commit && !repver.DryRun {
		// Process: Construct commit message
		commitMessage := command.GitOptions.BuildCommitMessage(argumentValues)

		// Process: Commit changes to git
		output, err := git.AddAndCommitFiles(commitFiles, commitMessage)
		if err != nil {
			// This error isn't in the flowchart because we previously checked we are in a git repo
			printErrorAndExit(505, "Internal error could not add and commit files")
		}
		repver.Debugln("Changes committed successfully\n%s", output)

		// Decision: Push changes to remote?
		if command.GitOptions.Push && newBranchName != "" {
			remote := command.GitOptions.Remote
			if remote == "" {
				remote = "origin"
			}

			// Process: Push changes to remote
			output, err = git.PushChanges(remote, newBranchName)
			if err != nil {
				// This error isn't in the flowchart because we previously checked we are in a git repo
				printErrorAndExit(506, "Internal error failed to push changes")
			}
			repver.Debugln("Changes pushed successfully\n%s", output)

			// Decision: Create pull request?
			if command.GitOptions.PullRequest == "GITHUB_CLI" {
				output, err = git.CreateGitHubPullRequest()
				if err != nil {
					printErrorAndExit(508, "Failed to create GitHub pull request")
				}
				repver.Debugln("Created GitHub pull request\n%s", output)
			}
		}
	} else if command.GitOptions.Commit && repver.DryRun {
		// In dry run mode, just show what would be committed
		commitMessage := command.GitOptions.BuildCommitMessage(argumentValues)
		fmt.Printf("[DRYRUN] Would commit changes with message: \"%s\"\n", commitMessage)
		fmt.Printf("[DRYRUN] Files that would be added to the commit:\n")
		for _, file := range commitFiles {
			fmt.Printf("  - %s\n", file)
		}

		if command.GitOptions.Push {
			remote := command.GitOptions.Remote
			if remote == "" {
				remote = "origin"
			}
			fmt.Printf("[DRYRUN] Would push changes to remote '%s' branch '%s'\n", remote, newBranchName)
		}

		if command.GitOptions.PullRequest == "GITHUB_CLI" {
			fmt.Println("[DRYRUN] Would create GitHub pull request")
		}
	}

	// Decision: Return to original branch?
	if command.GitOptions.ReturnToOriginalBranch && !repver.DryRun && anyFileModified {
		// Process: Switch back to original branch
		output, err := git.SwitchToBranch(originalBranchName)
		if err != nil {
			// This error isn't in the flowchart because we previously checked we are in a git repo
			printErrorAndExit(507, "Internal error failed to switch back to original branch")
		}
		repver.Debugln("Returned to original branch\n%s", output)

		// Decision: Delete new branch?
		if command.GitOptions.DeleteBranch && command.GitOptions.CreateBranch {

			// Process: Delete new branch
			output, err = git.DeleteLocalBranch(newBranchName)
			if err != nil {
				// This error isn't in the flowchart because we previously checked we are in a git repo
				printErrorAndExit(509, "Internal error failed to delete new branch")
			}
			repver.Debugln("Deleted branch\n%s", output)
		}
	} else if command.GitOptions.ReturnToOriginalBranch && repver.DryRun {
		fmt.Printf("[DRYRUN] Would switch back to original branch '%s'\n", originalBranchName)

		if command.GitOptions.DeleteBranch && command.GitOptions.CreateBranch {
			fmt.Printf("[DRYRUN] Would delete branch '%s'\n", newBranchName)
		}
	}
}

// generateHelpMessage creates a formatted help message showing all available commands
// and their required parameters from the configuration
func generateHelpMessage(config *repver.RepverConfig) string {
	var help strings.Builder

	help.WriteString("USAGE:\n")
	help.WriteString("  repver --command=<command_name> [--param-<n>=<value> ...] [OPTIONS]\n\n")

	help.WriteString("AVAILABLE COMMANDS:\n")

	if len(config.Commands) == 0 {
		help.WriteString("  No commands defined in .repver configuration\n")
		return help.String()
	}

	// Get the longest command name for proper padding
	maxNameLen := 0
	for _, cmd := range config.Commands {
		if len(cmd.Name) > maxNameLen {
			maxNameLen = len(cmd.Name)
		}
	}

	// Sort the commands alphabetically for easier reading
	cmdNames := make([]string, 0, len(config.Commands))
	cmdMap := make(map[string]*repver.RepverCommand)
	for i, cmd := range config.Commands {
		cmdNames = append(cmdNames, cmd.Name)
		cmdMap[cmd.Name] = &config.Commands[i]
	}
	sort.Strings(cmdNames)

	// Print each command with its parameters
	for _, name := range cmdNames {
		cmd := cmdMap[name]

		// Get parameters for this command
		params, err := cmd.GetParameterNames()
		if err != nil {
			continue // Skip if we can't get parameters
		}

		// Format command name with padding
		padding := strings.Repeat(" ", maxNameLen-len(name)+2)
		help.WriteString(fmt.Sprintf("  %s%s", name, padding))

		// Include example usage
		if len(params) > 0 {
			paramList := strings.Join(params, ", ")
			help.WriteString(fmt.Sprintf("Parameters: [%s]\n", paramList))

			// Add complete example
			help.WriteString(fmt.Sprintf("    Example: repver --command=%s", name))
			for _, param := range params {
				help.WriteString(fmt.Sprintf(" --param-%s=<value>", param))
			}
			help.WriteString("\n\n")
		} else {
			help.WriteString("No parameters required\n")
			help.WriteString(fmt.Sprintf("    Example: repver --command=%s\n\n", name))
		}
	}

	help.WriteString("OPTIONS:\n")
	help.WriteString("  --debug    Enable debug output\n")
	help.WriteString("  --dry-run  Show what would be changed without modifying files or performing git operations\n")

	return help.String()
}

func printErrorAndExit(errNum int, errMsg string, helpMsg ...string) {
	fmt.Fprintf(os.Stderr, "Error (%d): %s\n", errNum, errMsg)
	if len(helpMsg) > 0 && helpMsg[0] != "" {
		fmt.Fprintln(os.Stderr, "\n"+helpMsg[0])
	}
	os.Exit(errNum)
}
