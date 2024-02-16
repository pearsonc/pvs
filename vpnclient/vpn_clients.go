package vpnclient

import (
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient/expressvpn"
)

type Client struct {
	Binary         string
	ConfigManager  *expressvpn.ConfigFileManager
	ProcessManager *supervisor.ProcessManager
	ProcessIdName  string
}

type ClientInterface interface {
	StartVPN() error
	StopVPN() error
	RestartVPN() error
	RotateVPN() error

	getConfig() *Client
	validateConfig() error
}
