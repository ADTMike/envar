package envar

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

// load reads environment variables from a given .env file and sets them in the current environment.
// If the file doesn't exist, it logs a warning and skips loading.
func load(filePath string, loadedVars map[string]bool, wg *sync.WaitGroup, ce chan error) {
	if wg != nil {
		defer wg.Done()
	}

	// Check if the file exists before proceeding
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Log a warning but don't stop execution
		log.Printf("Warning: .env file not found: %s", filePath)
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

		// Set the environment variable if not already set
		if !loadedVars[key] {
			// Log the environment variable being set (for debugging purposes)
			err := os.Setenv(key, value)
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
