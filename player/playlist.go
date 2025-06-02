package player

import (
	"os"

	"github.com/gopxl/beep"
	log "github.com/sirupsen/logrus"
)

// Song represents an entry for the playlist.
type Song struct {
	Streamer beep.StreamSeekCloser
	Format   beep.Format
	Prop     *AudioProperties
	Next     *Song
	Prev     *Song
}

func NewSong(songPath string) (*Song, error) {
	var song Song
	f, err := os.Open(songPath)
	if err != nil {
		return nil, err
	}

	song.Prop, err = NewAudioProperties(songPath)
	if err != nil {
		// It is possible to play a song without properties.
		// Error for debugging purposes only.
		log.Debugf("Error while retrieving audio file meta data: %v", err)
	}

	fileType := getFileType(songPath)
	song.Streamer, song.Format, err = streamerForType(fileType, f)
	if err != nil {
		// The song file should only be closed if an error occurs.
		// Streamer will take care of that.
		f.Close()
		return nil, ErrUnsupportedFormat
	}

	return &song, nil
}

// Playlist represents a collection of songs with functionality to queue songs.
//
// TODO:
// * Read playlist from file.
// * Move backward/forward.
type Playlist struct {
	Head     *Song
	size     int
	SongChan chan *Song
}

// Queue adds a new song to the end of the playlist.
func (p *Playlist) Queue(song *Song) {
	if p.Head == nil {
		p.Head = song
	} else {
		current := p.Head
		for current.Next != nil {
			current = current.Next
		}
		current.Next = song
		song.Prev = current
	}
	p.size++
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
