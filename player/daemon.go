package player

import (
	"fmt"
	"net"
	"net/http"
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
		pid, err := syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{
			Env:   append(os.Environ(), "FORKED=1"),
			Files: []uintptr{uintptr(syscall.Stderr), uintptr(syscall.Stdout), uintptr(syscall.Stdin)},
		})
		if err != nil {
			log.Errorf("ForkExec failed: %v\n", err)
			return err
		}

		fmt.Printf("[PID:%d] Playing\n", pid)
		log.Debugf("Process forked with PID:%d\n", pid)
		os.Exit(0)
	}

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

	defer func() {
		err := os.Remove(lockFile)
		if err != nil {
			log.Errorf("Unable to remove lock file: %v", err)
		} else {
			log.Debug("Remove lockfile")
		}
	}()

	srv := NePlayerServer()
	srv.Queue(&songPath, &struct{}{})
	srv.ready()

	rpc.Register(srv)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp4", ":4400")
	if err != nil {
		return err
	}

	go func() {
		err := http.Serve(listener, nil)
		if err != nil {
			log.Error(err)
		}
	}()

	var ok bool
	for {
		select {
		case srv.currentSong, ok = <-srv.nextSong():
			if ok {
				srv.ctrl = &beep.Ctrl{Streamer: beep.Loop(1, srv.currentSong.Streamer), Paused: false}
				srv.vol = &effects.Volume{
					Streamer: srv.ctrl,
					Base:     2,
					Volume:   currentVol,
					Silent:   false,
				}
				speaker.Init(srv.currentSong.Format.SampleRate, srv.currentSong.Format.SampleRate.N(time.Second/10))
				speaker.Play(beep.Seq(srv.vol, beep.Callback(func() {
					srv.ready()
				})))
			} else {
				return nil
			}
		case <-srv.done:
			log.Debug("Player stopped.")
			return nil
		}
	}
}
