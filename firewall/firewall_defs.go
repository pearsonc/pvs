package firewall

type firewall struct {
	enabled        bool
	adpName        string
	privateNetwork string
}

type Firewall interface {
	AllowTraffic() error
	StopTraffic() error
}
