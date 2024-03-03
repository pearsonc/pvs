package expressvpn

type configFileManager struct {
	dir              string
	preferredConfigs []string
	fileName         string
}

type ConfigFileManager interface {
	GetConfigDir() string
	GetFileName() string
	Initialise() error
}
