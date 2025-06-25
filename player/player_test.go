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

func Test_mapVolumeToScale(t *testing.T) {
	tests := []struct {
		name string
		vol  float64
		want float64
	}{
		{"min volume", -12, 0},
		{"negative volume", -10, 4},
		{"zero volume", 0, 24},
		{"positive volume", 6, 36},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapVolumeToScale(tt.vol); got != tt.want {
				t.Errorf("mapVolumeToScale(%v) = %v; want %v", tt.vol, got, tt.want)
			}
		})
	}
}

func Test_mapScaleToVolume(t *testing.T) {
	tests := []struct {
		name  string
		scale float64
		want  float64
	}{
		{"min scale", 0, -12},
		{"positive scale", 24, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapScaleToVolume(tt.scale); got != tt.want {
				t.Errorf("mapScaleToVolume(%v) = %v; want %v", tt.scale, got, tt.want)
			}
		})
	}
}
