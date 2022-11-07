#! /bin/bash

# shellcheck disable=SC2034
ContainerPort=           # fill in for port mapping on container
# shellcheck disable=SC2034
ExternalPort=            # fill in for port mapping on container
DockerContainerName=          # docker container name
DockerImageName=              # docker image name

if [ -z "$ContainerPort" ] || [ -z "$ExternalPort" ] || [ -z "$DockerContainerName" ] || [ -z "$DockerImageName" ]
then
  printf "Please fill in variables to run script.\n"
  exit 1
fi

if [ ! -f .env ]
then
  printf ".env file could not be found in the project directory."
fi

docker container stop "$DockerContainerName"
docker container rm "$DockerContainerName"

docker run -it -p "$ExternalPort":"$ContainerPort" --env-file .env --name "$DockerContainerName" "$DockerImageName"