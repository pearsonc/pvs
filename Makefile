build:
	CGO_ENABLED=0 GOOS=linux go mod tidy
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/pvs ./main.go
	cp -r vpnclient/openvpn/expressvpn/vpn_configs /bin/vpn_configs
# On Deployment ensure /config/openvpn-credentials.txt is present with expressvpn credentials

clean:
	rm -rf ./bin

.PHONY: build clean
