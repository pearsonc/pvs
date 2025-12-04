package vpnclient

import (
	"context"
	"pearson-vpn-service/firewall"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient/openvpn"
)

type client struct {
	binary          string
	configManager   openvpn.ConfigFileManager
	processManager  supervisor.ProcessManager
	firewallManager firewall.Firewall
	processId       string
	cancelRotate    context.CancelFunc
	dnsCheckCancel  context.CancelFunc
	binaryOutput    bool
}

type Client interface {
	StartVPN() error
	StopVPN() error
	RestartVPN() error
	EnableAutoRotateVPN()
	GetActiveConfig() string
	GetConfigDir() string
	GetProcessId() string
	GetStatus() (supervisor.ProcessStatus, error)

	allowTraffic()
	stopTraffic()
}
