# Building the PVS binary
### Install Golang (Only required to build the pvs binary

```bash
wget https://go.dev/dl/go1.21.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.1.linux-amd64.tar.gz
sudo rm -rf go1.21.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
grep -qxF 'export PATH=$PATH:/usr/local/go/bin' /etc/profile || sudo sed -i '$aexport PATH=$PATH:/usr/local/go/bin' /etc/profile
```

#### Git clone the pvs project to a temp directory
```bash
git clone git@github.com:pearsonc/pvs.git /tmp/pvs-build
```

### Install the required packages
```bash
sudo apt install gcc dpkg-dev gpg -y
```

#### Build the pvs binary
This will build the pvs binary and package it into a deb file in the bin directory

```bash
cd /tmp/pvs-build
make build
```

# Setting up a ubuntu-server gateway with PVS

#### Download Ubuntu Server 20.04.2 LTS and install using the minimal installation option.

### Install updates and vim, curl, network tools and htop
```bash
sudo apt clean && sudo apt update && sudo apt upgrade -y \
&& sudo apt dist-upgrade -y && sudo apt autoremove -y \
&& sudo apt install curl vim htop net-tools -y
```
### Configure Network Interfaces
Edit the network configuration file using a text editor, the yaml filename may differ just ensure there is only one file in the directory.
It is important to note that your private network should have no dns or routes configured as it is a private network and the gateway will handle the routing.

```bash
sudo vi /etc/netplan/01-network-manager-all.yaml
```

Update the below config to meet your requirements

```yaml
network:
  ethernets:
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
    eth1:
      dhcp4: no
      addresses:
        - 192.168.1.2/24
  version: 2
```

Apply the changes
```bash
sudo netplan apply
```

### Enable IP Forwarding
Enable IP forwarding to allow gateway functionality and then make it persist after reboot.

```bash
sudo sysctl -w net.ipv4.ip_forward=1
echo "net.ipv4.ip_forward = 1" | sudo tee -a /etc/sysctl.conf
```

### Install the PVS deb file
```bash
sudo dpkg -i -f /tmp/pvs-build/bin/pvs_0.0.1_amd64.deb
```

### Add ExpressVPN credentials to the openvpn-credentials.txt file
```bash
sudo echo "vpnusername" | sudo tee /config/openvpn-credentials.txt
sudo echo "vpnpassword" | sudo tee -a /config/openvpn-credentials.txt
```

### Remove the last commands from the bash history to prevent the password from being stored in the bash history file
```bash
for i in {1..2}; do history -d $(($HISTCMD-1)); done
```

### Add the service to systemd and start it
```bash
sudo systemctl daemon-reload
sudo systemctl enable pvs.service
sudo systemctl start pvs.service
```

### A log file can be found at

```bash
tail -f /var/log/pvs.log
```