#!/bin/bash

set_rules() {
    # Block all outgoing traffic first
    iptables -A OUTPUT -j DROP

    # Allow all traffic from eth1 to access the host
    iptables -I OUTPUT -o eth1 -j ACCEPT

    # Allow VPN connections over UDP for port 1195
    iptables -I OUTPUT -p udp --dport 1195 -j ACCEPT

    # Allow VPN connections over TCP for port 1195
    iptables -I OUTPUT -p tcp --dport 1195 -j ACCEPT

    # Allow DNS resolution using external DNS servers (UDP)
    iptables -I OUTPUT -p udp --dport 53 -j ACCEPT

    # Allow DNS resolution using external DNS servers (TCP)
    iptables -I OUTPUT -p tcp --dport 53 -j ACCEPT

    # Allow loopback traffic for resolve conf to work
    iptables -I OUTPUT 1 -o lo -j ACCEPT

    # Allow loopback traffic for resolve conf to work
    iptables -I INPUT 1 -i lo -j ACCEPT

    echo "Rules have been set."
}

reset_iptables() {
    # Flush all rules
    iptables -F

    # Set default policies to ACCEPT
    iptables -P INPUT ACCEPT
    iptables -P FORWARD ACCEPT
    iptables -P OUTPUT ACCEPT

    echo "iptables has been reset and everything is allowed."
}

show_rules() {
    # Display current iptables rules
    iptables -L -v -n

    echo "Above are the current iptables rules."
}

clear_firewall() {
    # Flush all rules
    iptables -F

    # Delete all custom user-defined chains
    iptables -X

    # Flush all rules in the NAT table
    iptables -t nat -F

    # Delete all custom user-defined chains in the NAT table
    iptables -t nat -X

    # Flush all rules in the mangle table
    iptables -t mangle -F

    # Delete all custom user-defined chains in the mangle table
    iptables -t mangle -X

    # Set default policy for INPUT chain to ACCEPT
    iptables -P INPUT ACCEPT

    # Set default policy for FORWARD chain to ACCEPT
    iptables -P FORWARD ACCEPT

    # Set default policy for OUTPUT chain to ACCEPT
    iptables -P OUTPUT ACCEPT

    echo "Firewall rules have been cleared and default policies set to ACCEPT."
}

# Check for user input
if [ "$1" == "set" ]; then
    set_rules
elif [ "$1" == "reset" ]; then
    reset_iptables
elif [ "$1" == "show" ]; then
    show_rules
elif [ "$1" == "clear" ]; then
    clear_firewall
else
    echo "Usage: $0 [set|reset|show|clear]"
    echo "  set   - Apply the specified rules"
    echo "  reset - Reset iptables and allow everything"
    echo "  show  - Display the current iptables rules"
    echo "  clear - Clear firewall rules and set default policies to ACCEPT"
fi