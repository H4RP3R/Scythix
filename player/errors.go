package player

import "fmt"

var (
	ErrNoFilePath   = fmt.Errorf("file not specified")
	ErrFailedToFork = fmt.Errorf("failed to fork process")
)
