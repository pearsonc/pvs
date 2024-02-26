build:
	CGO_ENABLED=0 GOOS=linux go mod tidy
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/pvs ./main.go
# On Deployment ensure /config/openvpn-credentials.txt is present with expressvpn credentials

clean:
	rm -rf ./bin

.PHONY: build clean
