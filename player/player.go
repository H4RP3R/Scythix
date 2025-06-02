package player

import (
	"flag"
	"fmt"
	"net/rpc"
	"os"
	"path"
	"path/filepath"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/h2non/filetype"
	log "github.com/sirupsen/logrus"

	"scythix/conf"
	"scythix/env"
)

const (
	logDir      = ".cache"
	logFileName = "scythix.log"
)

// pathExists returns true if the file at the given path exists and false otherwise.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

// normalizePath takes a string representation of a path and returns an absolute
// and clean representation of that path.
func normalizePath(path string) (string, error) {
	path = filepath.Clean(path)
	var err error
	if path, err = filepath.Abs(path); err != nil {
		return "", err
	}
	return path, nil
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

// mapVolumeToScale maps a volume scale starting at -12 with step 0.5 to a scale
// whose first value is 0 and step 1.
func mapVolumeToScale(vol float64) float64 {
	return (vol + 12) * 2
}

// mapScaleToVolume is the opposite of mapVolumeToScale and is used for setting the
// volume level to a certain value appropriate to a scale with an initial value of 0.
func mapScaleToVolume(scale float64) float64 {
	return (scale / 2) - 12
}

func connectRPC() *rpc.Client {
	client, err := rpc.Dial("unixpacket", "/tmp/scythix.sock")
	if err != nil {
		if !pathExists(lockFile) {
			log.Debug("Player server not running")
			os.Exit(0)
		}
		log.Fatalf("Player server connection failed: %v", err)
	}

	return client
}

func Run() {
	var (
		path     string
		queued   string
		pause    bool
		stop     bool
		next     bool
		rew      bool
		mute     bool
		turnUp   bool
		turnDown bool
		vol      int
		info     bool
		list     bool
	)

	flag.StringVar(&path, "play", "", "Starts playing the specified audio file")
	flag.StringVar(&queued, "queue", "", "Add the specified audio file to the playback queue")
	flag.BoolVar(&pause, "pause", false, "Pause playback")
	flag.BoolVar(&stop, "stop", false, "Stop playback")
	flag.BoolVar(&next, "next", false, "Next track")
	flag.BoolVar(&rew, "rew", false, "Rewind to previous track")
	flag.BoolVar(&mute, "mute", false, "Mute sound")
	flag.BoolVar(&turnUp, "turn-up", false, "Increase volume")
	flag.BoolVar(&turnDown, "turn-down", false, "Decrease volume")
	flag.IntVar(&vol, "vol", -1, "Set volume value")
	flag.BoolVar(&info, "info", false, "Display track info")
	flag.BoolVar(&list, "list", false, "Display current playlist")
	flag.Parse()

	switch {
	case pause == true:
		client := connectRPC()
		defer client.Close()
		if err := client.Call("PlayerServer.Pause", &struct{}{}, &struct{}{}); err != nil {
			log.Error(err)
		}
	case stop == true:
		client := connectRPC()
		defer client.Close()
		if err := client.Call("PlayerServer.Stop", &struct{}{}, &struct{}{}); err != nil {
			log.Error(err)
		} else {
			fmt.Println("See you.")
		}
	case next == true:
		client := connectRPC()
		defer client.Close()
		if err := client.Call("PlayerServer.Next", &struct{}{}, &struct{}{}); err != nil {
			log.Error(err)
		}
	case rew == true:
		client := connectRPC()
		defer client.Close()
		if err := client.Call("PlayerServer.Rewind", &struct{}{}, &struct{}{}); err != nil {
			log.Error(err)
		}
	case mute == true:
		client := connectRPC()
		defer client.Close()
		if err := client.Call("PlayerServer.Mute", &struct{}{}, &struct{}{}); err != nil {
			log.Error(err)
		}
	case turnUp == true:
		var volLvl float64
		client := connectRPC()
		defer client.Close()
		if err := client.Call("PlayerServer.TurnUp", &struct{}{}, &volLvl); err != nil {
			log.Error(err)
		} else {
			fmt.Printf("vol: %g\n", volLvl)
		}
	case turnDown == true:
		var volLvl float64
		client := connectRPC()
		defer client.Close()
		if err := client.Call("PlayerServer.TurnDown", &struct{}{}, &volLvl); err != nil {
			log.Error(err)
		} else {
			fmt.Printf("vol: %g\n", volLvl)
		}
	case vol > -1:
		var volLvl float64
		client := connectRPC()
		defer client.Close()
		if err := client.Call("PlayerServer.SetVol", &vol, &volLvl); err != nil {
			log.Error(err)
		} else if float64(vol) != volLvl {
			fmt.Printf("vol: %g\n", volLvl)
		}
	case info == true:
		var prop AudioProperties
		client := connectRPC()
		defer client.Close()
		if err := client.Call("PlayerServer.TrackInfo", &struct{}{}, &prop); err != nil {
			log.Error(err)
		} else {
			prop.Display()
		}
	case list == true:
		var playlist string
		client := connectRPC()
		defer client.Close()
		if err := client.Call("PlayerServer.PlaylistInfo", &struct{}{}, &playlist); err != nil {
			log.Error(err)
		} else {
			fmt.Println(playlist)
		}
	case path != "":
		if ok := pathExists(path); ok {
			if pathExists(lockFile) {
				log.Debug("Attempt to run more then one instance of the program")
				fmt.Println("Already in use")
			} else {
				path, err := normalizePath(path)
				if err != nil {
					log.Error(err)
					fmt.Printf("Failed to normalize path: %v", err)
					return
				}
				err = RunDaemon(path)
				if err != nil {
					log.Error(err)
					fmt.Printf("Unable to run Scythix: %v", err)
				}
			}
		} else {
			log.Fatal(ErrInvalidPath)
		}
	case queued != "":
		if ok := pathExists(queued); ok {
			// If the lock file exists, then playback is on.
			if pathExists(lockFile) {
				queued, err := normalizePath(queued)
				if err != nil {
					log.Error(err)
					fmt.Printf("Unable to get absolute file path: %v", err)
					return
				}
				client := connectRPC()
				defer client.Close()
				err = client.Call("PlayerServer.Queue", &queued, &struct{}{})
				if err != nil {
					log.Error(err)
					fmt.Printf("Unable to queue: %v", err)
				}
			}
		}
	}
}

func init() {
	homeDir, err := env.GetHomeDir()
	if err != nil {
		log.Error(err)
	}

	logDir := path.Join(homeDir, logDir)
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.Error(err)
	}

	logPath := path.Join(logDir, logFileName)
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	logLevel := log.DebugLevel

	if playerConf, err := conf.Load(); err == nil {
		levelMap := map[string]log.Level{
			"debug": log.DebugLevel,
			"info":  log.InfoLevel,
			"warn":  log.WarnLevel,
			"error": log.ErrorLevel,
			"fatal": log.FatalLevel,
			"panic": log.PanicLevel,
		}
		if level, ok := levelMap[playerConf.LogLevel]; ok {
			logLevel = level
		}
	}
	log.SetOutput(logFile)
	log.SetLevel(logLevel)
}
