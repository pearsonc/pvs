package main

import (
	"context"
	"os"
	"os/signal"
	"pearson-vpn-service/api/web"
	"pearson-vpn-service/logconfig"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	// Handle signals in a goroutine so SIGTERM works even when the main
	// goroutine is blocked in StartVPN's retry loop.
	var vpnReady atomic.Bool
	var vpnClient vpnclient.Client
	go func() {
		<-c
		logconfig.Log.Println("Received shutdown signal")
		if vpnReady.Load() {
			done := make(chan struct{})
			go func() {
				if err := vpnClient.StopVPN(); err != nil {
					logconfig.Log.Errorf("Failed to stop VPN: %v", err)
				}
				close(done)
			}()
			select {
			case <-done:
				logconfig.Log.Println("PVS stopped")
			case <-time.After(30 * time.Second):
				logconfig.Log.Error("Shutdown timed out after 30s, forcing exit")
			}
		} else {
			logconfig.Log.Println("Shutdown during startup, nothing to clean up")
		}
		cancel()
		os.Exit(0)
	}()

	logconfig.Log.Info("Starting PVS Service...")
	var err error
	vpnClient, err = vpnclient.NewClient()
	if err != nil {
		logconfig.Log.Fatalf("Error creating VPN client: %v", err)
	}
	if err := vpnClient.StartVPN(); err != nil {
		logconfig.Log.Fatalf("Error starting VPN: %v", err)
	}
	processStatus, err := vpnClient.GetStatus()
	if err != nil {
		logconfig.Log.Fatalf("Error getting VPN status: %v", err)
	} else if processStatus != supervisor.Running {
		logconfig.Log.Println("VPN Status: ", processStatus.String())
		logconfig.Log.Fatalf("VPN failed to start")
	}

	vpnReady.Store(true)
	logconfig.Log.Println("PVS started successfully")
	server := web.NewServer(vpnClient)
	go server.Start(ctx)

	// Block until signal handler exits the process
	select {}
}
