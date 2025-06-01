package player

import (
	"fmt"
	"math"
	"strings"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"
	log "github.com/sirupsen/logrus"
)

// PlayerServer represents a server for managing music playback via RPC.
type PlayerServer struct {
	PID         int
	playlist    *Playlist
	currentSong *Song

	ctrl *beep.Ctrl
	vol  *effects.Volume

	done chan struct{}
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
	close(p.done)

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

// Queue adds a song by path to the playlist.
func (p *PlayerServer) Queue(songPath *string, reply *struct{}) error {
	song, err := NewSong(*songPath)
	if err != nil {
		return err
	}

	p.playlist.Queue(song)
	log.Debugf("Add song to playlist, songs in queue: %d", p.playlist.Size())

	return nil
}

// TrackInfo returns the metadata of the current song in the playlist.
func (p *PlayerServer) TrackInfo(args *struct{}, prop *AudioProperties) error {
	*prop = *p.currentSong.Prop

	return nil
}

// PlaylistInfo writes a formatted string containing the playlist contents,
// including song numbers, file names, and an indicator for the currently playing song.
func (p *PlayerServer) PlaylistInfo(args *struct{}, infoMsg *string) error {
	var sb strings.Builder
	numCap := int(math.Log10(float64(p.playlist.Size())))
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
// Otherwise, closes the current song's streamer to start the next one.
func (p *PlayerServer) Next(args *struct{}, reply *struct{}) error {
	speaker.Lock()
	if p.currentSong.Next == nil {
		close(p.done)
	} else {
		p.currentSong.Streamer.Close()
	}
	speaker.Unlock()

	return nil
}

// ready signals the playlist that it should send the next song to the SongChan channel.
func (p *PlayerServer) ready() {
	if p.currentSong != nil && p.currentSong.Next == nil {
		close(p.playlist.NextChan)
	} else {
		p.playlist.NextChan <- struct{}{}
	}
}

// nextSong Returns the channel for receiving songs from the playlist.
func (p *PlayerServer) nextSong() chan *Song {
	return p.playlist.SongChan
}

func NewPlayerServer() *PlayerServer {
	p := PlayerServer{
		playlist: NewPlaylist(),
		ctrl:     &beep.Ctrl{},
		vol:      &effects.Volume{},
		done:     make(chan struct{}),
	}

	return &p
}
