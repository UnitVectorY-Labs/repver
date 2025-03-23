package repver

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Execute applies the values map to the file at the path using the regex pattern
// values is a map where keys are named capture group names and values are the replacement values
// If a named capture group in the regex doesn't have a corresponding value in the map,
// the original text will be preserved
func (t *RepverTarget) Execute(values map[string]string) error {
	Debugln("Execute: Starting execution for path: %s with pattern: %s", t.Path, t.Pattern)

	// Read the file content
	Debugln("Execute: Reading file: %s", t.Path)
	content, err := os.ReadFile(t.Path)
	if err != nil {
		Debugln("Execute: ERROR reading file: %v", err)
		return err
	}
	Debugln("Execute: Successfully read %d bytes from file", len(content))

	// Compile the regex pattern
	Debugln("Execute: Compiling regex pattern: %s", t.Pattern)
	re, err := regexp.Compile(t.Pattern)
	if err != nil {
		Debugln("Execute: ERROR compiling regex: %v", err)
		return err
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
						return fmt.Errorf("no replacement value for named group '%s'", name)
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
		return err
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
	} else {
		Debugln("Execute: File content was modified")
	}

	// Write the modified content back to the file
	Debugln("Execute: Writing %d bytes back to file: %s", len(modifiedContent), t.Path)
	err = os.WriteFile(t.Path, []byte(modifiedContent), 0644)
	if err != nil {
		Debugln("Execute: ERROR writing file: %v", err)
		return err
	}
	Debugln("Execute: Successfully wrote file")

	Debugln("Execute: Completed execution for path: %s", t.Path)
	return nil
}
