#!/bin/bash
docker compose -f ./tests/docker-compose.yml build --no-cache
docker compose -f ./tests/docker-compose.yml up -d
docker compose -f ./tests/docker-compose.yml exec \
    -T host2 go test -v -timeout 1m ./...