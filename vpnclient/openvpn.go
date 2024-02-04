package vpnclient

import (
	"fmt"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient/expressvpn"
)

var _ ClientInterface = (*Client)(nil)

func NewClient(ProcessManager *supervisor.ProcessManager) (*Client, error) {

	conf, err := expressvpn.NewConfigFileManager()
	if err != nil {
		return nil, fmt.Errorf("error creating config manager: %w", err)
	}

	client := &Client{
		Binary:         "openvpn",
		ProcessManager: ProcessManager,
		ConfigManager:  conf,
	}
	return client, nil
}

func (vpn *Client) StartVPN() error {

	/*	connectionArgs := []string{"--config", vpn.ConfigManager.FileName, "--auth-nocache"}
		processID := vpn.ProcessManager.CreateProcess(vpn.Binary, connectionArgs...)
		err := vpn.ProcessManager.StartProcess(processID)

		if err != nil {
			return err
		}*/

	return nil
}

func (vpn *Client) StopVPN() error {
	//TODO implement me
	panic("implement me")
}

func (vpn *Client) RestartVPN() error {
	//TODO implement me
	panic("implement me")
}

func (vpn *Client) RotateVPN() error {
	//TODO implement me
	panic("implement me")
}

func (vpn *Client) getConfig() *Client {
	//TODO implement me
	panic("implement me")
}

func (vpn *Client) validateConfig() error {
	//TODO implement me
	panic("implement me")
}
