#! /bin/bash

if [ ! -f crypto-server.go ]
then
  printf "Please run script in the working directory of api/"
  exit 1
fi

docker buildx build --platform linux/amd64 -t crypto-server-app .