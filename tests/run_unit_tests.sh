#!/bin/bash
docker compose -f ./tests/docker-compose.yml build --no-cache
docker compose -f ./tests/docker-compose.yml up -d

# コンテナが完全に起動するまで停止
sleep 10

docker compose -f ./tests/docker-compose.yml exec \
    -T host2 go test -v -timeout 1m ./...

docker container stop tests-host1-1 tests-host2-1
docker container rm tests-host1-1 tests-host2-1
