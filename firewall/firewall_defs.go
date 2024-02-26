package firewall

type firewall struct {
	adpName        string
	privateNetwork string
}

type Firewall interface {
	AllowTraffic() error
	StopTraffic() error
}
