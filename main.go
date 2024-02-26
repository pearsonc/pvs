package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"pearson-vpn-service/api/web"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient"
	"syscall"
	"time"
)

// main @TODO: change sleep time to a more reliable method check output of process for success
func main() {

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Starting VPN Service...")
	vpnClient, err := vpnclient.NewClient()
	if err != nil {
		log.Fatalf("Error creating VPN client: %v", err)
	}

	if err := vpnClient.StartVPN(); err != nil {
		log.Fatalf("Error starting VPN: %v", err)
	}

	time.Sleep(5 * time.Second)

	processStatus, err := vpnClient.GetStatus()
	if err != nil {
		log.Fatalf("Error getting VPN status: %v", err)
	} else if processStatus != supervisor.Running {
		log.Println("VPN Status: ", processStatus.String())
		log.Println("Process Output: ", vpnClient.GetProcessOutput())
		log.Fatalf("VPN failed to start")
	}

	fmt.Printf("VPN started successfully\n")
	server := web.NewServer(vpnClient)
	go server.Start(ctx)

	select {
	case <-c:
		log.Println("Received shutdown signal")
		err := vpnClient.StopVPN()
		if err != nil {
			log.Fatalf("Failed to run vpnClient.StopVPN(): %v", err)
		}
		cancel()
	}
}
