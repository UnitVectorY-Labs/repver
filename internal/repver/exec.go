package repver

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Execute performs the regex replacement on the file specified by Path.
// It reads the file, applies the regex pattern, and writes back the modified content.
// The values map contains the replacement values for named capture groups in the regex pattern.
// It returns true if the content was modified, false otherwise.
// In dry run mode, it outputs the changes that would be made without modifying the file.
// It returns an error if any issues occur during file reading, regex compilation, or writing.
func (t *RepverTarget) Execute(values map[string]string) (bool, error) {
	Debugln("Execute: Starting execution for path: %s with pattern: %s", t.Path, t.Pattern)

	// Read the file content
	Debugln("Execute: Reading file: %s", t.Path)
	content, err := os.ReadFile(t.Path)
	if err != nil {
		Debugln("Execute: ERROR reading file: %v", err)
		return false, err
	}
	Debugln("Execute: Successfully read %d bytes from file", len(content))

	// Compile the regex pattern
	Debugln("Execute: Compiling regex pattern: %s", t.Pattern)
	re, err := regexp.Compile(t.Pattern)
	if err != nil {
		Debugln("Execute: ERROR compiling regex: %v", err)
		return false, err
	}
	Debugln("Execute: Successfully compiled regex pattern")

	// Get the named capture groups
	names := re.SubexpNames()
	Debugln("Execute: Found %d subexpressions (including full match)", len(names))
	for i, name := range names {
		if i > 0 && name != "" {
			Debugln("Execute: Named group %d: %s", i, name)
		}
	}

	// Process the file line by line
	Debugln("Execute: Processing file line by line")
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var modifiedLines []string
	lineNum := 0
	matchesFound := 0
	contentModified := false

	// Track changes for dry run mode
	var changes []struct {
		lineNumber int
		oldLine    string
		newLine    string
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		Debugln("Execute: Processing line %d: %s", lineNum, line)

		if re.MatchString(line) {
			matchesFound++
			Debugln("Execute: Match found on line %d", lineNum)

			// Process the line with named groups
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 { // At least one capture group
				// Start with the original line
				modifiedLine := line

				// Process each named capture group
				for i, name := range names {
					if i == 0 || name == "" {
						continue // Skip the full match and unnamed groups
					}

					// Check if we have a replacement value for this named group
					replacement, exists := values[name]
					if !exists {
						Debugln("Execute: No replacement value for named group '%s', keeping original text", name)
						// return an error
						return false, fmt.Errorf("no replacement value for named group '%s'", name)
					}

					// Find indices of this specific capture group in the modified line
					// We need to recompute matches after each replacement as indices might change
					reTemp := regexp.MustCompile(t.Pattern)
					tempMatches := reTemp.FindStringSubmatchIndex(modifiedLine)
					if len(tempMatches) > 2*i+1 {
						start, end := tempMatches[2*i], tempMatches[2*i+1]
						capturedText := modifiedLine[start:end]
						Debugln("Execute: Replacing named group '%s': '%s' with '%s'",
							name, capturedText, replacement)

						// Replace just this capture group
						modifiedLine = modifiedLine[:start] + replacement + modifiedLine[end:]
					}
				}

				Debugln("Execute: Line after replacements: '%s'", modifiedLine)

				// If line was changed, record it for dry run mode
				if line != modifiedLine {
					contentModified = true
					changes = append(changes, struct {
						lineNumber int
						oldLine    string
						newLine    string
					}{lineNum, line, modifiedLine})
				}

				modifiedLines = append(modifiedLines, modifiedLine)
			} else {
				// If no capture groups, keep the line as is
				Debugln("Execute: No capture groups found, keeping line unchanged")
				modifiedLines = append(modifiedLines, line)
			}
		} else {
			// No match, keep the line as is
			modifiedLines = append(modifiedLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		Debugln("Execute: ERROR scanning file: %v", err)
		return false, err
	}

	if matchesFound == 0 {
		Debugln("Execute: No matches found in any line of the file")
	} else {
		Debugln("Execute: Found %d matches across the file", matchesFound)
	}

	// Join lines back into content
	modifiedContent := strings.Join(modifiedLines, "\n")

	// Check if we need to add a final newline
	if len(content) > 0 && content[len(content)-1] == '\n' {
		modifiedContent += "\n"
	}

	// Check if content was modified
	if string(content) == modifiedContent {
		Debugln("Execute: No changes were made to the file content")
		return false, nil
	} else {
		Debugln("Execute: File content was modified")
	}

	// In dry run mode, output the changes that would be made
	if DryRun {
		if len(changes) > 0 {
			fmt.Printf("File: %s\n", t.Path)
			for _, change := range changes {
				fmt.Printf("  Line %d:\n", change.lineNumber)
				fmt.Printf("    - %s\n", change.oldLine)
				fmt.Printf("    + %s\n", change.newLine)
			}
			return contentModified, nil
		} else {
			fmt.Printf("File: %s (no changes)\n", t.Path)
			return false, nil
		}
	}

	// Write the modified content back to the file
	Debugln("Execute: Writing %d bytes back to file: %s", len(modifiedContent), t.Path)
	err = os.WriteFile(t.Path, []byte(modifiedContent), 0644)
	if err != nil {
		Debugln("Execute: ERROR writing file: %v", err)
		return false, err
	}
	Debugln("Execute: Successfully wrote file")

	Debugln("Execute: Completed execution for path: %s", t.Path)
	return contentModified, nil
}
