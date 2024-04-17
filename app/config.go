package app

import (
	"encoding/json"
	"github.com/MoonBaZZe/hauler/common"
	"github.com/MoonBaZZe/hauler/node"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func MakeConfig() (*common.Config, error) {
	cfg := node.DefaultNodeConfig

	// 1: Load config file.
	err := readConfigFromFile(&cfg)
	if err != nil {
		return nil, err
	}

	// 2: Make dir paths absolute
	if err := cfg.MakePathsAbsolute(); err != nil {
		return nil, err
	}

	// 3: Log config
	if j, err := json.MarshalIndent(cfg, "", "    "); err == nil {
		common.GlobalLogger.Info("Using the following hauler config: \n", string(j))
	}
	common.GlobalLogger.Info("using hauler config", cfg)

	// 4: Write it so a default one is created after the first run
	if err := common.WriteConfig(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func readConfigFromFile(cfg *common.Config) error {
	// second read default settings
	dataPath := cfg.DataPath
	configPath := filepath.Join(dataPath, common.DefaultNodeConfigFileName)
	if err := os.MkdirAll(dataPath, os.ModePerm); err != nil {
		return err
	}

	if jsonConf, err := ioutil.ReadFile(configPath); err == nil {
		err = json.Unmarshal(jsonConf, &cfg)
		if err == nil {
			return nil
		}
		log.Print("Config malformed: please check", "error", err)
		return err
	} else {
		log.Print("Config file missing: you can provide a data path using the --data flag or provide a config file using the --config flag", "configPath", configPath)
	}
	return nil
}
