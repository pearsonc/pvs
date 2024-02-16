# Start from a Debian-based Golang 1.20 image.
FROM golang:1.20.6-alpine3.18 as builder

# Set the current working directory inside the container.
WORKDIR /app

# Copy the local package files to the container's workspace.
COPY . .

# Download dependencies.
RUN go mod download

# Build the command inside the container.
# Notice the output file name is now crawler.bin
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/pvs.bin ./main.go


# Use a minimal alpine image.
FROM alpine:latest

# Install the required packages.
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN apk add --no-cache openvpn openresolv htop

# Set the working directory to /app/.
WORKDIR /app/

# Copy the binary and config from the `builder` image.
# Now it copies to /app/ and the binary is named crawler.bin
COPY --from=builder /app/bin/pvs.bin .
COPY --from=builder /app/vpnclient/expressvpn/vpn_configs vpn_configs/


# Copy credentials file to the container.
COPY --from=builder /app/openvpn-credentials.txt /config/openvpn-credentials.txt
RUN chmod 600 /config/openvpn-credentials.txt
#COPY --from=builder /app/config.yml .



# Expose port for the application.
EXPOSE 8080

# Run the binary.
CMD ["./pvs.bin"]