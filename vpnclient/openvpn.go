package vpnclient

import (
	"context"
	"fmt"
	"log"
	"pearson-vpn-service/firewall"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient/openvpn/expressvpn"
	"time"
)

func NewClient() (Client, error) {
	ProcessManager := supervisor.NewManager()
	conf, err := expressvpn.NewConfigFileManager()
	FirewallManager := firewall.NewFirewallManager()
	if err != nil {
		return nil, fmt.Errorf("error creating config manager: %w", err)
	}
	return &client{
		binary:          "openvpn",
		processManager:  ProcessManager,
		firewallManager: FirewallManager,
		configManager:   conf,
	}, nil
}

func (vpn *client) StartVPN() error {
	connectionArgs := []string{"--config", vpn.GetConfigDir() + vpn.GetActiveConfig(), "--auth-nocache"}
	vpn.processId = vpn.processManager.CreateProcess(vpn.binary, connectionArgs...)
	err := vpn.processManager.StartProcess(vpn.processId)
	if err != nil {
		return err
	}
	vpn.allowTraffic()
	vpn.processManager.StartMonitor()
	go vpn.EnableRotateVPN()
	return nil
}

func (vpn *client) StopVPN() error {

	if !vpn.processManager.IsProcessRunning(vpn.processId) {
		log.Println("VPN process is not running, no need to stop it.")
		return nil
	}
	vpn.stopTraffic()
	if vpn.cancelRotate != nil {
		vpn.cancelRotate()
	}
	err := vpn.processManager.StopProcess(vpn.processId)
	if err != nil {
		return err
	}
	return nil

	/*	vpn.stopTraffic()
		if vpn.cancelRotate != nil {
			vpn.cancelRotate()
		}
		err := vpn.processManager.StopProcess(vpn.processId)
		if err != nil {
			return err
		}
		return nil*/
}

// RestartVPN @TODO: Make rotation time configurable
func (vpn *client) RestartVPN() error {
	err := vpn.processManager.RestartProcess(vpn.processId)
	if err != nil {
		return err
	}
	return nil
}
func (vpn *client) EnableRotateVPN() {
	ctx, cancel := context.WithCancel(context.Background())
	vpn.cancelRotate = cancel
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			break
		case <-ticker.C:
			fmt.Println("Rotating VPN connection...")
			err := vpn.configManager.Initialise()
			if err != nil {
				fmt.Println("Error rotating VPN connection: ", err)
				break
			}
			vpn.processManager.StopMonitor()
			err = vpn.StopVPN()
			if err != nil {
				fmt.Println("Error rotating VPN connection: ", err)
				break
			}
			err = vpn.StartVPN()
			if err != nil {
				fmt.Println("Error rotating VPN connection: ", err)
				break
			}
			vpn.processManager.StartMonitor()
			fmt.Println("Rotated VPN connection successfully")
		}

		return
	}

}
func (vpn *client) GetActiveConfig() string {
	return vpn.configManager.GetFileName()
}
func (vpn *client) GetConfigDir() string {
	return vpn.configManager.GetConfigDir()
}
func (vpn *client) GetProcessId() string {
	return vpn.processId
}
func (vpn *client) GetStatus() (supervisor.ProcessStatus, error) {
	return vpn.processManager.GetStatus(vpn.processId)
}
func (vpn *client) GetProcessOutput() string {
	return vpn.processManager.GetProcessOutput(vpn.processId)
}
func (vpn *client) allowTraffic() {
	if fireErr := vpn.firewallManager.AllowTraffic(); fireErr != nil {
		log.Println("error allowing traffic:", fireErr)
		panic(fireErr)
	}
}
func (vpn *client) stopTraffic() {
	if fireErr := vpn.firewallManager.StopTraffic(); fireErr != nil {
		log.Println("error stopping traffic:", fireErr)
		panic(fireErr)
	}
}
