package vpnclient

import (
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient/openvpn/expressvpn"
)

type client struct {
	binary         string
	configManager  expressvpn.ConfigFileManager
	processManager supervisor.ProcessManager
	processId      string
}

type Client interface {
	StartVPN() error
	StopVPN() error
	RestartVPN() error
	RotateVPN() error
	GetActiveConfig() string
	GetConfigDir() string
	GetProcessId() string
	GetStatus() (supervisor.ProcessStatus, error)
	GetProcessOutput() string
}
