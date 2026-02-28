package web

import (
	"context"
	"net/http"
	"pearson-vpn-service/app_config"
	"pearson-vpn-service/logconfig"
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

	bindAddress := app_config.Config.GetString("api.bind_address")
	if bindAddress == "" {
		bindAddress = "127.0.0.1:8080"
	}

	server := &http.Server{Addr: bindAddress}

	go func() {
		// [Rule: Log to Files] Use zerolog instead of log.Fatal
		logconfig.Log.Fatal().Err(server.ListenAndServe()).Msg("HTTP API server stopped")
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
