VERSION := 1.0.0-3
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
	@mkdir -p $(CONFIG_DIR)
	@mkdir -p $(SYSTEMD_DIR)
	@mkdir -p $(LOG_DIR)
	@touch $(CONFIG_DIR)/openvpn-credentials.txt
	@touch $(LOG_DIR)/$(PACKAGE_NAME).log
	@touch $(CONFIG_DIR)/openvpn-credentials.txt
	@cp -r vpnclient/openvpn/expressvpn/vpn_configs $(BUILD_DIR)/vpn_configs
	@cp -r pvs.service $(SYSTEMD_DIR)


copy_control_file:
	@cp $(DEBIAN_CONTROL_FILE_SRC) $(DEBIAN_PACKAGE_DIR)/DEBIAN/control

build_package:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o $(BUILD_DIR)/$(PACKAGE_NAME) ./main.go
	@dpkg --build $(DEBIAN_PACKAGE_DIR)
	@echo "Package built at $(DEBIAN_PACKAGE_DIR).deb"

clean:
	@rm -rf ./bin

