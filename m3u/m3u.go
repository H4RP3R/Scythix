// Package m3u provides functionality to read and write M3U playlist files.
// It only supports the basic M3U directives needed by the Scythix project.
package m3u

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"scythix/env"
	"scythix/playlist"
)

var (
	ErrMissingHeader       = fmt.Errorf("missing m3u file header")
	ErrUnableParsePlaylist = fmt.Errorf("M3U file is empty or malformed")
)

// Save writes the contents of the playlist to a file in the M3U format at the specified path.
func Save(playlist *playlist.Playlist, path string) error {
	sb := strings.Builder{}
	sb.WriteString("#EXTM3U\n")
	for _, song := range playlist.ListSongs() {
		sb.WriteString("#EXTINF:," + song.Prop.Title + "\n")
		sb.WriteString(song.FullPath + "\n")
	}

	err := os.WriteFile(path, []byte(sb.String()), 0644)
	if err != nil {
		return err
	}

	return nil
}

// Load reads an M3U file from the given path and returns a slice of Song pointers.
// It returns an error if any occurs.
func Load(path string) ([]*playlist.Song, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	if len(lines) < 2 {
		return nil, ErrUnableParsePlaylist
	}

	if lines[0] != "#EXTM3U" {
		return nil, ErrMissingHeader
	}

	songs := []*playlist.Song{}
	for i := 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "#EXTINF:") {
			i++
			if i >= len(lines) {
				break
			}
			if env.PathExists(lines[i]) {
				song, err := playlist.NewSong(lines[i])
				if err != nil {
					log.Errorf("Failed to load song: %v", err)
				}
				songs = append(songs, song)
			}
		}
	}
	return songs, nil
}
