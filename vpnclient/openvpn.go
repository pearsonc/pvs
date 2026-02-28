package vpnclient

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"pearson-vpn-service/app_config"
	"pearson-vpn-service/firewall"
	"pearson-vpn-service/logconfig"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient/openvpn"
	"strings"
	"syscall"
	"time"
)

type Message struct {
	Line    string
	Success bool
}

func NewClient() (Client, error) {
	ProcessManager := supervisor.NewManager()
	conf, err := openvpn.NewConfigFileManager()
	FirewallManager := firewall.NewFirewallManager()
	BinaryOutput := app_config.Config.GetBool("openvpn.output")
	if err != nil {
		return nil, fmt.Errorf("error creating config manager: %w", err)
	}
	maxRetries := app_config.Config.GetInt("monitoring.process_restart_limit")
	if maxRetries <= 0 {
		maxRetries = 3
	}
	return &client{
		binary:              "openvpn",
		processManager:      ProcessManager,
		firewallManager:     FirewallManager,
		configManager:       conf,
		binaryOutput:        BinaryOutput,
		maxRecoveryAttempts: maxRetries,
	}, nil
}
func (vpn *client) StartVPN() error {
	if vpn.processManager.IsProcessRunning(vpn.processId) {
		logconfig.Log.Info("VPN process is already running, no need to start it.")
		return nil
	}
	logconfig.Log.Infof("Starting VPN with config: %s%s", vpn.GetConfigDir(), vpn.GetActiveConfig())
	var err error
	for i := 0; i < 5; i++ {
		err = vpn.startOpenVPN()
		if err == nil {
			return nil
		}
		logconfig.Log.Warnf("StartVPN attempt %d/5 failed: %v, selecting new config", i+1, err)
		if initErr := vpn.configManager.Initialise(); initErr != nil {
			return fmt.Errorf("failed to initialise new config after attempt %d: %w", i+1, initErr)
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("all 5 StartVPN attempts failed, last error: %w", err)
}
func (vpn *client) startOpenVPN() error {
	connectionArgs := []string{"--config", vpn.GetConfigDir() + vpn.GetActiveConfig(), "--auth-nocache"}
	logconfig.Log.Infof("Launching OpenVPN: %s %v", vpn.binary, connectionArgs)
	processId, err := vpn.processManager.CreateProcess(vpn.binary, connectionArgs...)
	if err != nil {
		return fmt.Errorf("failed to create VPN process: %w", err)
	}
	vpn.processId = processId
	if err := vpn.processManager.StartProcess(vpn.processId); err != nil {
		return fmt.Errorf("failed to start VPN process %s: %w", vpn.processId, err)
	}
	stdoutStream, err := vpn.processManager.GetStdoutStream(vpn.processId)
	if err != nil {
		return fmt.Errorf("failed to get stdout stream for process %s: %w", vpn.processId, err)
	}
	scanner := bufio.NewScanner(stdoutStream)
	if waitErr := vpn.waitForConnection(scanner); waitErr != nil {
		_ = vpn.processManager.StopProcess(vpn.processId)
		return fmt.Errorf("VPN connection failed for process %s: %w", vpn.processId, waitErr)
	}
	vpn.allowTraffic()
	logconfig.Log.Info("Enabling VPN process monitor...")
	vpn.processManager.StartMonitor(func(processID string) {
		vpn.recoverVPN("process-monitor")
	})
	go vpn.EnableAutoRotateVPN()
	ctx, cancel := context.WithCancel(context.Background())
	vpn.dnsCheckCancel = cancel
	go vpn.StartNetworkCheck(ctx)

	return nil
}
func (vpn *client) StopVPN() error {
	logconfig.Log.Infof("Stopping VPN (process: %s, config: %s)", vpn.processId, vpn.GetActiveConfig())
	vpn.processManager.StopMonitor()
	if vpn.processManager.IsProcessRunning(vpn.processId) {
		if err := vpn.processManager.StopProcess(vpn.processId); err != nil {
			return fmt.Errorf("failed to stop VPN process %s: %w", vpn.processId, err)
		}
	}
	vpn.stopTraffic()
	if vpn.cancelRotate != nil {
		logconfig.Log.Info("Cancelling VPN rotation")
		vpn.cancelRotate()
	}
	if vpn.dnsCheckCancel != nil {
		vpn.dnsCheckCancel()
		vpn.dnsCheckCancel = nil
		logconfig.Log.Info("Stopping DNS resolution checks...")
	}
	logconfig.Log.Info("VPN stopped successfully")
	return nil
}
func (vpn *client) RestartVPN() error {
	logconfig.Log.Info("Restarting VPN...")
	if err := vpn.StopVPN(); err != nil {
		return fmt.Errorf("restart failed during stop: %w", err)
	}
	if err := vpn.StartVPN(); err != nil {
		return fmt.Errorf("restart failed during start: %w", err)
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
			return
		case <-ticker.C:
			if rotateErr := vpn.RotateVPN(); rotateErr != nil {
				logconfig.Log.Errorf("Timed rotation failed: %v", rotateErr)
				vpn.recoverVPN("timed-rotation")
				return
			}
			return // Successful rotation spawns new goroutine via startOpenVPN
		}
	}
}
func (vpn *client) RotateVPN() error {
	logconfig.Log.Info("Rotating VPN connection...")
	if err := vpn.configManager.Initialise(); err != nil {
		return fmt.Errorf("rotation failed during config initialise: %w", err)
	}
	logconfig.Log.Infof("New config selected for rotation: %s%s", vpn.GetConfigDir(), vpn.GetActiveConfig())
	if err := vpn.StopVPN(); err != nil {
		return fmt.Errorf("rotation failed during stop: %w", err)
	}
	if err := vpn.StartVPN(); err != nil {
		return fmt.Errorf("rotation failed during start: %w", err)
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
	}
}
func (vpn *client) recoverVPN(source string) {
	vpn.mu.Lock()
	if vpn.recovering {
		vpn.mu.Unlock()
		logconfig.Log.Infof("Recovery already in progress, ignoring trigger from %s", source)
		return
	}
	vpn.recovering = true
	vpn.mu.Unlock()

	defer func() {
		vpn.mu.Lock()
		vpn.recovering = false
		vpn.mu.Unlock()
	}()

	for attempt := 1; attempt <= vpn.maxRecoveryAttempts; attempt++ {
		logconfig.Log.Warnf("VPN recovery attempt %d/%d triggered by %s",
			attempt, vpn.maxRecoveryAttempts, source)

		if err := vpn.RotateVPN(); err != nil {
			logconfig.Log.Errorf("Recovery attempt %d/%d failed: %v",
				attempt, vpn.maxRecoveryAttempts, err)

			if attempt < vpn.maxRecoveryAttempts {
				backoff := time.Duration(attempt*10) * time.Second
				logconfig.Log.Infof("Waiting %v before next recovery attempt", backoff)
				time.Sleep(backoff)
			}
			continue
		}

		logconfig.Log.Infof("VPN recovery succeeded on attempt %d/%d", attempt, vpn.maxRecoveryAttempts)
		return
	}

	logconfig.Log.Errorf("All %d recovery attempts failed — requesting shutdown (last resort)",
		vpn.maxRecoveryAttempts)
	vpn.requestShutdown()
}

func (vpn *client) requestShutdown() {
	logconfig.Log.Errorf("CRITICAL: All self-recovery options exhausted, requesting process shutdown for systemd restart")
	_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
}
func (vpn *client) waitForConnection(scanner *bufio.Scanner) error {
	logconfig.Log.Info("Waiting for OpenVPN connection to be established...")
	ch := make(chan Message, 100)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if vpn.binaryOutput {
				logconfig.Log.Info(line)
			}
			if strings.Contains(line, "RTNETLINK") {
				continue
			}
			if strings.Contains(line, "Initialization Sequence Completed") {
				ch <- Message{Success: true}
				return
			}
			ch <- Message{Line: line}
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

func checkTCPReachability(target string) error {
	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		return fmt.Errorf("TCP reachability check to %s failed: %w", target, err)
	}
	if conn != nil {
		if err := conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection to %s: %w", target, err)
		}
	}
	return nil
}

func (vpn *client) StartNetworkCheck(ctx context.Context) {

	// Target a reliable host and port for a network check (e.g., Google HTTPS)
	const checkTarget = "www.google.com:443"

	checkPeriod := app_config.Config.GetInt64("openvpn.network_check_minutes")

	if checkPeriod <= 0 {
		checkPeriod = 1
	}

	logconfig.Log.Info("Enabling network reachability checks every: ", checkPeriod, " minute(s)")
	ticker := time.NewTicker(time.Duration(checkPeriod) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logconfig.Log.Info("Stopping network reachability check...")
			return

		case <-ticker.C:
			if err := checkTCPReachability(checkTarget); err != nil {
				logconfig.Log.Warnf("Network reachability check failed: %v", err)
				vpn.recoverVPN("network-check")
				return
			}
			logconfig.Log.Info("Network reachability check succeeded")
		}
	}
}
