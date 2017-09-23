
.PHONY: all build

all: build

build:
	go build -o bin/rktup cli/main.go
