package player

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"syscall"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"
	log "github.com/sirupsen/logrus"

	"scythix/conf"
)

const lockFile = "/var/lock/scythix.lock"

const (
	defaultVol  float64 = -5
	volStep     float64 = 0.5
	volLimitMax float64 = 0
	volLimitMin float64 = -12
)

func RunDaemon(songPath string) error {
	// Check if the process is not a child (not forked).
	if _, isChild := os.LookupEnv("FORKED"); !isChild {
		// Fork and execute a new process with the same program arguments and
		// an additional environment variable "FORKED=1".
		exePath, err := os.Executable()
		if err != nil {
			log.Errorf("Failed to fork process: %v", err)
			return fmt.Errorf("%w: %v", ErrFailedToFork, err)
		}
		pid, err := syscall.ForkExec(exePath, os.Args, &syscall.ProcAttr{
			Env:   append(os.Environ(), "FORKED=1"),
			Files: []uintptr{uintptr(syscall.Stderr), uintptr(syscall.Stdout), uintptr(syscall.Stdin)},
		})
		if err != nil {
			log.Errorf("ForkExec failed: %v", err)
			return fmt.Errorf("%w: %v", ErrFailedToFork, err)
		}

		fmt.Printf("[PID:%d] Playing\n", pid)
		log.Debugf("Process forked with PID:%d", pid)
		os.Exit(0)
	}

	done := make(chan struct{})

	f, err := os.Create(lockFile)
	if err != nil {
		log.Error(err)
	} else {
		f.Close()
		log.Debug("Create lockfile")
	}

	playerConf, err := conf.Load()
	if err != nil {
		log.Debug(err)
	} else {
		log.Debug("Read config file")
	}

	if playerConf == nil {
		log.Debug("Create new config file")
		playerConf, err = conf.CreateDefault()
		if err != nil {
			fmt.Printf("Unable to create default config file: %v\n", err)
		}
	}

	// Check if the volume level is within acceptable limits.
	currentVol := mapScaleToVolume(playerConf.VolLevel)
	if currentVol > volLimitMax {
		currentVol = volLimitMax
		msg := "Volume level exceeds maximum (requested: %f, limited to: %f)"
		log.Debug(msg, currentVol, volLimitMax)
	} else if currentVol < volLimitMin {
		currentVol = volLimitMin
		msg := "Volume level is below minimum (requested: %f, limited to: %f)"
		log.Debug(msg, currentVol, volLimitMin)
	}

	srv := NewPlayerServer()
	srv.Queue(&songPath, &struct{}{})
	go srv.ready()

	defer func() {
		playerConf.VolLevel = mapVolumeToScale(srv.vol.Volume)
		conf.Write(playerConf)

		err := os.Remove(lockFile)
		if err != nil {
			log.Errorf("Unable to remove lock file: %v", err)
		} else {
			log.Debug("Remove lockfile")
		}
	}()

	rpc.Register(srv)
	listener, err := net.Listen("unixpacket", "/tmp/scythix.sock")
	if err != nil {
		return err
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				log.Error(err)
				return
			}
			rpc.ServeConn(conn)
		}
	}()

	speaker.Init(beep.SampleRate(44100), beep.SampleRate(44100).N(time.Second/10))

	for {
		select {
		case song, ok := <-srv.nextSong():
			if ok {
				defer song.Streamer.Close()
				srv.ctrl = &beep.Ctrl{Streamer: beep.Loop(1, song.Streamer), Paused: false}
				srv.vol = &effects.Volume{
					Streamer: srv.ctrl,
					Base:     2,
					Volume:   currentVol,
					Silent:   false,
				}
				resampled := beep.Resample(4, srv.currentSong.Format.SampleRate, beep.SampleRate(44100), srv.vol)
				speaker.Play(beep.Seq(resampled, beep.Callback(func() {
					currentVol = srv.vol.Volume
					srv.currentSong = srv.currentSong.Next
					srv.ready()
				})))
			} else {
				close(done)
				return nil
			}
		case <-srv.done:
			close(done)
			log.Debug("Player stopped.")
			return nil
		}
	}
}
