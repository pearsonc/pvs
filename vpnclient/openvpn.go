package vpnclient

import "pearson-vpn-service/supervisor"

var _ ClientInterface = (*Client)(nil)

func NewClient(ProcessManager *supervisor.ProcessManager) *Client {
	client := &Client{
		Binary:         "openvpn",
		ProcessManager: ProcessManager,
	}
	return client
}

func (vpn *Client) StartVPN() error {

	connectionArgs := []string{"--config", vpn.Config, "--auth-nocache"}
	processID := vpn.ProcessManager.CreateProcess(vpn.Binary, connectionArgs...)
	err := vpn.ProcessManager.StartProcess(processID)

	if err != nil {
		return err
	}

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
