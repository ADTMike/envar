# Envar - A Go Package for Loading and Binding Environment Variables

**Envar** is a Go package designed to load environment variables from `.env` files and bind them to the fields of a struct. It supports various data types like strings, integers, booleans, floats, and more. This makes managing application configuration in a flexible and environment-specific way very easy.

## Features

- Load environment variables from `.env` files.
- Bind environment variables to struct fields.
- Support for a wide range of types: `string`, `int`, `uint`, `bool`, `float`,`slice`, `time.Duration`.
- Supports multiple `.env` files, loaded concurrently.
- Prevents overwriting of already loaded environment variables.
- Logs warnings when environment variable issues occur.

## Installation

To install the `envar` package, use `go get`:

```bash
go get github.com/ADTMike/envar
```

## Structure for the example
```bash
├── main.go              # Main entry point for the application
├── .env                 # Environment variable configuration file
├── test                 # Directory for custom environment variable files
│   ├── .env             # Custom environment variables for testing
├── test2
│   ├── sub              # Another custom environment variable file
|   |   ├── .env         # Custom environment variables for testing
│   ├── sub2             # Another custom environment variable file
|   |   ├── .env         # Custom environment variables for testing
```
## Content

```.env
# .env configuration for the application

# Database URL (string)
DATABASE_URL=

# Port (integer)
PORT=

# Debug flag (boolean)
DEBUG=

# Timeout duration (time.Duration in Go) ex 30s
TIMEOUT=

# API keys (comma-separated list for slice)
API_KEYS=key1,key2,key3

```
## Example
```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ADTMike/envar"
)

type Config struct {
	DatabaseURL string        `env:"DATABASE_URL"`
	Port        int           `env:"PORT"`
	Debug       bool          `env:"DEBUG"`
	Timeout     time.Duration `env:"TIMEOUT"`
	APIKeys     []string      `env:"API_KEYS"`
}

func main() {
	var config Config

	if err := envar.Bind(&config); err != nil {
		log.Fatalf("Error binding environment variables: %v", err)
	}
	fmt.Printf("Config .env: %+v\n", config)

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	if err := envar.Bind(&config, currentDir+"/test"); err != nil {
		log.Fatalf("Error binding environment variables: %v", err)
	}
	fmt.Printf("Config /test : %+v\n", config)
	if err := envar.Bind(&config, currentDir+"/test2/sub", currentDir+"/test2/sub2"); err != nil {
		log.Fatalf("Error binding environment variables: %v", err)
	}
	fmt.Printf("Config: /test2/sub /test2/sub2 %+v\n", config)

}

```
