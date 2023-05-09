PACKAGE_NAME := baiyecha/ipvs-manager

# compress golang binary size: https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/
build: install
	ls -al
	rm -rf artifacts
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o  release/ipvs-manager  $(PACKAGE_NAME)/cmd/api
	scp ./release/ipvs-manager root@192.168.143.77:/root/tuzhigen/
	scp ./release/ipvs-manager root@192.168.143.78:/root/tuzhigen/
	scp ./release/ipvs-manager root@192.168.143.80:/root/tuzhigen/
	scp ./release/ipvs-manager root@192.168.143.228:/root/tuzhigen/

build-mac: install
	ls -al
	rm -rf artifacts
	GO111MODULE=on CGO_ENABLED=0 go build -o  release/ipvs-manager  $(PACKAGE_NAME)/cmd/api

install:
	go mod tidy
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative grpc/proto/*.proto
