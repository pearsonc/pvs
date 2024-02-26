package web

import (
	"context"
	"log"
	"net/http"
	"os"
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
func (s *Server) Start(ctx context.Context) {

	http.HandleFunc("/", s.handleStatus)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}

	server := &http.Server{Addr: ":" + port}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	<-ctx.Done()
	server.Shutdown(ctx)
}
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	processStatus, _ := s.VpnClient.GetStatus()
	_, _ = w.Write([]byte("VPN Server Process is: " + processStatus.String() + "\n"))
	_, _ = w.Write([]byte("VPN Using Config: " + s.VpnClient.GetActiveConfig() + "\n"))
	_, _ = w.Write([]byte("VPN Config Directory: " + s.VpnClient.GetConfigDir() + "\n"))
}
