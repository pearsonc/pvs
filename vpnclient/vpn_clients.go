package vpnclient

import "pearson-vpn-service/supervisor"

type Client struct {
	Binary         string
	Config         string
	ProcessManager *supervisor.ProcessManager
}

type ClientInterface interface {
	StartVPN() error
	StopVPN() error
	RestartVPN() error
	RotateVPN() error

	getConfig() *Client
	validateConfig() error
}
