#!/usr/bin/env bash

GOOS=darwin GOARCH=amd64 go build -o sockv5er.darwin.amd64
GOOS=darwin GOARCH=arm64 go build -o sockv5er.darwin.arm64
GOOS=linux GOARCH=amd64 go build -o sockv5er.linux.amd64
GOOS=linux GOARCH=arm64 go build -o sockv5er.linux.arm64
GOOS=windows GOARCH=amd64 go build -o sockv5er.windows.amd64.exe
