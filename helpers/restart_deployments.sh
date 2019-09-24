#!/bin/bash

function usage() {
  echo " ./restart_deployments.sh $POD_PATTERN_OPTIONAL"
  echo "  example: "
  echo " ./restart_deployments.sh quorum"
  exit 1
}

POD_PATTERN="quorum"
if [ ! -z $1 ]; then
    POD_PATTERN=$1
fi

if [ "$#" -eq 2 ]; then
   echo "setting namespace to $2"
   NAMESPACE="--namespace=$2"
fi

DEPLOYMENTS=$(kubectl get deployments $NAMESPACE | grep $POD_PATTERN |  awk '{print $1}')
for DEP in $DEPLOYMENTS; do
  echo "scaling down [$DEP]"
  kubectl scale deployment $DEP --replicas=0 $NAMESPACE 
  echo "scaling back up [$DEP]"
  kubectl scale deployment $DEP --replicas=1 $NAMESPACE 
done
