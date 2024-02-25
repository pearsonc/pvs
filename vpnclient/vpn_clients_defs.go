package vpnclient

import (
	"context"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient/openvpn/expressvpn"
)

type client struct {
	binary         string
	configManager  expressvpn.ConfigFileManager
	processManager supervisor.ProcessManager
	processId      string
	cancelRotate   context.CancelFunc
}

type Client interface {
	StartVPN() error
	StopVPN() error
	RestartVPN() error
	EnableRotateVPN()
	GetActiveConfig() string
	GetConfigDir() string
	GetProcessId() string
	GetStatus() (supervisor.ProcessStatus, error)
	GetProcessOutput() string
}
