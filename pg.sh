#!/usr/bin/env sh

mkdir -p "$HOME"/postgres-data
docker run -d --rm --name dev-postgres \
-e POSTGRES_HOST_AUTH_METHOD=trust \
-v "$HOME"/postgres-data/:/var/lib/postgresql/data \
-p 127.0.0.1:5432:5432 \
postgres:13
