package vpnclient

import (
	"fmt"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient/expressvpn"
)

var _ ClientInterface = (*Client)(nil)

func NewClient(ProcessManager *supervisor.ProcessManager) (*Client, error) {

	conf, err := expressvpn.NewConfigFileManager()
	if err != nil {
		return nil, fmt.Errorf("error creating config manager: %w", err)
	}

	client := &Client{
		Binary:         "openvpn",
		ProcessManager: ProcessManager,
		ConfigManager:  conf,
	}
	return client, nil
}

func (vpn *Client) StartVPN() error {

	/*	connectionArgs := []string{"--config", vpn.ConfigManager.FileName, "--auth-nocache"}
		processID := vpn.ProcessManager.CreateProcess(vpn.Binary, connectionArgs...)
		err := vpn.ProcessManager.StartProcess(processID)

		if err != nil {
			return err
		}*/

	return nil
}

func (vpn *Client) StopVPN() error {
	//TODO implement me
	panic("implement me")
}

func (vpn *Client) RestartVPN() error {
	//TODO implement me
	panic("implement me")
}

func (vpn *Client) RotateVPN() error {
	//TODO implement me
	panic("implement me")
}

func (vpn *Client) getConfig() *Client {
	//TODO implement me
	panic("implement me")
}

func (vpn *Client) validateConfig() error {
	//TODO implement me
	panic("implement me")
}

/*func (ipc *IpcOpenVPN) ValidateConfig(configFile string) error {

	// Check if resolvconf is installed
	_, err := exec.LookPath("resolvconf")
	if err != nil {
		return fmt.Errorf("resolvconf is not installed, please install it to proceed")
	}

	// Check if the necessary resolvconf lines already exist in the file
	grepCmdResolvConf := exec.Command("grep", "-q", "up /etc/openvpn/update-resolv-conf", configFile)
	err = grepCmdResolvConf.Run()

	// If the necessary lines don't exist, add them
	if err != nil {
		sedCmdResolvConf := exec.Command("sed", "-i", "$a script-security 2\\nup /etc/openvpn/update-resolv-conf\\ndown /etc/openvpn/update-resolv-conf", configFile)
		err = sedCmdResolvConf.Run()
		if err != nil {
			return fmt.Errorf("error adding resolvconf settings: %v", err)
		}
	}

	// Update the cipher and remove keysize 256
	sedCmd1 := exec.Command("sed", "-i", "-e", "s/cipher AES-256-CBC/data-ciphers AES-256-GCM /", "-e", "/keysize 256/d", configFile)
	// Replace deprecated ns-cert-type with remote-cert-tls
	sedCmd2 := exec.Command("sed", "-i", "s/ns-cert-type server/remote-cert-tls server/", configFile)
	// Remove any existing redirect-gateway directives
	sedCmd3 := exec.Command("sed", "-i", "/redirect-gateway/d", configFile)

	err = sedCmd1.Run()
	if err != nil {
		return fmt.Errorf("error executing first sed command: %v", err)
	}
	err = sedCmd2.Run()
	if err != nil {
		return fmt.Errorf("error executing second sed command: %v", err)
	}
	err = sedCmd3.Run()
	if err != nil {
		return fmt.Errorf("error executing third sed command: %v", err)
	}

	// Check if "redirect-gateway def1" already exists in the file
	grepCmd := exec.Command("grep", "-q", "redirect-gateway def1", configFile)
	err = grepCmd.Run()

	// If "redirect-gateway def1" doesn't exist, add it
	if err != nil {
		sedCmd4 := exec.Command("sed", "-i", "$a redirect-gateway def1", configFile)
		err = sedCmd4.Run()
		if err != nil {
			return fmt.Errorf("error executing fourth sed command: %v", err)
		}
	}

	// Check if "keepalive" already exists in the file
	grepCmdKeepAlive := exec.Command("grep", "-q", "keepalive 60 120", configFile)
	err = grepCmdKeepAlive.Run()

	// If "keepalive" doesn't exist, add it
	if err != nil {
		sedCmdKeepAlive := exec.Command("sed", "-i", "$a keepalive 60 120", configFile)
		err = sedCmdKeepAlive.Run()
		if err != nil {
			return fmt.Errorf("error adding keepalive settings: %v", err)
		}
	}

	return nil
}

func (ipc *IpcOpenVPN) getConfig(envVarName string) (string, error) {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// Try to get a random config from the environment variable first
	configs := os.Getenv(envVarName)
	if configs != "" {
		configList := strings.Split(configs, ",")
		if len(configList) > 0 {
			randomConfig := strings.TrimSpace(configList[r.Intn(len(configList))])
			filePath := "vpn-configs/" + randomConfig

			// Check if the file exists
			if _, err := os.Stat(filePath); err == nil {
				log.Println("Using random config file from environment var:", randomConfig)
				return filePath, nil
			} else if os.IsNotExist(err) {
				log.Println("The config file you provided does not exist", filePath)
				log.Println("Using a random config file from the vpn-configs folder")
			} else {
				log.Println("Error checking file:", err)
				return "", err
			}
		}
	}

	// If not found in environment variable or the list is empty, get a random config from the folder
	dir, err := os.Open("vpn-configs")
	if err != nil {
		return "", fmt.Errorf("failed to open directory: %v", err)
	}
	defer func(dir *os.File) {
		err := dir.Close()
		if err != nil {
			log.Println("error closing directory:", err)
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
	return "vpn-configs/" + randomFile, nil
}*/
