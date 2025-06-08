package env

import (
	"fmt"
	"os"
)

var ErrInvalidPath = fmt.Errorf("invalid path specified")

func GetHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return homeDir, nil
}

// pathExists returns true if the file at the given path exists and false otherwise.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}
