.PHONY: build_linux_64 build_linux_arm_32 format lint test

build_linux_64:
	rm -f gphotos_download && go build

build_linux_arm_32:
	rm -f gphotos_download && CGO_LDFLAGS="-Xlinker -rpath=lib/synology-ds216j-libc-2.26 -static" CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go build

format:
	go fmt ./...
	gci -w .

lint:
	golangci-lint run

test:
	go test ./...