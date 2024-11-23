package envar

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"
)

// Bind loads environment variables from .env files and binds them to the fields of the provided struct.
// The struct fields must be tagged with `env` tags that correspond to the names of the environment variables.
// If no file paths are provided, it defaults to looking for a .env file in the current working directory or a custom directory.
func Bind(i interface{}, rPath ...string) error {
	// Validate the input struct (ensure it's a pointer to a struct)
	if reflect.TypeOf(i).Kind() != reflect.Ptr || reflect.ValueOf(i).Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct, got %v", reflect.TypeOf(i))
	}

	// Initialize a map to track loaded environment variables to avoid duplicates
	loadedVars := make(map[string]bool)
	var wg sync.WaitGroup
	Ce := make(chan error, 10) // Buffered channel to hold multiple errors
	var errors []error

	// If no rPath is provided, load the default .env file from the current directory or the specified directory
	if len(rPath) == 0 {
		// the current working directory
		defaultDir, _ := os.Getwd()

		// Set the path to the default .env file in the specified directory
		defaultEnvPath := filepath.Join(defaultDir, ".env")
		wg.Add(1)
		go load(defaultEnvPath, loadedVars, &wg, Ce)
	}

	// If file paths are provided, load the corresponding .env files concurrently
	for _, path := range rPath {
		wg.Add(1)
		go load(filepath.Join(path, ".env"), loadedVars, &wg, Ce)
	}

	// Wait for all goroutines to finish loading the .env files
	wg.Wait()

	// Close the error channel and collect errors
	close(Ce)
	for err := range Ce {
		errors = append(errors, err)
	}

	// If any critical errors were found, return them
	if len(errors) > 0 {

		for _, err := range errors {
			log.Println(err)
		}
		// Return the last error encountered
		return errors[len(errors)-1]
	}

	// Reflect on the struct to get the fields and their tags
	v := reflect.ValueOf(i).Elem()
	t := v.Type()

	// Iterate over the fields of the struct and set the values from environment variables
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag.Get("env")

		// If an env tag exists, get the corresponding environment variable value
		if envValue := os.Getenv(tag); envValue != "" {
			// Set the field value if the environment variable exists
			if field.CanSet() {
				err := convert(envValue, field)
				if err != nil {
					log.Printf("Warning: Could not convert env variable %s to field type: %v", tag, err)
				}
			} else {
				log.Printf("Warning: Field %s cannot be set (maybe it's unexported)", tag)
			}
		} else {
			log.Printf("Warning: Environment variable %s not found", tag)
		}
	}

	return nil
}
