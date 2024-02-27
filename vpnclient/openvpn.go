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
	stdoutStream, err := vpn.processManager.GetStdoutStream(vpn.processId)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(stdoutStream)
	logconfig.Log.Println("Waiting for OpenVPN connection to be established...")
	if waitErr := vpn.waitForConnection(scanner); waitErr != nil {
		return waitErr
	}
	logconfig.Log.Println("OpenVPN connection established successfully!")
	logconfig.Log.Println("Routing all traffic through VPN...")
	vpn.allowTraffic()
	logconfig.Log.Println("Enabling VPN process monitor...")
	vpn.processManager.StartMonitor()
	logconfig.Log.Println("Enabling VPN rotation...")
	go vpn.EnableRotateVPN()

	return nil
}

func (vpn *client) StopVPN() error {

	vpn.processManager.StopMonitor()

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
	logconfig.Log.Println("Blocking all traffic")
	vpn.stopTraffic()
	if vpn.cancelRotate != nil {
		logconfig.Log.Println("Cancelling VPN rotation")
		vpn.cancelRotate()
	}
	logconfig.Log.Println("VPN process stopped successfully")
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
			initErr := vpn.configManager.Initialise()
			if initErr != nil {
				logconfig.Log.Println("Error rotating VPN connection: ", initErr)
				fmt.Println("Error rotating VPN connection: ", initErr)
				break
			}
			stopErr := vpn.StopVPN()
			if stopErr != nil {
				logconfig.Log.Println("Error stopping vpn during rotation: ", stopErr)
				break
			}
			startErr := vpn.StartVPN()
			if startErr != nil {
				logconfig.Log.Println("Error starting vpn during rotation: ", startErr)
				break
			}
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

	var failureOutput string

	// Listen for messages on the channel
	for {
		select {
		case msg := <-ch:
			if msg.Success {
				logconfig.Log.Println("OpenVPN connection established successfully!")
				return nil
			} else {
				failureOutput += msg.Line
			}
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logconfig.Log.Println("Timed out waiting for OpenVPN to connect.")
				logconfig.Log.Println("OpenVPN output: ", failureOutput)
			}
			return ctx.Err()
		}
	}
}
