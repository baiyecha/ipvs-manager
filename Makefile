PACKAGE_NAME := baiyecha/ipvs-manager

# compress golang binary size: https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/
build: install
	ls -al
	rm -rf artifacts
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o  release/ipvs-manager  $(PACKAGE_NAME)/cmd/api
	# scp ./release/ipvs-manager root@192.168.143.77:/root/tuzhigen/
	# scp ./release/ipvs-manager root@192.168.143.78:/root/tuzhigen/
	# scp ./release/ipvs-manager root@192.168.143.80:/root/tuzhigen/
	scp ./release/ipvs-manager root@192.168.143.228:/root/tuzhigen/

build-mac:
	ls -al
	rm -rf artifacts
	GO111MODULE=on CGO_ENABLED=0 go build -o  release/ipvs-manager  $(PACKAGE_NAME)/cmd/api

install:
	go mod tidy
