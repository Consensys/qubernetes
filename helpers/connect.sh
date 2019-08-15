#!/bin/bash

if [ "$#" -lt 2 ]; then
  echo " ./connect.sh $NODE_NUM $CONTAINER"
  echo "  example: "
  echo " ./connect.sh node2 quorum"
  exit 1
fi

POD=$(kubectl get pods | grep Running | grep $1 |  awk '{print $1}')
echo "connecting to POD [$POD]"
kubectl exec -it $POD -c $2 /bin/ash 
