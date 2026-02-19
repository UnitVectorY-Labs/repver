package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/UnitVectorY-Labs/repver/internal/git"
	"github.com/UnitVectorY-Labs/repver/internal/repver"
)

var Version = "dev" // This will be set by the build systems to the release version

// main is the entry point for the repver command-line tool.
func main() {
	// Set the build version from the build info if not set by the build system
	if Version == "dev" || Version == "" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
				Version = bi.Main.Version
			}
		}
	}

	// Pre-parse static flags to handle --version and --exists before loading .repver
	preParse := flag.NewFlagSet("preparse", flag.ContinueOnError)
	preParse.SetOutput(os.Stderr)
	preCommand := preParse.String("command", "", "Command to execute")
	preExists := preParse.Bool("exists", false, "Check whether .repver exists and contains the specified command")
	preVersion := preParse.Bool("version", false, "Print version")
	preDebug := preParse.Bool("debug", false, "Enable debug mode")
	preDryRun := preParse.Bool("dry-run", false, "Dry run mode")

	// Register param-* flags dynamically to avoid unknown flag errors during pre-parse
	// We'll accept any --param-* flags here but not use them
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--param-") {
			parts := strings.SplitN(arg, "=", 2)
			paramName := strings.TrimPrefix(parts[0], "--")
			if preParse.Lookup(paramName) == nil {
				preParse.String(paramName, "", "")
			}
		}
	}

	// Parse pre-parse flags - errors are handled by falling through to normal mode
	_ = preParse.Parse(os.Args[1:])

	// Handle --version early
	if *preVersion {
		fmt.Println("Version:", Version)
		return
	}

	// Handle --exists mode
	if *preExists {
		handleExistsMode(*preCommand)
		return
	}

	// Set debug and dry-run from pre-parse for early debugging
	repver.Debug = *preDebug
	repver.DryRun = *preDryRun

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

	// Process: Validate params and extract named groups
	extractedGroups := make(map[string]string)
	for _, param := range command.Params {
		value, exists := argumentValues[param.Name]
		if exists {
			// Validate the value against the param pattern
			if err := param.ValidateValue(value); err != nil {
				printErrorAndExit(108, fmt.Sprintf("Parameter '%s' validation failed: %v", param.Name, err))
			}

			// Extract named groups from the value
			groups, err := param.ExtractNamedGroups(value)
			if err != nil {
				printErrorAndExit(109, fmt.Sprintf("Failed to extract groups from parameter '%s': %v", param.Name, err))
			}
			for k, v := range groups {
				extractedGroups[k] = v
			}
			repver.Debugln("Extracted groups from param '%s': %v", param.Name, groups)
		}
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
		modified, err := target.Execute(argumentValues, extractedGroups)

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

// handleExistsMode handles the --exists flag behavior.
// It checks if .repver exists and contains the specified command.
// Exits with 0 if successful, 1 otherwise.
func handleExistsMode(command string) {
	// Check if --command is provided
	if command == "" {
		fmt.Fprintln(os.Stderr, "--command is required with --exists")
		os.Exit(1)
	}

	// Check if .repver exists
	if _, err := os.Stat(".repver"); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, ".repver not found")
		os.Exit(1)
	}

	// Load .repver
	config, err := repver.Load(".repver")
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid .repver")
		os.Exit(1)
	}

	// Validate .repver
	if err := config.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, "invalid .repver")
		os.Exit(1)
	}

	// Check if command exists
	_, err = config.GetCommand(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "command not found: %s\n", command)
		os.Exit(1)
	}

	// Success - command exists
	os.Exit(0)
}
