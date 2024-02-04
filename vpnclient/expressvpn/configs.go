package expressvpn

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type ConfigFileManager struct {
	Dir              string
	PreferredConfigs string
	FileName         string
}

func NewConfigFileManager() (*ConfigFileManager, error) {
	ConfigFile := &ConfigFileManager{
		Dir:              "/home/chperso/20051/projects/snafu/pearson-vpn-service/bin/vpn_configs/", //configDir,
		PreferredConfigs: os.Getenv("VPN_CONFIGS"),
	}
	if file, err := ConfigFile.getRandomConfigFile(); err != nil {
		return nil, err
	} else {
		ConfigFile.FileName = file
	}

	return ConfigFile, nil
}

func (config *ConfigFileManager) setConfigFile() error {

	file, err := config.getRandomConfigFile()
	if err != nil {
		return err
	}
	config.FileName = file

	return nil
}

func (config *ConfigFileManager) validateConfigFile() error {
	//TODO implement me
	panic("implement me")
}

func (config *ConfigFileManager) getRandomConfigFile() (string, error) {

	// Create random value using time
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	if len(config.PreferredConfigs) > 0 {
		configList := strings.Split(config.PreferredConfigs, ",")
		randomConfig := strings.TrimSpace(configList[r.Intn(len(configList))])
		filePath := randomConfig
		if _, err := os.Stat(filePath); err == nil {
			return filePath, nil
		}
	} else { // No Preferred configs found, move on and randomly select any
		dir, err := os.Open(config.Dir)
		if err != nil {
			return "", fmt.Errorf("failed to open directory: %v", err)
		}
		defer func(dir *os.File) {
			err := dir.Close()
			if err != nil {
				log.Fatalf("error closing directory with error: %v", err)
			}
		}(dir)

		files, err := dir.Readdirnames(0) // 0 to read all files and folders
		if err != nil {
			return "", fmt.Errorf("failed to list files in directory: %v", err)
		}

		if len(files) == 0 {
			return "", fmt.Errorf("no config files found in directory")
		}

		randomFile := files[r.Intn(len(files))]
		log.Println("Using config file:", randomFile)
		return randomFile, nil
	}

	return "", fmt.Errorf("unable to find config files")
}
