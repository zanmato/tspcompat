WINDOWS=windows_amd64.exe
LINUX=linux_amd64
DARWIN=darwin_amd64

.PHONY: all test clean

all: test build

test:
	go test ./...

build: windows linux darwin

windows: $(WINDOWS)

linux: $(LINUX)

darwin: $(DARWIN)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -o bin/tspcompatproxy_$(WINDOWS) -ldflags="-s -w" ./cmd/proxy/main.go
	env GOOS=windows GOARCH=amd64 go build -o bin/tspcompatapi_$(WINDOWS) -ldflags="-s -w" ./cmd/api/main.go

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -o bin/tspcompatproxy_$(LINUX) -ldflags="-s -w" ./cmd/proxy/main.go
	env GOOS=linux GOARCH=amd64 go build -o bin/tspcompatapi_$(LINUX) -ldflags="-s -w" ./cmd/api/main.go

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -o bin/tspcompatproxy_$(DARWIN) -ldflags="-s -w" ./cmd/proxy/main.go
	env GOOS=darwin GOARCH=amd64 go build -o bin/tspcompatapi_$(DARWIN) -ldflags="-s -w" ./cmd/api/main.go

clean:
	rm -rf bin/