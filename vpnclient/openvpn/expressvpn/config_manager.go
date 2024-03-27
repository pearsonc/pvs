package expressvpn

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"pearson-vpn-service/app_config"
	"pearson-vpn-service/logconfig"
	"strings"
	"time"
)

func NewConfigFileManager() (ConfigFileManager, error) {

	customDir := app_config.Config.GetString("openvpn.config_dir")
	dir := ""
	if customDir != "" {
		dir = customDir
	} else {
		dir = "vpn_configs/"
	}
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	ConfigFile := &configFileManager{
		dir:              dir,
		preferredConfigs: app_config.Config.GetStringSlice("openvpn.preferred_configs"),
	}
	if err := ConfigFile.Initialise(); err != nil {
		return nil, err
	}
	return ConfigFile, nil
}

func (config *configFileManager) Initialise() error {
	file, err := config.getRandomConfigFile()
	if err != nil {
		return err
	}
	config.fileName = file
	err = config.validateConfigFile()
	if err != nil {
		return err
	}
	return nil
}
func (config *configFileManager) getRandomConfigFile() (string, error) {

	// Create random value using time
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	if len(config.preferredConfigs) > 0 {
		selectedConfig := config.preferredConfigs[rand.Intn(len(config.preferredConfigs))]
		fileName := selectedConfig

		logconfig.Log.Info("Preferred config files found, selected at random:", fileName)
		if _, err := os.Stat(config.dir + fileName); err == nil {
			return fileName, nil
		} else {
			return "", fmt.Errorf("the config file you provided does not exist: %v", fileName)
		}
	} else { // No Preferred configs found, move on and randomly select any
		dir, err := os.Open(config.dir)
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
		logconfig.Log.Info("No preferred config files found, selected a random file from config dir:", randomFile)
		return randomFile, nil
	}
}
func (config *configFileManager) validateConfigFile() error {

	if err := config.setupResolveConf(); err != nil {
		return err
	}
	if err := config.setupCiphersAndCerts(); err != nil {
		return err
	}
	if err := config.setupAuthUserPath(); err != nil {
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
func (config *configFileManager) setupResolveConf() error {

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
	content, err := os.ReadFile(config.dir + config.fileName)
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
		err = os.WriteFile(config.dir+config.fileName, []byte(text), 0644)
		if err != nil {
			return fmt.Errorf("error config manger - setupResolveConf error writing modified content back to file: %v", err)
		}
	}
	return nil
}
func (config *configFileManager) setupCiphersAndCerts() error {

	filePath := config.dir + config.fileName
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error config manger - setupCiphersAndCerts error reading file: %v", err)
	}
	text := string(content)
	text = strings.ReplaceAll(text, "cipher AES-256-CBC", "data-ciphers AES-256-GCM")
	text = strings.ReplaceAll(text, "keysize 256", "")
	text = strings.ReplaceAll(text, "ns-cert-type server", "remote-cert-tls server")

	// Remove any existing fallback cipher configuration, as BF-CBC is not supported and replace with AES-256-GCM
	text = strings.ReplaceAll(text, "\ndata-ciphers-fallback BF-CBC", "")
	if !strings.Contains(text, "\ndata-ciphers-fallback AES-256-GCM") {
		text += "\ndata-ciphers-fallback AES-256-GCM"
	} else {
		text = strings.ReplaceAll(text, "\ndata-ciphers-fallback AES-256-GCM", "")
		text += "\ndata-ciphers-fallback AES-256-GCM"
	}

	modifiedContent := []byte(text)
	err = os.WriteFile(filePath, modifiedContent, 0644)
	if err != nil {
		return fmt.Errorf("error config manger - setupCiphersAndCerts error writing modified content back to file: %v", err)
	}
	return nil
}

func (config *configFileManager) setupAuthUserPath() error {

	filePath := config.dir + config.fileName

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	desiredLine := "auth-user-pass /config/openvpn-credentials.txt"
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, "auth-user-pass") {
			lines[i] = desiredLine // Replace existing line
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, desiredLine)
	}
	updatedContent := strings.Join(lines, "\n")
	return os.WriteFile(filePath, []byte(updatedContent), 0644)
}

func (config *configFileManager) setupDefaultGateway() error {
	// Remove any existing redirect-gateway directives
	sedCmd := exec.Command("sed", "-i", "/redirect-gateway/d", config.dir+config.fileName)
	err := sedCmd.Run()
	if err != nil {
		return fmt.Errorf("error executing third sed command: %v", err)
	}
	sedCmd1 := exec.Command("sed", "-i", "$a redirect-gateway def1", config.dir+config.fileName)
	err = sedCmd1.Run()
	if err != nil {
		return fmt.Errorf("error executing fourth sed command: %v", err)
	}
	return nil
}
func (config *configFileManager) setupKeepAlive() error {
	// Check if "keepalive" already exists in the file
	grepCmdKeepAlive := exec.Command("grep", "-q", "keepalive 60 120", config.dir+config.fileName)
	err := grepCmdKeepAlive.Run()

	// If "keepalive" doesn't exist, add it
	if err != nil {
		sedCmdKeepAlive := exec.Command("sed", "-i", "$a keepalive 60 120", config.dir+config.fileName)
		err = sedCmdKeepAlive.Run()
		if err != nil {
			return fmt.Errorf("error adding keepalive settings: %v", err)
		}
	}

	return nil
}

func (config *configFileManager) GetConfigDir() string {
	return config.dir
}
func (config *configFileManager) GetFileName() string {
	return config.fileName
}
