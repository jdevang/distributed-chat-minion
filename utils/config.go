package utils

import (
	"bufio"
	"os"
	"strings"
)

func ReadConfigFile(filename string) (map[string]string, error) {
	config := make(map[string]string)

	// Open the file for reading
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}

		// Split the line into key and value components
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}

		// Trim whitespace from key and value components
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Store the key-value pair in the config map
		config[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}
