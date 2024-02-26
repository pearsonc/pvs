# Setting up a ubuntu-server gateway and firewall

#### Download Ubuntu Server 20.04.2 LTS and install using the minimal installation option.

### Install updates and vim, curl and htop
You may need to first configure networking but if you have an internet connection start here if not jump to step 2.
```bash
sudo apt clean && sudo apt update && sudo apt upgrade -y \
&& sudo apt dist-upgrade -y && sudo apt autoremove -y \
&& sudo apt install curl vim htop net-tools wget git make openvpn iptables-persistent dnsutils iputils-ping -y
```
### Configure Network Interfaces
Edit the network configuration file using a text editor, the yaml filename may differ just ensure there is only one file in the directory.

```bash
sudo vi /etc/netplan/01-network-manager-all.yaml
```

Delete the content and replace with the below:

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
```bash
sudo netplan apply
```

### Enable IP Forwarding
Enable IP forwarding to allow gateway functionality and then make it persist after reboot.

```bash
sudo sysctl -w net.ipv4.ip_forward=1
echo "net.ipv4.ip_forward = 1" | sudo tee -a /etc/sysctl.conf
```


### Install Golang (Only required to build the pvs binary

```bash
wget https://go.dev/dl/go1.21.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.1.linux-amd64.tar.gz
sudo rm -rf go1.21.1.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
grep -qxF 'export PATH=$PATH:/usr/local/go/bin' /etc/profile || sudo sed -i '$aexport PATH=$PATH:/usr/local/go/bin' /etc/profile
```

### Install wd40vpn service

```bash
sudo groupadd pvs
sudo useradd -r -s /bin/false -g pvs pvs
```

#### Give the user sudo access to the openvpn command
```bash
sudo visudo
pvs ALL=(ALL) NOPASSWD: /usr/sbin/openvpn
```

#### Create the directory for the service
```bash
sudo mkdir -p /etc/pvs/
```

#### Git clone the homelab project to a temp directory
```bash
git clone git@github.com:pearsonc/pvs.git /tmp/pvs-build
```

#### build the wd40vpn binary and move it to the service directory along with the vpn configs and systemd service file
```bash
cd /tmp/pvs-build
make build
sudo cp bin/pvs /etc/pvs/pvs
sudo cp -r vpn-configs /etc/wd40vpn/vpn-configs
sudo cp pvs.service /etc/systemd/system/pvs.service
```

### Set permissions on the service directory
```bash
sudo chown -R pvs:pvs /etc/pvs
sudo find /etc/pvs -type d -exec chmod 750 {} \;
sudo find /etc/pvs -type f -exec chmod 640 {} \;
sudo chmod 750 /etc/pvs/pvs
```

### Add the service to systemd and start it
```bash
sudo systemctl daemon-reload
sudo systemctl enable pvs.service
sudo systemctl start pvs.service
```

### Delete Homelab repo
```bash
sudo rm -rf /tmp/pvs-build
```