package main

import (
	"context"
	"os"
	"os/signal"
	"pearson-vpn-service/api/web"
	"pearson-vpn-service/logconfig"
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
	logconfig.Log.Info("Starting VPN Service...")
	vpnClient, err := vpnclient.NewClient()
	if err != nil {
		logconfig.Log.Fatalf("Error creating VPN client: %v", err)
	}

	if err := vpnClient.StartVPN(); err != nil {
		logconfig.Log.Fatalf("Error starting VPN: %v", err)
	}

	time.Sleep(5 * time.Second)

	processStatus, err := vpnClient.GetStatus()
	if err != nil {
		logconfig.Log.Fatalf("Error getting VPN status: %v", err)
	} else if processStatus != supervisor.Running {
		logconfig.Log.Println("VPN Status: ", processStatus.String())
		logconfig.Log.Println("Process Output: ", vpnClient.GetProcessOutput())
		logconfig.Log.Fatalf("VPN failed to start")
	}

	logconfig.Log.Println("VPN started successfully")
	server := web.NewServer(vpnClient)
	go server.Start(ctx)

	select {
	case <-c:
		logconfig.Log.Println("Received shutdown signal")
		err := vpnClient.StopVPN()
		if err != nil {
			logconfig.Log.Fatalf("Failed to run vpnClient.StopVPN(): %v", err)
		}
		cancel()
	}
}
