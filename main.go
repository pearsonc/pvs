package main

import (
	"log"
	"pearson-vpn-service/api/web"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient"
	"time"
)

// main @TODO: change sleep time to a more reliable method check output of process for success
func main() {
	if vpnClient, err := vpnclient.NewClient(); err != nil {
		log.Fatalf("Error creating VPN client: %v", err)
	} else {
		if err := vpnClient.StartVPN(); err != nil {
			log.Fatalf("Error starting VPN: %v", err)
		}
		time.Sleep(10 * time.Second) // leave time for the VPN to start
		if processStatus, err := vpnClient.GetStatus(); err != nil {
			log.Fatalf("Error getting VPN status: %v", err)
		} else {
			if processStatus != supervisor.Running {
				log.Println("VPN Status: ", processStatus.String())
				log.Println("Process Output: ", vpnClient.GetProcessOutput())
				log.Fatalf("VPN failed to start")
			}
			server := web.NewServer(vpnClient)
			server.Start(8080)
		}
	}
}
