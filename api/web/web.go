package web

import (
	"fmt"
	"log"
	"net/http"
	"pearson-vpn-service/vpnclient"
)

type Server struct {
	VpnClient vpnclient.Client
}

func NewServer(vpnClient vpnclient.Client) *Server {
	return &Server{
		VpnClient: vpnClient,
	}
}
func (s *Server) Start(port int) {
	http.HandleFunc("/status", s.handleStatus)
	log.Printf("Starting server on port %d...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	processStatus, _ := s.VpnClient.GetStatus()
	_, _ = w.Write([]byte("VPN Server Process is: " + processStatus.String() + "\n"))
	_, _ = w.Write([]byte("VPN Using Config: " + s.VpnClient.GetActiveConfig() + "\n"))
	_, _ = w.Write([]byte("VPN Config Directory: " + s.VpnClient.GetConfigDir() + "\n"))
}
