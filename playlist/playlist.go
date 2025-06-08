package playlist

import (
	"fmt"
	"os"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/h2non/filetype"
	log "github.com/sirupsen/logrus"
)

var ErrUnsupportedFormat = fmt.Errorf("unsupported format")

// Playlist represents a collection of songs with functionality to queue songs.
type Playlist struct {
	Head     *Song
	size     int
	SongChan chan *Song
}

// Queue adds a new song to the end of the playlist.
func (p *Playlist) Queue(songs ...*Song) {
	for _, s := range songs {
		if p.Head == nil {
			p.Head = s
		} else {
			current := p.Head
			for current.Next != nil {
				current = current.Next
			}
			current.Next = s
			s.Prev = current
		}
		p.size++
	}
}

// Size returns the number of songs in the playlist.
func (p *Playlist) Size() int {
	return p.size
}

// ListSongs returns a slice of all songs in the playlist, in the order they appear.
func (p *Playlist) ListSongs() []Song {
	songs := []Song{}
	current := p.Head
	for current != nil {
		songs = append(songs, *current)
		current = current.Next
	}

	return songs
}

// NewPlaylist creates a new Playlist object and starts a goroutine that
// controls the lifetime of the playlist.
func NewPlaylist() *Playlist {
	p := &Playlist{}
	p.SongChan = make(chan *Song)

	return p
}

// getFileType takes a string representation of a path to a file and returns its extension.
func getFileType(path string) string {
	f, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	kind, err := filetype.Match(f)
	if err != nil {
		log.Fatal(err)
	}

	return kind.Extension
}

// streamerForType returns a StreamSeekCloser, Format, and error for the given file type.
// The returned StreamSeekCloser is used to read audio data from the file.
func streamerForType(fileType string, file *os.File) (beep.StreamSeekCloser, beep.Format, error) {
	switch fileType {
	case "mp3":
		return mp3.Decode(file)
	case "flac":
		return flac.Decode(file)
	default:
		return nil, beep.Format{}, ErrUnsupportedFormat
	}
}
