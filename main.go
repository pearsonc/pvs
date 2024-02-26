package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"pearson-vpn-service/api/web"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient"
	"syscall"
	"time"
)

// main @TODO: change sleep time to a more reliable method check output of process for success
func main() {

	fmt.Println("Starting VPN Service...")
	vpnClient, err := vpnclient.NewClient()
	if err != nil {
		log.Fatalf("Error creating VPN client: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
	go server.Start(8080)

	<-ctx.Done()
	fmt.Println("Shutdown signal received, stopping services...")
	if err := vpnClient.StopVPN(); err != nil {
		log.Fatalf("Failed to stop VPN: %v", err)
	}
	fmt.Println("VPN service stopped.")
}
