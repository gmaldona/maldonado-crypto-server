#! /bin/bash

docker stop crypto-server-app
docker rm crypto-server-app
docker run -it --name crypto-server-app crypto-server-app