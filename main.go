package main

import (
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient"
)

func main() {

	sysctlSupervisor := supervisor.NewManager()
	vpnClient := vpnclient.NewClient(sysctlSupervisor)

	err := vpnClient.StartVPN()
	if err != nil {
		return
	}

}
