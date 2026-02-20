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
// The extractedGroups map contains named groups extracted from param patterns for use in transforms.
// It returns true if the content was modified, false otherwise.
// In dry run mode, it outputs the changes that would be made without modifying the file.
// It returns an error if any issues occur during file reading, regex compilation, or writing.
func (t *RepverTarget) Execute(values map[string]string, extractedGroups map[string]string) (bool, error) {
	Debugln("Processing file %s using pattern: %s", t.Path, t.Pattern)

	// Read the file content
	Debugln("Reading file: %s", t.Path)
	content, err := os.ReadFile(t.Path)
	if err != nil {
		Debugln("Failed to read file: %v", err)
		return false, err
	}
	Debugln("Read %d bytes from file", len(content))

	// Compile the regex pattern
	Debugln("Compiling pattern: %s", t.Pattern)
	re, err := regexp.Compile(t.Pattern)
	if err != nil {
		Debugln("Invalid regex pattern: %v", err)
		return false, err
	}

	// Get the named capture groups
	names := re.SubexpNames()
	Debugln("Found %d pattern groups (including full match)", len(names))
	for i, name := range names {
		if i > 0 && name != "" {
			Debugln("Group %d: %s", i, name)
		}
	}

	// Prepare effective values: apply transforms if specified
	effectiveValues := make(map[string]string)
	for k, v := range values {
		effectiveValues[k] = v
	}

	// If transform is specified, apply it to get the replacement value
	if t.Transform != "" && extractedGroups != nil {
		transformedValue := ApplyTransform(t.Transform, extractedGroups)
		Debugln("Transform applied: '%s' -> '%s'", t.Transform, transformedValue)

		// The transformed value replaces all named capture groups in this target
		// since transform produces a single output value for the entire replacement
		for i, name := range names {
			if i > 0 && name != "" {
				effectiveValues[name] = transformedValue
			}
		}
	}

	// Process the file line by line
	Debugln("Processing file contents")
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

		if re.MatchString(line) {
			matchesFound++
			Debugln("Found match on line %d", lineNum)

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
					replacement, exists := effectiveValues[name]
					if !exists {
						Debugln("Missing replacement value for group '%s'", name)
						return false, fmt.Errorf("no replacement value for named group '%s'", name)
					}

					// Find indices of this specific capture group in the modified line
					// We need to recompute matches after each replacement as indices might change
					reTemp := regexp.MustCompile(t.Pattern)
					tempMatches := reTemp.FindStringSubmatchIndex(modifiedLine)
					if len(tempMatches) > 2*i+1 {
						start, end := tempMatches[2*i], tempMatches[2*i+1]
						capturedText := modifiedLine[start:end]
						Debugln("Replacing '%s' with '%s' in group '%s'",
							capturedText, replacement, name)

						// Replace just this capture group
						modifiedLine = modifiedLine[:start] + replacement + modifiedLine[end:]
					}
				}

				Debugln("Updated line: '%s'", modifiedLine)

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
				Debugln("No capture groups found in match")
				modifiedLines = append(modifiedLines, line)
			}
		} else {
			// No match, keep the line as is
			modifiedLines = append(modifiedLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		Debugln("Error processing file: %v", err)
		return false, err
	}

	if matchesFound == 0 {
		Debugln("No matches found in file")
	} else {
		Debugln("Found %d matches in file", matchesFound)
	}

	// Join lines back into content
	modifiedContent := strings.Join(modifiedLines, "\n")

	// Check if we need to add a final newline
	if len(content) > 0 && content[len(content)-1] == '\n' {
		modifiedContent += "\n"
	}

	// Check if content was modified
	if string(content) == modifiedContent {
		Debugln("No changes were made to the file content")
		return false, nil
	} else {
		Debugln("File content was modified")
	}

	// We always print the changes, this is the point of repver so we always want to see what is being changed
	if len(changes) > 0 {
		fmt.Println("\nFILE CHANGES:")
		fmt.Printf("  File: %s\n", t.Path)

		for _, change := range changes {
			fmt.Printf("  +- Line %d:\n", change.lineNumber)
			fmt.Printf("  |  - %s\n", change.oldLine)
			fmt.Printf("  |  + %s\n", change.newLine)
		}
		fmt.Println("  +-")
	} else {
		fmt.Printf("\nFile: %s (no changes)\n", t.Path)
		return false, nil
	}

	// If in dry run mode, skip writing the file
	if DryRun {
		Debugln("Dry run mode enabled, skipping file write")
		return contentModified, nil
	}

	// Write the modified content back to the file
	Debugln("Writing changes to %s", t.Path)
	err = os.WriteFile(t.Path, []byte(modifiedContent), 0644)
	if err != nil {
		Debugln("Failed to write file: %v", err)
		return false, err
	}
	Debugln("Successfully updated file")

	return contentModified, nil
}
