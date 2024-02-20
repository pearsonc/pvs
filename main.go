package main

import (
	"log"
	"net/http"
	"os"
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
	processStatus, err := sysctlSupervisor.GetStatus(vpnClient.ProcessIdName)
	if processStatus != supervisor.Running {
		log.Println("VPN Status: ", processStatus.String())
		log.Println("Process Output: ", sysctlSupervisor.GetProcessOutput(vpnClient.ProcessIdName))
		log.Fatalf("VPN failed to start")
	} else {
		log.Println("VPN Status: ", processStatus.String())
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
	startServer()

}

func startServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
