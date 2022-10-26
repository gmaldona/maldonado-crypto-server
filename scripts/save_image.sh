#! /bin/bash

if [ ! -f crypto-server.go ]
then
  printf "Please run script in the working directory of server"
  exit 1
fi

if [ ! -d docker-save ]
then
    mkdir docker-save
fi

 docker save maldonado-crypto-server-image > ./docker-save/maldonado-crypto-server.tar