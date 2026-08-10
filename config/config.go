package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// LoadEnv loads environment variables from a .env file if it exists
func LoadEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		// If the file doesn't exist, skip it silently
		return
	}
	defer file.Close()

	log.Println("Loading configuration from local .env file...")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split by the first '=' character
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			// Strip surrounding quotes if present
			val = strings.Trim(val, `"'`)

			// Only set the variable if it doesn't already exist in the system shell
			if os.Getenv(key) == "" && val != "" {
				os.Setenv(key, val)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading .env file:", err)
	}
}
