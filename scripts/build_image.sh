#! /bin/bash

if [ -f crypto-server.go ]
then
  printf "Please run script in the working directory of api/"
fi

docker build -t crypto-server-app .