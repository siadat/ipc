test:
	lscpu
	pwd
	ls -lash
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

rm-queues: ipcctl
	# ipcs -q | grep -Po "0x\S+" | sort | uniq | xargs -i ipcrm -Q {}
	ipcs -q | grep -Po "0x\S+" | sort | uniq | xargs -i ipcctl rm {}
	
