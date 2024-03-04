package firewall

import (
	"fmt"
	"log"
	"os/exec"
	"pearson-vpn-service/app_config"
)

func NewFirewallManager() Firewall {

	adp := app_config.Config.GetString("firewall.adpName")
	privateNetwork := app_config.Config.GetString("firewall.privateNetwork")
	enabled := app_config.Config.GetBool("firewall.enabled")
	if adp == "" {
		adp = "eth1"
	}
	if privateNetwork == "" {
		privateNetwork = "192.168.1.0/24"
	}
	return &firewall{
		enabled:        enabled,
		adpName:        adp,
		privateNetwork: privateNetwork,
	}
}
func (f *firewall) AllowTraffic() error {
	if !f.enabled {
		return nil
	}
	if err := f.clearFirewall(); err != nil {
		return err
	}
	if cmdErr := f.executeCommands([]struct {
		args []string
		desc string
	}{
		{[]string{"-t", "nat", "-A", "POSTROUTING", "-s", f.privateNetwork, "-o", "tun0", "-j", "MASQUERADE"}, "NAT the traffic"},
	}); cmdErr != nil {
		return fmt.Errorf("could not execute commands: %w", cmdErr)
	}
	return nil
}
func (f *firewall) StopTraffic() error {
	if !f.enabled {
		return nil
	}
	if err := f.clearFirewall(); err != nil {
		return fmt.Errorf("could not clear firewall: %w", err)
	}
	if cmdErr := f.executeCommands([]struct {
		args []string
		desc string
	}{
		{[]string{"-A", "OUTPUT", "-j", "DROP"}, "block all outgoing traffic"},
		{[]string{"-I", "OUTPUT", "-o", f.adpName, "-j", "ACCEPT"}, "allow outgoing traffic on " + f.adpName},
		{[]string{"-I", "OUTPUT", "-p", "udp", "--dport", "1195", "-j", "ACCEPT"}, "allow UPP outgoing traffic on port 1195"},
		{[]string{"-I", "OUTPUT", "-p", "tcp", "--dport", "1195", "-j", "ACCEPT"}, "allow TCP outgoing traffic on port 1195"},
		{[]string{"-I", "OUTPUT", "-p", "udp", "--dport", "53", "-j", "ACCEPT"}, "allow UDP outgoing traffic on port 53"},
		{[]string{"-I", "OUTPUT", "-p", "tcp", "--dport", "53", "-j", "ACCEPT"}, "allow TCP outgoing traffic on port 53"},
		{[]string{"-I", "OUTPUT", "1", "-o", "lo", "-j", "ACCEPT"}, "Allow loopback traffic for resolve conf to work"},
		{[]string{"-I", "INPUT", "1", "-i", "lo", "-j", "ACCEPT"}, "Allow loopback traffic for resolve conf to work"},
	}); cmdErr != nil {
		return fmt.Errorf("could not execute commands: %w", cmdErr)
	}
	return nil
}
func (f *firewall) clearFirewall() error {
	if !f.enabled {
		return nil
	}
	if cmdErr := f.executeCommands([]struct {
		args []string
		desc string
	}{
		{[]string{"-F"}, "flush existing rules"},
		{[]string{"-P", "INPUT", "ACCEPT"}, "set default policy for INPUT"},
		{[]string{"-P", "FORWARD", "ACCEPT"}, "set default policy for FORWARD"},
		{[]string{"-P", "OUTPUT", "ACCEPT"}, "set default policy for OUTPUT"},
	}); cmdErr != nil {
		return fmt.Errorf("could not execute commands: %w", cmdErr)
	}
	return nil
}
func (f *firewall) executeCommands(iptablesCommands []struct {
	args []string
	desc string
}) error {
	for _, command := range iptablesCommands {
		cmd := exec.Command("iptables", command.args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error running iptables command to %s: %v, Output: %s", command.desc, err, output)
			return fmt.Errorf("could not %s: %w", command.desc, err)
		} else {
			// Uncomment for debugging
			//log.Printf("Successfully ran iptables command to %s, Output: %s", command.desc, output)
		}
	}
	return nil
}
