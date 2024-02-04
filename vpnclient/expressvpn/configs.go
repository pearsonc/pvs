package expressvpn

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
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
		Dir:              "vpn_configs/",                                                  //configDir,
		PreferredConfigs: "my_expressvpn_andorra_udp.ovpn,my_expressvpn_austria_udp.ovpn", //os.Getenv("VPN_CONFIGS"),
	}
	if err := ConfigFile.setConfigFile(); err != nil {
		return nil, err
	}

	return ConfigFile, nil
}

func (config *ConfigFileManager) setConfigFile() error {

	file, err := config.getRandomConfigFile()
	log.Println("Using config file:", file)
	if err != nil {
		return err
	}
	err = config.validateConfigFile()
	if err != nil {
		return err
	}
	config.FileName = file

	return nil
}

func (config *ConfigFileManager) getRandomConfigFile() (string, error) {

	// Create random value using time
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	if len(config.PreferredConfigs) > 0 {
		configList := strings.Split(config.PreferredConfigs, ",")
		randomConfig := strings.TrimSpace(configList[r.Intn(len(configList))])
		filePath := config.Dir + randomConfig
		if _, err := os.Stat(filePath); err == nil {
			return filePath, nil
		} else {
			return "", fmt.Errorf("the config file you provided does not exist: %v", filePath)
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
}

func (config *ConfigFileManager) validateConfigFile() error {

	if err := config.setupResolveConf(); err != nil {
		return err
	}
	return nil
}

func (config *ConfigFileManager) setupResolveConf() error {
	// Check if resolvconf is installed
	_, err := exec.LookPath("resolvconf")
	if err != nil {
		return fmt.Errorf("resolvconf is not installed, please install it to proceed")
	}

	// Check if the necessary resolvconf lines already exist in the file
	grepCmdResolvConf := exec.Command("grep", "-q", "up /etc/openvpn/update-resolv-conf", config.Dir+config.FileName)
	err = grepCmdResolvConf.Run()

	// If the necessary lines don't exist, add them

	log.Println(config.Dir + config.FileName)

	if err != nil {
		sedCmdResolvConf := exec.Command("sed", "-i", "$a script-security 2\\nup /etc/openvpn/update-resolv-conf\\ndown /etc/openvpn/update-resolv-conf", config.Dir+config.FileName)
		err1 := sedCmdResolvConf.Run()
		if err1 != nil {
			return fmt.Errorf("error adding resolvconf settings: %v", err1)
		}
	}

	return nil
}
