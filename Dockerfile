FROM golang:1.20.6-alpine3.18 as builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/pvs.bin ./main.go

FROM alpine:latest

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN apk add --no-cache openvpn openresolv htop iptables

WORKDIR /app/

COPY --from=builder /app/bin/pvs.bin .
COPY --from=builder /app/vpnclient/openvpn/expressvpn/vpn_configs vpn_configs/
COPY --from=builder /app/openvpn-credentials.txt /config/openvpn-credentials.txt

RUN chmod 600 /config/openvpn-credentials.txt

EXPOSE 8080

CMD ["./pvs.bin"]