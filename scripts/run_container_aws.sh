#! /bin/bash

# shellcheck disable=SC2034
ContainerPort=""                # fill in for port mapping on container
# shellcheck disable=SC2034
ExternalPort=""                 # fill in for port mapping on container
DockerContainerName=""          # docker container name
DockerImageName=""              # docker image name

if [ -z "$ContainerPort" ] || [ -z "$ExternalPort" ] || [ -z "$DockerContainerName" ] || [ -z "$DockerImageName" ]
then
  printf "Please fill in variables to run script.\n"
  exit 1
fi

docker container stop "$DockerContainerName"
docker container rm "$DockerContainerName"
docker run -dit -p "$ExternalPort":"$ContainerPort" --platform linux/amd64 --env PORT="$ContainerPort" --name "$DockerContainerName" "$DockerImageName"