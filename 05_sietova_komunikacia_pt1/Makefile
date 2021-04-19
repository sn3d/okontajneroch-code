all: rootfs build

rootfs:
	mkdir rootfs
	wget https://dl-cdn.alpinelinux.org/alpine/v3.13/releases/x86_64/alpine-minirootfs-3.13.2-x86_64.tar.gz -O /tmp/alpine.tar.gz
	tar xvf /tmp/alpine.tar.gz -C ./rootfs

build: 
	go build -o anton main.go

clean:
	rm -f ./anton
	rm -rf ./rootfs
