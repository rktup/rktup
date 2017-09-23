version="$(shell git describe --tags --always --dirty)"

.PHONY: all build

all: build

build:
	go build \
		-ldflags "-X github.com/schu/rktup.Version=$(version)" \
		-o bin/rktup cli/main.go
