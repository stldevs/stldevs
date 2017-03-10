#!/usr/bin/env bash

env GOOS=linux GOARCH=386 go build cmd/stldevs/stldevs.go
docker-compose build
rm stldevs
docker-compose up
