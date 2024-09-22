package kql

import (
	"fmt"
	"os"
)

func ParseQuery(relativePath string) (kqlQuery string, err error) {
	kqlQuery, err = readKqlFile(relativePath)
	if err != nil {
		return "", err
	}

	return kqlQuery, nil
}

func readKqlFile(relativePath string) (string, error) {
	// Open the file
	file, err := os.ReadFile(relativePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(file), nil
}

func validateQuery(kqlQuery string) error {
	if kqlQuery == "" {
		return fmt.Errorf("kql query cannot be empty")
	}
	return nil
}