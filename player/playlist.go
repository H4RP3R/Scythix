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
	Prop     *audioProperties
	Next     *Song
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
	NextChan chan struct{}
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
	}
	p.size++
}

// Size returns the number of songs in the playlist.
func (p *Playlist) Size() int {
	return p.size
}

// NewPlaylist creates a new Playlist object and starts a goroutine that
// controls the lifetime of the playlist.
func NewPlaylist() *Playlist {
	p := &Playlist{}
	p.SongChan = make(chan *Song)
	p.NextChan = make(chan struct{})

	go func() {
		for range p.NextChan {
			if p.Head == nil {
				close(p.SongChan)
				return
			}
			p.SongChan <- p.Head
			p.Head = p.Head.Next
		}
	}()

	return p
}
