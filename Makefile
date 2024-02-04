build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/pvs.bin ./main.go

clean:
	rm -rf ./bin

.PHONY: build clean
