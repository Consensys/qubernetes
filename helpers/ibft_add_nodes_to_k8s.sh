#!/bin/bash


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

QUORUM_POD_PATTERN=quorum

printf "${GREEN} When all deployments are back up, hit any key: ${NC} \n"
printf "${GREEN} in another window run 'kubectl get -w pods' or 'watch kubectl get pods'"
read BACK_UP

# echo "update perssion nodes sh $QHOME/permission-nodes/permissioned-update.sh;"
echo "Step 2: proposing IBFT validators to the network."

PODS=$(kubectl get pods $NAMESPACE | grep $QUORUM_POD_PATTERN | grep Running | awk '{print $1}')
for POD in $PODS; do
  kubectl exec $POD -c quorum -- sh /etc/quorum/qdata/node-management/ibft_propose_all.sh
done

