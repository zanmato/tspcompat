EXECUTABLE=tspcompat
WINDOWS=$(EXECUTABLE)_windows_amd64.exe
LINUX=$(EXECUTABLE)_linux_amd64
DARWIN=$(EXECUTABLE)_darwin_amd64

.PHONY: all test clean

all: test build

test:
	go test ./...

build: windows linux darwin

windows: $(WINDOWS)

linux: $(LINUX)

darwin: $(DARWIN)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -o bin/$(WINDOWS) -ldflags="-s -w" ./cmd/proxy/main.go

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -o bin/$(LINUX) -ldflags="-s -w" ./cmd/proxy/main.go

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -o bin/$(DARWIN) -ldflags="-s -w" ./cmd/proxy/main.go

clean:
	rm -rf bin/