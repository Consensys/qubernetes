#!/bin/bash

if [ "$#" -lt 2 ]; then
  echo " ./connect.sh $NODE_NUM $CONTAINER"
  echo "  example: "
  echo " ./connect.sh node2 quorum"
  exit 1
fi

if [ "$#" -eq 3 ]; then
   echo "setting namespace to $3"
   NAMESPACE="--namespace=$3"
fi

POD=$(kubectl get pods $NAMESPACE | grep Running | grep $1 |  awk '{print $1}')
echo "connecting to POD [$POD]"
kubectl $NAMESPACE exec -it $POD -c $2 /bin/ash 
