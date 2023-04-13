PACKAGE_NAME := baiyecha/ipvs-manager

# compress golang binary size: https://blog.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/
build: install
	ls -al
	rm -rf artifacts
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -o  release/ipvs-manager  $(PACKAGE_NAME)/cmd/api

build-mac:
	ls -al
	rm -rf artifacts
	GO111MODULE=on CGO_ENABLED=0 go build -o  release/ipvs-manager  $(PACKAGE_NAME)/cmd/api

install:
	go mod tidy
