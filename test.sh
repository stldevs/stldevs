#!/bin/bash
set -e
docker run -p 5432:5432 --name stldevs-db --rm -e POSTGRES_PASSWORD=pw -d postgres
sleep 5
go test ./...
docker stop stldevs-db
