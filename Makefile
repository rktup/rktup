version="$(shell git describe --tags --always --dirty)"

.PHONY: all build

all: build templates.go

templates.go:
	go-bindata -pkg rktup -o templates.go \
		index.html ac-discovery.html

build: templates.go
	go build \
		-ldflags "-X github.com/rktup/rktup.Version=$(version)" \
		-o bin/rktup cli/main.go
