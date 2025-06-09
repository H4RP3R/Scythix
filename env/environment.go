// Package env provides utility functions for interacting with the operating
// system environment, such as retrieving the user's home directory and
// checking the existence of file paths.
package env

import (
	"fmt"
	"os"
)

var ErrInvalidPath = fmt.Errorf("invalid path specified")

// GetHomeDir retrieves the user's home directory path.
// It returns an error if the home directory cannot be determined.
func GetHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return homeDir, nil
}

// PathExists checks whether a file or directory exists at the given path.
// It returns true if the path exists, and false otherwise.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
