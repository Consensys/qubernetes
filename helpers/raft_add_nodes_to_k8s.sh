#!/bin/bash

#set -x

if [ "$#" -eq 2 ]; then
   echo "setting namespace to $2"
   NAMESPACE="--namespace=$2"
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

## apply the new  configs from the out dir
## but don't update the genesis file / config.
for f in out/*
do
	if [[ "$f" == *"genesis"* ]]; then
	   echo "skip reapplying genesis config"
  else
	   kubectl apply -f $f
	fi
done
echo
kubectl get pods

## Run raft.addNode on one connected node.
printf "${GREEN} Enter node/pod name of cluster node to run add node on, e.g. node1: ${NC} \n "
read POD_NAME

## TODO: could test the permissioned-nodes.sh to see when it changes.
echo "Giving configs maps 40 seconds to sync up."
sleep 40

POD=$(kubectl get pods $NAMESPACE | grep Running | grep $POD_NAME |  awk '{print $1}')
kubectl $NAMESPACE exec $POD -c quorum -- cat /etc/quorum/qdata/dd/permissioned-nodes.json
echo " done permission nodes"
echo "-----------------------"
echo
kubectl $NAMESPACE exec $POD -c quorum -- /etc/quorum/qdata/node-management/raft_add_all_permissioned.sh

