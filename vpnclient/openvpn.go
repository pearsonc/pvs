package vpnclient

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"pearson-vpn-service/app_config"
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
	if vpn.processManager.IsProcessRunning(vpn.processId) {
		logconfig.Log.Println("VPN process is already running, no need to start it.")
		return nil
	}
	logconfig.Log.Println("Starting OpenVPN...")
	var err error
	for i := 0; i < 3; i++ {
		err = vpn.startOpenVPN()
		if err == nil {
			return nil
		}
		logconfig.Log.Printf("Attempt %d to start OpenVPN failed: %v\n", i+1, err)
		initErr := vpn.configManager.Initialise()
		if initErr != nil {
			return initErr
		}
		time.Sleep(5 * time.Second)
	}
	return err
}
func (vpn *client) startOpenVPN() error {
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
	if waitErr := vpn.waitForConnection(scanner); waitErr != nil {
		_ = vpn.processManager.StopProcess(vpn.processId)
		return waitErr
	}
	vpn.allowTraffic()

	logconfig.Log.Println("Enabling VPN process monitor...")
	vpn.processManager.StartMonitor()
	go vpn.EnableAutoRotateVPN()

	return nil
}
func (vpn *client) StopVPN() error {
	logconfig.Log.Info("Stopping VPN...")
	vpn.processManager.StopMonitor()
	if !vpn.processManager.IsProcessRunning(vpn.processId) {
		logconfig.Log.Info("VPN stopped successfully")
		return nil
	} else {
		err := vpn.processManager.StopProcess(vpn.processId)
		if err != nil {
			return err
		}
	}
	vpn.stopTraffic()
	if vpn.cancelRotate != nil {
		logconfig.Log.Println("Cancelling VPN rotation")
		vpn.cancelRotate()
	}
	logconfig.Log.Println("VPN stopped successfully")
	return nil
}
func (vpn *client) RestartVPN() error {
	logconfig.Log.Info("Restarting VPN...")
	if err := vpn.StopVPN(); err != nil {
		return err
	}
	if err := vpn.StartVPN(); err != nil {
		return err
	}
	logconfig.Log.Info("VPN restarted successfully")
	return nil
}
func (vpn *client) EnableAutoRotateVPN() {
	ctx, cancel := context.WithCancel(context.Background())
	vpn.cancelRotate = cancel
	rotatePeriod := app_config.Config.GetInt64("openvpn.rotate_minutes")
	if rotatePeriod <= 0 {
		rotatePeriod = 15
	}
	logconfig.Log.Info("Enabling auto VPN rotation every ", rotatePeriod, " minute(s)")
	ticker := time.NewTicker(time.Duration(rotatePeriod) * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			break
		case <-ticker.C:
			if rotateErr := vpn.RotateVPN(); rotateErr != nil {
				logconfig.Log.Errorf("Error rotating VPN connection: %v\n", rotateErr)
				break
			}
		}
		return
	}

}
func (vpn *client) RotateVPN() error {
	logconfig.Log.Info("Rotating VPN connection...")
	initErr := vpn.configManager.Initialise()
	if initErr != nil {
		return initErr
	}
	stopErr := vpn.StopVPN()
	if stopErr != nil {
		return stopErr
	}
	startErr := vpn.StartVPN()
	if startErr != nil {
		return startErr
	}
	logconfig.Log.Info("Rotated VPN connection successfully")
	return nil
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
	logconfig.Log.Info("Waiting for OpenVPN connection to be established...")
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
	for {
		select {
		case msg := <-ch:
			if msg.Success {
				logconfig.Log.Info("OpenVPN connection established successfully!")
				return nil
			} else {
				failureOutput += msg.Line
			}
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				logconfig.Log.Warn("Timed out waiting for OpenVPN to connect.")
				logconfig.Log.Error("OpenVPN output: ", failureOutput)
			}
			return ctx.Err()
		}
	}
}
