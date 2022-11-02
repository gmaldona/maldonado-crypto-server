#! /bin/bash

if [ ! -f crypto-server.go ]
then
  printf "Please run script in the working directory of api/"
  exit 1
fi



while getopts 'lph' OPTION; do
  case "$OPTION" in
    l)
      printf "Building image for local machine.\n"
      docker image rm maldonado-crypto-server-image-local
      docker build -t maldonado-crypto-server-image-local .
      ;;
    p)
      printf "Building image for AWS linux machine.\n"
      docker image rm maldonado-crypto-server-image
      docker buildx build --platform linux/amd64 -t maldonado-crypto-server-image .
      ;;
    h)
      printf "\055l\t Build a docker image for local machine.\n"
      printf "\055p\t Build a docker image for AWS linux machine.\n"
      ;;
    ?)
      printf "Building image for AWS linux machine.\n"
      docker image rm maldonado-crypto-server-image
      docker buildx build --platform linux/amd64 -t maldonado-crypto-server-image .
  ;;
esac
done


