package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadEnv loads environment variables from .env file if it exists
func LoadEnv() error {
	file, err := os.Open(".env")
	if err != nil {
		// .env file doesn't exist, that's ok
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		// Set environment variable
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set env var %s: %v", key, err)
		}
	}

	return scanner.Err()
}
