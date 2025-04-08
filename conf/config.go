package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"scythix/env"
)

const (
	defaultConfPath = ".config/scythix"
	confFileName    = "conf.json"
)

var (
	ErrLoadConfig  = fmt.Errorf("can't load config file")
	ErrTooManyArgs = fmt.Errorf("too many arguments for function call")
)

var HomeDir string

type config struct {
	VolLevel float64 `json:"volume_level"`
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

	if _, err := os.Stat(confPath); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadConfig, err)
	}

	confData, err := os.ReadFile(confPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(confData, cfg)
	if err != nil {
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

	defaultConf := config{
		VolLevel: 16,
	}

	f, err := os.Create(path.Join(confPath, confFileName))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	confData, err := json.Marshal(defaultConf)
	if err != nil {
		return nil, err
	}

	_, err = f.Write(confData)
	if err != nil {
		return nil, err
	}

	return &defaultConf, nil
}

func Write(cfg *config) error {
	homeDir, err := env.GetHomeDir()
	if err != nil {
		return err
	}

	confData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	confPath := path.Join(homeDir, defaultConfPath, confFileName)
	f, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(confData)
	if err != nil {
		return err
	}

	return nil
}
