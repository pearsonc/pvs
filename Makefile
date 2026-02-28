VERSION := 1.6.4
PACKAGE_NAME := pvs
DEBIAN_PACKAGE_DIR := bin/$(PACKAGE_NAME)_$(VERSION)_amd64
DEBIAN_CONTROL_FILE_SRC := package_metadata/control
BUILD_DIR := $(DEBIAN_PACKAGE_DIR)/usr/bin/$(PACKAGE_NAME)
CONFIG_DIR := $(DEBIAN_PACKAGE_DIR)/config
SYSTEMD_DIR := $(DEBIAN_PACKAGE_DIR)/etc/systemd/system
LOG_DIR := $(DEBIAN_PACKAGE_DIR)/var/log

# Main targets
.PHONY: all build setup_build_environment copy_control_file build_package clean

all: build

build: setup_build_environment copy_control_file build_package

setup_build_environment:
	@mkdir -p $(DEBIAN_PACKAGE_DIR)/DEBIAN
	@mkdir -p $(BUILD_DIR)
	@cp -r config.yml $(BUILD_DIR)
	@mkdir -p $(CONFIG_DIR)
	@mkdir -p $(SYSTEMD_DIR)
	@mkdir -p $(LOG_DIR)
	@mkdir -p $(BUILD_DIR)/expressvpn/vpn_configs
	@mkdir -p $(BUILD_DIR)/protonvpn/vpn_configs
	@touch $(CONFIG_DIR)/openvpn-credentials.txt
	@touch $(LOG_DIR)/$(PACKAGE_NAME).log
	@chmod 600 $(CONFIG_DIR)/openvpn-credentials.txt
	@# VPN configs (.ovpn) are deployed separately on the target — not bundled in .deb (C4 security fix)
	@if [ -d vpnclient/openvpn/expressvpn/vpn_configs ]; then cp -r vpnclient/openvpn/expressvpn/vpn_configs/* $(BUILD_DIR)/expressvpn/vpn_configs/ 2>/dev/null || true; fi
	@if [ -d vpnclient/openvpn/protonvpn/vpn_configs ]; then cp -r vpnclient/openvpn/protonvpn/vpn_configs/* $(BUILD_DIR)/protonvpn/vpn_configs/ 2>/dev/null || true; fi
	@cp -r pvs.service $(SYSTEMD_DIR)


copy_control_file:
	@cp $(DEBIAN_CONTROL_FILE_SRC) $(DEBIAN_PACKAGE_DIR)/DEBIAN/control
	@cp package_metadata/postinst $(DEBIAN_PACKAGE_DIR)/DEBIAN/postinst
	@chmod 755 $(DEBIAN_PACKAGE_DIR)/DEBIAN/postinst
	@mkdir -p $(DEBIAN_PACKAGE_DIR)/usr/share/doc/$(PACKAGE_NAME)
	@cp package_metadata/README $(DEBIAN_PACKAGE_DIR)/usr/share/doc/$(PACKAGE_NAME)/README

build_package:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(PACKAGE_NAME) ./main.go
	@dpkg-deb --root-owner-group --build $(DEBIAN_PACKAGE_DIR)
	@echo "Package built at $(DEBIAN_PACKAGE_DIR).deb"

clean:
	@rm -rf ./bin

