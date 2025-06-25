package player

import (
	"os"
	"testing"
)

func Test_normalizePath(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to set up test: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "absolute path",
			path:    "//tmp/",
			want:    "/tmp",
			wantErr: false,
		},
		{
			name:    "relative path",
			path:    ".",
			want:    currentDir,
			wantErr: false,
		},
		{
			name:    "nonexistent path",
			path:    "/path/that/does/not/exist",
			want:    "/path/that/does/not/exist",
			wantErr: false,
		},
		{
			name:    "path with redundant separators and parent dir",
			path:    "//tmp//../tmp",
			want:    "/tmp",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("normalizePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
