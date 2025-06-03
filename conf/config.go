package conf

import (
	"fmt"
	"os"
	"path"

	"github.com/BurntSushi/toml"

	"scythix/env"
)

const (
	defaultConfPath = ".config/scythix"
	confFileName    = "conf.toml"

	defaultLogLevel = "debug"

	defaultVolLevel   = 16
	defaultSampleRate = 44100
)

var (
	ErrLoadConfig  = fmt.Errorf("can't load config file")
	ErrTooManyArgs = fmt.Errorf("too many arguments for function call")
)

var HomeDir string

type config struct {
	VolLevel   float64 `toml:"volume_level"`
	LogLevel   string  `toml:"log_level"`
	SampleRate int     `toml:"sample_rate"`
}

func Load(argPath ...string) (*config, error) {
	homeDir, err := env.GetHomeDir()
	if err != nil {
		return nil, err
	}

	var confPath string
	if len(argPath) == 0 {
		confPath = path.Join(homeDir, defaultConfPath, confFileName)
	} else if len(argPath) == 1 {
		confPath = argPath[0]
	} else {
		return nil, ErrTooManyArgs
	}

	cfg := &config{}

	if _, err := toml.DecodeFile(confPath, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func CreateDefault() (*config, error) {
	homeDir, err := env.GetHomeDir()
	if err != nil {
		return nil, err
	}

	confPath := path.Join(homeDir, defaultConfPath)
	err = os.MkdirAll(confPath, 0755)
	if err != nil {
		return nil, err
	}

	conf := config{
		VolLevel:   defaultVolLevel,
		SampleRate: defaultSampleRate,
		LogLevel:   defaultLogLevel,
	}

	f, err := os.Create(path.Join(confPath, confFileName))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	encoder.Encode(conf)

	return &conf, nil
}

func Write(cfg *config) error {
	homeDir, err := env.GetHomeDir()
	if err != nil {
		return err
	}

	confPath := path.Join(homeDir, defaultConfPath, confFileName)
	f, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	encoder.Encode(cfg)

	return nil
}
