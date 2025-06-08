package playlist

import (
	"os"

	"github.com/gopxl/beep"
	log "github.com/sirupsen/logrus"
)

// Song represents an entry for the playlist.
type Song struct {
	Streamer beep.StreamSeekCloser
	Format   beep.Format
	FullPath string
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

	song.FullPath = songPath
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
