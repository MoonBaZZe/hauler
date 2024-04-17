package common

import (
	"encoding/json"
	"github.com/btcsuite/btcd/rpcclient"
	"os"
	"path/filepath"
)

var DefaultNodeConfigFileName = "config.json"

type Config struct {
	DataPath     string // default ~/.hauler
	GlobalState  uint8
	NoMEndpoints []string

	ProducerKeyFileName       string
	ProducerKeyFilePassphrase string
	ProducerIndex             uint32

	BtcConfig *rpcclient.ConnConfig
}

func (c *Config) MakePathsAbsolute() error {
	if c.DataPath == "" {
		c.DataPath = DefaultDataDir()
	} else {
		absDataDir, err := filepath.Abs(c.DataPath)
		if err != nil {
			return err
		}
		c.DataPath = absDataDir
	}

	return nil
}

func WriteConfig(cfg Config) error {
	// second read default settings
	dataPath := cfg.DataPath
	configPath := filepath.Join(dataPath, DefaultNodeConfigFileName)

	configBytes, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(configPath, configBytes, 0644); err != nil {
		return err
	}
	return nil
}
