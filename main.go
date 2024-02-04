package main

import (
	"log"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient"
)

func main() {
	log.Println("Starting pvs...")
	log.Println("Creating Supervisor")
	sysctlSupervisor := supervisor.NewManager()
	log.Println("Supervisor created")

	log.Println("Creating VPN Client")
	vpnClient, err := vpnclient.NewClient(sysctlSupervisor)
	log.Println("VPN Client created")

	if err != nil {
		log.Fatalf("Error creating VPN client: %v", err)
	}

	err = vpnClient.StartVPN()
	log.Println("VPN Started")

	if err != nil {
		log.Fatalf("Error starting VPN: %v", err)
	}

}
