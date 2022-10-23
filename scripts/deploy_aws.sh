#! /bin/bash

DockerImageName=""                  # image name
# shellcheck disable=SC2034
DockerContainerName=""              # container name
HOST=""                             # hostname/ip for ec2 instance
# shellcheck disable=SC2034
WORKDIR=""                          # workdir on ec2 instance
# shellcheck disable=SC2088
SSHKEY=""                           # location of private ssh key

if [ -z "$DockerImageName" ] || [ -z "$DockerContainerName" ] || [ -z "$HOST" ] || [ -z "$WORKDIR" ] || [ -z "$SSHKEY" ]
then
  printf "Please fill in variables to run script.\n"
  exit 1
fi

scp -i "$SSHKEY" docker-save/"$DockerImageName" \
  ec2-user@"$HOST":~/"$WORKDIR"

scp -i "$SSHKEY" scripts/run_container_aws.sh.real \
  ec2-user@$HOST:~/"$WORKDIR"

ssh -i "$SSHKEY" ec2-user@"$HOST" "cd "$WORKDIR" && \
    docker load < '$DockerImageName' && chmod +x run_container_aws.sh"

