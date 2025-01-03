#! /bin/bash
docker compose -f ./tests/docker-compose.yml build --no-cache
docker compose -f ./tests/docker-compose.yml up -d
HOST_2_LOOPBACK_IP=10.100.220.3
SOURCE_IP=10.200.100.2
docker compose -f ./tests/docker-compose.yml exec \
    -T host1 ping -c 5 $HOST_2_LOOPBACK_IP -I $SOURCE_IP

# docker-compose execの終了コードは実行したコマンド、
# ここでは`ping -c 5 $HOST_2_LOOPBACK_IP`のものである
# そのため、BGPでルートを交換し、pingが通れば0、それ以外の場合は1である
TEST_RESULT=$?
if [ $TEST_RESULT -eq 0 ]; then
    printf "\e[32m%s\e[m\n" "統合テストが成功しました"
else
    printf "\e[31m%s\e[m\n" "統合テストが失敗しました"
fi

docker container stop tests-host1-1 tests-host2-1
docker container rm tests-host1-1 tests-host2-1

exit $TEST_RESULT