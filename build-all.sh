#!/usr/bin/env bash

CGO_ENABLED=0  GOOS=linux GOARCH=amd64 go build -o BSC-linux-amd64 ./
CGO_ENABLED=0  GOOS=linux GOARCH=arm64 go build -o BSC-linux-arm64 ./
CGO_ENABLED=0  GOOS=darwin GOARCH=amd64  go build -o BSC-darwin-amd64 ./
CGO_ENABLED=0  GOOS=darwin GOARCH=arm64  go build -o BSC-darwin-arm64 ./
CGO_ENABLED=0  GOOS=windows GOARCH=amd64  go build -o BSC-windows-amd64.exe ./