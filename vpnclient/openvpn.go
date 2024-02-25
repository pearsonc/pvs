package vpnclient

import (
	"fmt"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient/openvpn/expressvpn"
)

func NewClient() (Client, error) {
	ProcessManager := supervisor.NewManager()
	conf, err := expressvpn.NewConfigFileManager()
	if err != nil {
		return nil, fmt.Errorf("error creating config manager: %w", err)
	}
	return &client{
		binary:         "openvpn",
		processManager: ProcessManager,
		configManager:  conf,
	}, nil
}

func (vpn *client) StartVPN() error {
	connectionArgs := []string{"--config", vpn.GetConfigDir() + vpn.GetActiveConfig(), "--auth-nocache"}
	vpn.processId = vpn.processManager.CreateProcess(vpn.binary, connectionArgs...)
	err := vpn.processManager.StartProcess(vpn.processId)
	if err != nil {
		return err
	}
	return nil
}

func (vpn *client) StopVPN() error {
	err := vpn.processManager.StopProcess(vpn.processId)
	if err != nil {
		return err
	}
	return nil
}

func (vpn *client) RestartVPN() error {
	err := vpn.processManager.RestartProcess(vpn.processId)
	if err != nil {
		return err
	}
	return nil
}

func (vpn *client) RotateVPN() error {
	err := vpn.StopVPN()
	if err != nil {
		return err
	}
	err = vpn.configManager.Initialise()
	if err != nil {
		return err
	}
	err = vpn.StartVPN()
	if err != nil {
		return err
	}
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

func (vpn *client) GetProcessOutput() string {
	return vpn.processManager.GetProcessOutput(vpn.processId)
}
