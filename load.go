package envar

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

// load reads environment variables from a given .env file and sets them in the current environment.
// It also resolves any variables that are placeholders, and substitutes them with their corresponding values.
func load(filePath string, loadedVars map[string]bool, wg *sync.WaitGroup, ce chan error) {
	if wg != nil {
		defer wg.Done()
	}

	// Check if the file exists before proceeding
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// add a error but don't stop execution
		ce <- fmt.Errorf("Error .env file not found: %s", filePath)
		return
	}

	// Open the .env file for reading
	file, err := os.Open(filePath)
	if err != nil {
		ce <- fmt.Errorf("Error opening .env file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	// Read each line of the file and set the corresponding environment variable
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines or comments (lines starting with #)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		// Split the line into KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Expand any environment variables in the value
		expandedValue, err := expandVariables(value)
		if err != nil {
			ce <- fmt.Errorf("Error expanding variables in %s: %v", key, err)
			return
		}

		// Set the environment variable if not already set
		if !loadedVars[key] {
			// Log the environment variable being set (for debugging purposes)
			err := os.Setenv(key, expandedValue)
			if err != nil {
				ce <- fmt.Errorf("Error setting environment variable %s: %v", key, err)
				return
			}
			loadedVars[key] = true // Mark the variable as loaded
		}
	}

	// Check for errors in scanning the file
	if err := scanner.Err(); err != nil {
		ce <- fmt.Errorf("Error reading .env file %s: %v", filePath, err)
		return
	}
}

// expandVariables processes a string and replaces placeholders in the form ${VAR_NAME} with the corresponding environment variable value.
func expandVariables(value string) (string, error) {
	// Define a regular expression to match environment variable patterns like ${VAR_NAME}
	re := regexp.MustCompile(`\${([^}]+)}`)

	// Function to replace the placeholder with the actual value of the environment variable
	replacer := func(match string) string {
		// Extract the variable name from the match (removing ${ and })
		varName := match[2 : len(match)-1]

		// Get the value of the environment variable
		varValue := os.Getenv(varName)

		// If the variable is not set, return the original placeholder
		if varValue == "" {
			return match
		}

		return varValue
	}

	// Replace all placeholders in the value
	expandedValue := re.ReplaceAllStringFunc(value, replacer)

	// After one round of replacements, we need to check if any placeholders remain.
	// If any are found, do another round of expansion until no more changes are made.
	for {
		expandedValueNew := re.ReplaceAllStringFunc(expandedValue, replacer)
		if expandedValueNew == expandedValue {
			break
		}
		expandedValue = expandedValueNew
	}

	return expandedValue, nil
}
