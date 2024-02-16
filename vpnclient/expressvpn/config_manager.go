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
	if err := ConfigFile.initialise(); err != nil {
		return nil, err
	}
	return ConfigFile, nil
}
func (config *ConfigFileManager) initialise() error {
	file, err := config.getRandomConfigFile()
	if err != nil {
		return err
	}
	config.FileName = file
	log.Println("Config file set to:", config.FileName)
	log.Println("Config file path set to:", config.Dir)
	err = config.validateConfigFile()
	if err != nil {
		return err
	}
	return nil
}
func (config *ConfigFileManager) getRandomConfigFile() (string, error) {

	// Create random value using time
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	if len(config.PreferredConfigs) > 0 {
		configList := strings.Split(config.PreferredConfigs, ",")
		randomConfig := strings.TrimSpace(configList[r.Intn(len(configList))])
		fileName := randomConfig
		if _, err := os.Stat(config.Dir + fileName); err == nil {
			return fileName, nil
		} else {
			return "", fmt.Errorf("the config file you provided does not exist: %v", fileName)
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
	if err := config.setupCiphersAndCerts(); err != nil {
		return err
	}
	if err := config.setupDefaultGateway(); err != nil {
		return err
	}
	if err := config.setupKeepAlive(); err != nil {
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
	filePath := "/etc/openvpn/update-resolv-conf"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // update resolvconf script does not exist this is fine if system is not ubuntu/debian
	}
	requiredLines := []string{
		"script-security 2",
		"up /etc/openvpn/update-resolv-conf",
		"down /etc/openvpn/update-resolv-conf",
	}
	content, err := os.ReadFile(config.Dir + config.FileName)
	if err != nil {
		return fmt.Errorf("error config manger - setupResolveConf error reading file: %v", err)
	}
	text := string(content)
	allLinesPresent := true
	for _, line := range requiredLines {
		if !strings.Contains(text, line) {
			allLinesPresent = false
			break
		}
	}
	if !allLinesPresent {
		for _, line := range requiredLines {
			if !strings.Contains(text, line) {
				text += "\n" + line
			}
		}
		err = os.WriteFile(config.Dir+config.FileName, []byte(text), 0644)
		if err != nil {
			return fmt.Errorf("error config manger - setupResolveConf error writing modified content back to file: %v", err)
		}
	}
	return nil
}
func (config *ConfigFileManager) setupCiphersAndCerts() error {

	filePath := config.Dir + config.FileName
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error config manger - setupCiphersAndCerts error reading file: %v", err)
	}
	text := string(content)
	text = strings.ReplaceAll(text, "cipher AES-256-CBC", "data-ciphers AES-256-GCM")
	text = strings.ReplaceAll(text, "keysize 256", "")
	text = strings.ReplaceAll(text, "ns-cert-type server", "remote-cert-tls server")

	// Remove any existing fallback cipher configuration, as BF-CBC is not supported
	text = strings.ReplaceAll(text, "\ndata-ciphers-fallback BF-CBC", "")
	// Set a secure cipher as fallback
	text += "\ndata-ciphers-fallback AES-256-GCM"

	modifiedContent := []byte(text)
	err = os.WriteFile(filePath, modifiedContent, 0644)
	if err != nil {
		return fmt.Errorf("error config manger - setupCiphersAndCerts error writing modified content back to file: %v", err)
	}
	return nil
}
func (config *ConfigFileManager) setupDefaultGateway() error {
	// Remove any existing redirect-gateway directives
	sedCmd := exec.Command("sed", "-i", "/redirect-gateway/d", config.Dir+config.FileName)
	err := sedCmd.Run()
	if err != nil {
		return fmt.Errorf("error executing third sed command: %v", err)
	}
	sedCmd1 := exec.Command("sed", "-i", "$a redirect-gateway def1", config.Dir+config.FileName)
	err = sedCmd1.Run()
	if err != nil {
		return fmt.Errorf("error executing fourth sed command: %v", err)
	}
	return nil
}
func (config *ConfigFileManager) setupKeepAlive() error {
	// Check if "keepalive" already exists in the file
	grepCmdKeepAlive := exec.Command("grep", "-q", "keepalive 60 120", config.Dir+config.FileName)
	err := grepCmdKeepAlive.Run()

	// If "keepalive" doesn't exist, add it
	if err != nil {
		sedCmdKeepAlive := exec.Command("sed", "-i", "$a keepalive 60 120", config.Dir+config.FileName)
		err = sedCmdKeepAlive.Run()
		if err != nil {
			return fmt.Errorf("error adding keepalive settings: %v", err)
		}
	}

	return nil
}
