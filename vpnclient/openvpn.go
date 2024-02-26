package vpnclient

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"pearson-vpn-service/firewall"
	"pearson-vpn-service/logconfig"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient/openvpn/expressvpn"
	"strings"
	"time"
)

type Message struct {
	Line    string
	Success bool
}

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

	// Wait for connection confirmation
	stdoutStream, err := vpn.processManager.GetStdoutStream(vpn.processId)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdoutStream)

	if waitErr := vpn.waitForConnection(scanner); waitErr != nil {
		return waitErr
	}

	vpn.allowTraffic()
	vpn.processManager.StartMonitor()
	go vpn.EnableRotateVPN()
	return nil
}

func (vpn *client) StopVPN() error {
	if !vpn.processManager.IsProcessRunning(vpn.processId) {
		logconfig.Log.Println("VPN process is not running, no need to stop it.")
		return nil
	} else {
		logconfig.Log.Println("VPN process is running, stopping it now")
		err := vpn.processManager.StopProcess(vpn.processId)
		if err != nil {
			return err
		}
	}
	logconfig.Log.Println("Stopping Firewall Traffic")
	vpn.stopTraffic()
	if vpn.cancelRotate != nil {
		logconfig.Log.Println("Cancelling VPN rotation")
		vpn.cancelRotate()
	}
	return nil
}

// RestartVPN @TODO: Make add checks for VPN connection status
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
			logconfig.Log.Println("Rotating VPN connection...")
			err := vpn.configManager.Initialise()
			if err != nil {
				logconfig.Log.Println("Error rotating VPN connection: ", err)
				fmt.Println("Error rotating VPN connection: ", err)
				break
			}
			vpn.processManager.StopMonitor()
			err = vpn.StopVPN()
			if err != nil {
				logconfig.Log.Println("Error rotating VPN connection: ", err)
				break
			}
			err = vpn.StartVPN()
			if err != nil {
				logconfig.Log.Println("Error rotating VPN connection: ", err)
				break
			}
			vpn.processManager.StartMonitor()
			logconfig.Log.Println("Rotated VPN connection successfully")
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
func (vpn *client) allowTraffic() {
	if fireErr := vpn.firewallManager.AllowTraffic(); fireErr != nil {
		logconfig.Log.Fatalf("error allowing traffic: %v", fireErr)
	}
}
func (vpn *client) stopTraffic() {
	if fireErr := vpn.firewallManager.StopTraffic(); fireErr != nil {
		logconfig.Log.Fatalf("error stopping traffic: %v", fireErr)
		panic(fireErr)
	}
}

func (vpn *client) waitForConnection(scanner *bufio.Scanner) error {
	ch := make(chan Message, 100)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "RTNETLINK") {
				continue
			}

			if strings.Contains(line, "Initialization Sequence Completed") {
				ch <- Message{Success: true}
				return
			} else if strings.Contains(line, "DEPRECATED OPTION:") || strings.Contains(line, "WARNING:") {
				ch <- Message{Line: line}
			} else {
				ch <- Message{Line: line}
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- Message{Line: err.Error()}
		}
		close(ch)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Listen for messages on the channel
	for {
		select {
		case msg := <-ch:
			if msg.Success {
				logconfig.Log.Println("OpenVPN connection established successfully!")
				return nil
			} else {
				// Debug by showing all output
				logconfig.Log.Println(msg)
				time.Sleep(1 * time.Second)
				//log.Println(msg)
				//time.Sleep(1 * time.Second)
			}
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logconfig.Log.Println("Timed out waiting for OpenVPN to connect.")
			}
			return ctx.Err()
		}
	}
}
