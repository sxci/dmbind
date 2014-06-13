SHELL := /bin/bash

all: build

build:
	source env.sh
	rm -rf pkg
	go build -o dmbind run.go

run: build
	./dmbind
