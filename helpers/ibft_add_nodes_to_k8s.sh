#!/bin/bash


if [ "$#" -eq 2 ]; then
   echo "setting namespace to $2"
   NAMESPACE="--namespace=$2"
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

## apply the new  configs from the out dir
## make sure the genesis didn't change.
kubectl apply -f out/00-quorum-persistent-volumes.yaml
kubectl apply -f out/02-quorum-shared-config.yaml
kubectl apply -f out/03-quorum-services.yaml
kubectl apply -f out/04-quorum-keyconfigs.yaml
echo
kubectl get pods

echo "Step 1: redeploy all pods for now."
#helpers/restart_deployments.sh
helpers/redeploy.sh
echo "  Waiting for pods to restart."
# this might take a while, so for now have the user check when all the pods
# are back up and prompt the script to continue.
kubectl get pods
printf "${GREEN} When all deployments are back up, hit any key: ${NC} \n"
printf "${GREEN} in another window run 'kubectl get -w pods' or 'watch kubectl get pods'"
read BACK_UP

# echo "update perssion nodes sh $QHOME/permission-nodes/permissioned-update.sh;"
echo "Step 2: proposing IBFT validators to the network."

POD_PATTERN=quorum
PODS=$(kubectl get pods $NAMESPACE | grep $POD_PATTERN | grep Running | awk '{print $1}')
for POD in $PODS; do
  kubectl exec $POD -c quorum /etc/quorum/qdata/node-management/ibft_propose_all.sh
done

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color
echo
printf " ${GREEN} Enter path to update deployment yaml \n"
printf " or <enter> to run all updates in the default \n"
printf " out/deployments/ directory: ${NC} \n"
read PATH_TO_DEPLOYMENT_YAML

if [ -z $PATH_TO_DEPLOYMENT_YAML ]; then
  kubectl apply -f out/deployments/
else
  kubectl apply -f $PATH_TO_DEPLOYMENT_YAML
fi

