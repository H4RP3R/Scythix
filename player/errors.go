package player

import "fmt"

var (
	ErrNoFilePath        = fmt.Errorf("file not specified")
	ErrInvalidPath       = fmt.Errorf("invalid path specified")
	ErrUnsupportedFormat = fmt.Errorf("unsupported format")
	ErrFailedToFork      = fmt.Errorf("failed to fork process")
)
