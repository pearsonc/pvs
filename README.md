# 🛡️ PVS: The Automated VPN Rotator, Management, and Monitoring Service

**PVS (Pearson VPN Service)** is a robust daemon designed for Ubuntu-based systems that automatically manages, monitors, and rotates OpenVPN connections to enhance privacy and ensure consistent connectivity.

PVS currently supports **ExpressVPN** and **ProtonVPN**.

## ✨ Key Features
* **Connection Rotation:** Randomly selects and connects to a new VPN configuration file from a directory or a user-defined preferred list in the `config.yml` file.
* **Health Monitoring:** Performs a reliable TCP dial check to an external network target. If the check fails, PVS immediately attempts to rotate the connection and reinitialise the tunnel.
* **Gateway Firewall Protection:** Utilises IPTables to prevent local network traffic from leaking to the internet when the VPN connection is rotating, stops, or fails. 
* * **Note:** This feature is essential for gateway servers but must be disabled if PVS is running on a single-interface server to prevent from being locked out.
* **Flexible Deployment:** Runs as a systemd service on any Ubuntu-based system (Desktop or Server).

## 🚀 Installation & Quick Start
This release does not contain post-install scripts. Please follow these steps carefully to set up the PVS service.

**1. Download and Install the DEB Package**
```bash
# 1. Download the latest DEB file
wget https://github.com/pearsonc/pvs/releases/download/v1.6.0/pvs_1.6.0_amd64.deb

# 2. Install the package
# The -f flag attempts to correct potential dependency issues.
sudo apt-get install -f ./pvs_1.6.0_amd64.deb
```

**2. Add VPN Credentials Securely**
PVS uses a separate credentials file for OpenVPN login. This file must be created and protected with strict permissions.
```bash
# 1. Create the credentials directory
sudo mkdir -p /config

# 2. Add credentials to openvpn-credentials.txt
#    Replace 'vpn username' and 'vpn password' with your actual VPN login credentials
#    (Note: This is the VPN login, not your web login).
sudo echo "vpn username" | sudo tee /config/openvpn-credentials.txt
sudo echo "vpn password" | sudo tee -a /config/openvpn-credentials.txt

# 5. Remove the credential commands from bash history for security
for i in {1..2}; do history -d $(($HISTCMD-1)); done

# 4. Secure the credentials file (Owner Read/Write only)
sudo chmod 600 /config/openvpn-credentials.txt
```

**3. Update Configuration**
```bash
# Open and edit the main configuration file
sudo vi /usr/bin/pvs/config.yml
```

**Key Configuration Warnings:** 




| Scenario                                    | Setting                   | Action Required                                                                                                          |
|---------------------------------------------|---------------------------|--------------------------------------------------------------------------------------------------------------------------|
| **Single-interface server (e.g., VPS)**     | `enable_gateway_firewall` | Set to `false` to avoid being locked out.                                                                                |
| **Gateway server with multiple interfaces** | `enable_gateway_firewall` | Set to `true` to protect against leaks. Ensure `local LAN interface` and `subnet` are correct for your network topology. |

* **Rotation Interval:** Configure the rotation frequency using rotate_minutes.
* **Preferred Configs:** You can set a list |of preferred VPN configurations. By default, the list excludes UK locations.
* **VPN Configuration Files Locations:**
```bash
ls -la /usr/bin/pvs/expressvpn/vpn_configs
ls -la /usr/bin/pvs/protonvpn/vpn_configs
```

**4. Enable and Start the Service**
Reload the systemd daemon, enable the service to start on boot, and start it immediately.
```bash
sudo systemctl daemon-reload
sudo systemctl enable pvs.service
sudo systemctl start pvs.service
```

# 🔒 AppArmor Configuration (Ubuntu Desktop Users)
If you are running PVS on an Ubuntu desktop, you must configure AppArmor to allow the OpenVPN service to read the necessary files.

**Create the local AppArmor override file:**
```bash
sudo nano /etc/apparmor.d/local/openvpn
```
**Add the necessary read rules:**
```bash
# Allow read access to the credentials file
/config/openvpn-credentials.txt r,

# Allow read access to all config files in the provider directories
/usr/bin/pvs/protonvpn/vpn_configs/* r,
/usr/bin/pvs/expressvpn/vpn_configs/* r,
```
**Reload the AppArmor profiles:**
```bash
sudo apparmor_parser -r /etc/apparmor.d/openvpn
```

# ⚙️ Advanced: Setting up a PVS Gateway Server
This guide assumes you are starting with a minimal installation of Ubuntu Server with two network interfaces: one for the local LAN and one for the internet.

**1. Initial Prep and Tools**
```bash
# Update system and install required network and utility tools
sudo apt clean && sudo apt update && sudo apt upgrade -y \
&& sudo apt dist-upgrade -y && sudo apt autoremove -y \
&& sudo apt install curl vim net-tools -y
```

**2. Configure Network Interfaces (Netplan)**
```bash
# Edit the Netplan configuration file (filename may vary)
sudo vi /etc/netplan/01-network-manager-all.yaml
```
**Example Netplan Configuration:**
```yaml
network:
  ethernets:
    # eth0: WAN/Internet Interface (Example: Static IP, pointing to router via) Note your interfaces may differ
    eth0:
      addresses:
        - 192.168.0.11/24
      nameservers:
        addresses:
          - 1.1.1.1
          - 1.0.0.1
        search: []
      routes:
        - to: default
          via: 192.168.0.1
    # eth1: Private LAN Interface (No DNS or Routes configured here)
    eth1:
      dhcp4: no
      addresses:
        - 192.168.1.2/24
  version: 2
```
**Apply the Netplan Configuration:**
```bash
sudo netplan apply
``` 

**3. Enable IP Forwarding**

Enable IP forwarding to allow the server to route packets between the interfaces.

```bash
# Enable persistently across reboots
echo "net.ipv4.ip_forward = 1" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
``` 

**4. Final Setup**

Complete the installation using steps 1-4 from the Installation & Quick Start section above.

# 🪵 Viewing Logs
Monitor the service status and troubleshoot any connection or rotation issues by viewing the PVS log file, also you can use the system journal to help debug service issues.:
```bash
tail -f /var/log/pvs.log

sudo journalctl -u pvs.service -f

```
