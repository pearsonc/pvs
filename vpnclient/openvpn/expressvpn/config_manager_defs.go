package expressvpn

type configFileManager struct {
	dir              string
	preferredConfigs map[string]interface{}
	fileName         string
}

type ConfigFileManager interface {
	GetConfigDir() string
	GetFileName() string
	Initialise() error
}
