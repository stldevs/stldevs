#!/bin/bash
docker run -p 5432:5432 --name stldevs-db --rm -e POSTGRES_PASSWORD=pw -d postgres
go test ./...
docker stop stldevs-db
