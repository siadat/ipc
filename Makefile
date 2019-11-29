test:
	go version
	go vet .
	go test -v .

clean:
	rm -rf ftok
	rm -rf ipcctl

ftok:
	gcc -o ftok c_ftok/ftok.c

ipcctl:
	go build ./cmd/ipcctl/

install:
	go install ./...

rm-queues:
	# ipcs -q | grep -Po "0x\S+" | sort | uniq | xargs -i ipcrm -Q {}
	ipcs -q | grep -Po "0x\S+" | sort | uniq | xargs -i ipcctl rm {}

	# Warning: the following will rm everything
	# ipcrm -a
	
# For cross arch building
arm-cross-build-ipcctl:
	GOARCH=arm GOOS=linux go build -o ipcctl-arm ./cmd/ipcctl/

arm-cross-build-test:
	# Required on non-Arm systems: apt-get install gcc-arm-linux-gnueabi
	CC=arm-linux-gnueabi-gcc GOARCH=arm GOOS=linux CGO_ENABLED=1 go test -count=1 -c

