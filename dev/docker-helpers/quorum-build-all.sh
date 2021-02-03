#!/bin/bash

function usage() {
  echo " ./quorum-build-all.sh IMAGE_NAME"
  echo "  example: ./quorum-build-all.sh "
  echo "  example: ./quorum-build-all.sh quorum-test-ibft"
}

echo "$#"
if [[ "$#" -gt 1 ]];
then
  usage
  exit 1
fi

IMAGE_NAME=quorum-local

if [[ "$#" -eq 1 ]];
then
  IMAGE_NAME=$1
fi

echo "building quorum image $IMAGE_NAME"

eval $(minikube docker-env)
docker build -t quorum-base-local -f Dockerfile.quorumbase . && docker build -t $IMAGE_NAME -f Dockerfile.quorum .