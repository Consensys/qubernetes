#!/bin/bash

function usage() {
  echo " ./quorum-build.sh IMAGE_NAME"
  echo "  example: ./quorum-build.sh "
  echo "  example: ./quorum-build.sh quorum-test-ibft"
}

if [[ $# -gt 1 ]];
then
  usage
fi

IMAGE_NAME=quorum-local

if [[ $# -eq 1 ]];
then
  IMAGE_NAME=$1
fi

eval $(minikube docker-env)
echo "building quorum image $IMAGE_NAME"

docker build -t $IMAGE_NAME -f Dockerfile.quorum .