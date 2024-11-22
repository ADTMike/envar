package envar

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// load reads environment variables from a given .env file and sets them in the current environment.
// If the file doesn't exist, it logs a warning and skips loading.
func load(filePath string, loadedVars map[string]bool, wg *sync.WaitGroup, ce chan error) {
	defer wg.Done()

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

// convertToType converts a string value to the specified field type.
func convertToType(value string, field reflect.Value) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			durationValue, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("failed to convert %s to time.Duration: %v", value, err)
			}
			field.Set(reflect.ValueOf(durationValue))
			break
		}
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to int: %v", value, err)
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to uint: %v", value, err)
		}
		field.SetUint(uintValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to convert %s to bool: %v", value, err)
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("failed to convert %s to float: %v", value, err)
		}
		field.SetFloat(floatValue)
	case reflect.Complex64, reflect.Complex128:
		complexValue, err := strconv.ParseComplex(value, 128)
		if err != nil {
			return fmt.Errorf("failed to convert %s to complex: %v", value, err)
		}
		field.SetComplex(complexValue)
	case reflect.Slice:
		field.Set(reflect.ValueOf(strings.Split(value, ",")))
	default:
		// Handle unsupported types
		return fmt.Errorf("unsupported field type %s for value %s", field.Kind(), value)
	}
	return nil
}

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
		// For now, we return the first error but you can choose to log or aggregate further.
		for _, err := range errors {
			log.Println(err)
		}
		return fmt.Errorf("encountered errors while loading environment variables")
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
				err := convertToType(envValue, field)
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
