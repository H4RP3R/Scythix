package player

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"
	log "github.com/sirupsen/logrus"

	"scythix/env"
	"scythix/m3u"
	"scythix/playlist"
)

// PlayerServer represents a server for managing music playback via RPC.
type PlayerServer struct {
	PID         int
	playlist    *playlist.Playlist
	currentSong *playlist.Song
	playlistDir string

	ctrl *beep.Ctrl
	vol  *effects.Volume

	stopOnce sync.Once
	done     chan struct{}
}

// Pause toggle the player's paused state.
func (p *PlayerServer) Pause(args *struct{}, reply *struct{}) error {
	speaker.Lock()
	p.ctrl.Paused = !p.ctrl.Paused
	speaker.Unlock()

	log.Debug("Player paused.")

	return nil
}

// Stop halts playback and signals the daemon to finish by closing the `done` channel.
func (p *PlayerServer) Stop(args *struct{}, reply *struct{}) error {
	p.stopOnce.Do(func() {
		close(p.done)
	})

	log.Debug("Got stop command")

	return nil
}

// Mute toggles the mute state of the player.
func (p *PlayerServer) Mute(args *struct{}, reply *struct{}) error {
	speaker.Lock()
	p.vol.Silent = !p.vol.Silent
	speaker.Unlock()

	if p.vol.Silent == true {
		log.Debug("Player muted.")
	} else {
		log.Debug("Player unmuted.")
	}

	return nil
}

// TurnUp increments the player's volume by a predefined step, up to a maximum limit.
func (p *PlayerServer) TurnUp(args *struct{}, reply *float64) error {
	speaker.Lock()
	p.vol.Silent = false
	if p.vol.Volume < volLimitMax {
		p.vol.Volume += volStep
	}
	speaker.Unlock()

	*reply = mapVolumeToScale(p.vol.Volume)
	log.Debugf("Volume set to %g", mapVolumeToScale(p.vol.Volume))

	return nil
}

// TurnDown decreases the player's volume by a predefined step, not going below the minimum limit.
func (p *PlayerServer) TurnDown(args *struct{}, reply *float64) error {
	speaker.Lock()
	p.vol.Silent = false
	if p.vol.Volume > volLimitMin {
		p.vol.Volume -= volStep
	}
	speaker.Unlock()

	*reply = mapVolumeToScale(p.vol.Volume)
	log.Debugf("Volume set to %g", mapVolumeToScale(p.vol.Volume))

	return nil
}

// SetVol sets the player's volume to a specific level, adjusting within the maximum limit.
func (p *PlayerServer) SetVol(arg *int, reply *float64) error {
	vol := mapScaleToVolume(float64(*arg))
	speaker.Lock()
	if vol > volLimitMax {
		vol = volLimitMax
	}
	p.vol.Volume = vol
	p.vol.Silent = false
	speaker.Unlock()

	*reply = mapVolumeToScale(vol)
	log.Debugf("Volume turned down to %g", mapVolumeToScale(p.vol.Volume))

	return nil
}

// Queue adds songs to the end of the playlist by its file path.
// If the path is a .m3u or .m3u8 file, the entire playlist is loaded and queued.
func (p *PlayerServer) Queue(targetPath *string, reply *struct{}) error {
	if strings.HasSuffix(*targetPath, ".m3u") || strings.HasSuffix(*targetPath, ".m3u8") {
		songs, err := m3u.Load(*targetPath)
		if err != nil {
			return err
		}

		p.playlist.Queue(songs...)
		if p.currentSong == nil {
			p.currentSong = p.playlist.Head
		}
		log.Debugf("Playlist loaded, songs in queue: %d", p.playlist.Size())
		return nil
	}

	song, err := playlist.NewSong(*targetPath)
	if err != nil {
		return err
	}

	if p.currentSong == nil {
		p.currentSong = song
	}

	p.playlist.Queue(song)
	log.Debugf("Add song to playlist, songs in queue: %d", p.playlist.Size())

	return nil
}

// TrackInfo returns the metadata of the current song in the playlist.
func (p *PlayerServer) TrackInfo(args *struct{}, prop *playlist.AudioProperties) error {
	*prop = *p.currentSong.Prop

	return nil
}

// PlaylistInfo writes a formatted string containing the playlist contents,
// including song numbers, file names, and an indicator for the currently playing song.
func (p *PlayerServer) PlaylistInfo(args *struct{}, infoMsg *string) error {
	var sb strings.Builder
	numCap := int(math.Log10(float64(p.playlist.Size()))) + 1
	for i, song := range p.playlist.ListSongs() {
		if song == *p.currentSong {
			sb.WriteRune('►')
		} else {
			sb.WriteRune(' ')
		}
		sb.WriteString(fmt.Sprintf("%0*d [%s]\n", numCap, i+1, song.Prop.FileName))
	}

	*infoMsg = sb.String()
	return nil
}

// Next skips to the next track. If the current track is the last one, stops playback.
func (p *PlayerServer) Next(args *struct{}, reply *struct{}) error {
	speaker.Lock()
	if p.currentSong.Next == nil {
		close(p.done)
	} else {
		p.ctrl.Paused = true
		p.currentSong = p.currentSong.Next
		p.ready()
	}
	speaker.Unlock()

	return nil
}

// Rewind rewinds to the previous track. If the current track is the first one,
// rewinds to the start of the current track.
func (p *PlayerServer) Rewind(args *struct{}, reply *struct{}) error {
	speaker.Lock()
	if p.currentSong.Prev == nil {
		p.currentSong.Streamer.Seek(0)
	} else {
		p.ctrl.Paused = true
		p.currentSong = p.currentSong.Prev
		p.ready()
	}
	defer speaker.Unlock()

	return nil
}

// SavePlaylist saves the current playlist to a file in the M3U format at the specified path.
// If the path is "-", it will be replaced with the default playlist directory.
// If the directory does not exist, it will be created.
// The file name will be in the format "YYYY-MM-DD_HH-MM-SS.m3u".
// The method returns the path of the saved file through the reply parameter.
func (p *PlayerServer) SavePlaylist(dir *string, reply *string) error {
	if *dir == "-" {
		homeDir, err := env.GetHomeDir()
		if err != nil {
			return err
		}
		*dir = path.Join(homeDir, p.playlistDir)
	} else {
		if !env.PathExists(*dir) {
			return env.ErrInvalidPath
		}
	}

	err := os.Mkdir(*dir, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	t := time.Now()
	fileName := fmt.Sprintf("%d-%02d-%02d_%02d-%02d-%02d.m3u", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	err = m3u.Save(p.playlist, path.Join(*dir, fileName))
	if err != nil {
		return err
	}

	*reply = path.Join(*dir, fileName)
	return nil
}

// ready signals the playlist that it should send the next song to the SongChan channel.
func (p *PlayerServer) ready() {
	if p.currentSong != nil {
		p.currentSong.Streamer.Seek(0)
		p.playlist.SongChan <- p.currentSong
	} else {
		close(p.done)
	}
}

// nextSong Returns the channel for receiving songs from the playlist.
func (p *PlayerServer) nextSong() chan *playlist.Song {
	return p.playlist.SongChan
}

func NewPlayerServer(playlistDir string) *PlayerServer {
	p := PlayerServer{
		playlist:    playlist.NewPlaylist(),
		playlistDir: playlistDir,
		ctrl:        &beep.Ctrl{},
		vol:         &effects.Volume{},
		done:        make(chan struct{}),
	}

	return &p
}
