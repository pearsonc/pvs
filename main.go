package main

import (
	"log"
	"net/http"
	"pearson-vpn-service/supervisor"
	"pearson-vpn-service/vpnclient"
	"time"
)

func main() {
	sysctlSupervisor := supervisor.NewManager()
	vpnClient, err := vpnclient.NewClient(sysctlSupervisor)
	if err != nil {
		log.Fatalf("Error creating VPN client: %v", err)
	}
	err = vpnClient.StartVPN()
	log.Println("Attempting to connect to VPN...")
	if err != nil {
		log.Fatalf("Error starting VPN: %v", err)
	}
	time.Sleep(10 * time.Second)
	if sysctlSupervisor.GetStatus(vpnClient.ProcessIdName) != "running" {
		log.Println("VPN Status: ", sysctlSupervisor.GetStatus(vpnClient.ProcessIdName))
		log.Println("Process Output: ", sysctlSupervisor.GetProcessOutput(vpnClient.ProcessIdName))
		log.Fatalf("VPN failed to start")
	}
	log.Println("VPN connected successfully...")

	// Set up a simple HTTP server
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("PVS is running"))
		if err != nil {
			return
		}
	})

	// Start the HTTP server
	port := "80"
	log.Printf("Starting HTTP server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
