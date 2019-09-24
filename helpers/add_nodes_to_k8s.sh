#!/bin/bash


if [ "$#" -eq 2 ]; then
   echo "setting namespace to $2"
   NAMESPACE="--namespace=$2"
fi

echo " Run full restart / redeploy [Y/N]"
read FULL_REDEPLOY
if [ "$FULL_REDEPLOY" = "Y" ] || [ "$FULL_REDEPLOY" = "y" ]; then
  ## apply the new  configs from the out dir
  ## make sure the genesis didn't change.
  kubectl apply -f out/00-quorum-persistent-volumes.yaml
  kubectl apply -f out/02-quorum-shared-config.yaml
  kubectl apply -f out/03-quorum-services.yaml
  kubectl apply -f out/04-quorum-keyconfigs.yaml
  echo
  kubectl get pods
  echo
fi

echo " Enter pod name to run update on (once pods are running again): "
read POD_NAME

helpers/restart_deployments.sh $POD_NAME
echo "  waiting for pod to restart."

POD=""
while [ -z $POD ]; do
    ## wait for the cluster to come back up.
    POD=$(kubectl get pods $NAMESPACE | grep Running | grep $POD_NAME |  awk '{print $1}')
    sleep 5
    echo " waiting for 5..."
done

echo "running update on POD [$POD] but first giving it 20 more seconds to startup"
sleep 20

kubectl exec $POD -c quorum /etc/quorum/qdata/contracts/raft_add_all_permissioned.sh
helpers/raft_add_existing_update.sh $POD

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color
echo
printf " ${GREEN} Enter path to update deployment yaml \n"
printf " or <enter> to run all updates in the default \n"
printf " out/deployments/updates/ directory: ${NC} \n"
read PATH_TO_DEPLOYMENT_YAML

if [ -z $PATH_TO_DEPLOYMENT_YAML ]; then
  kubectl apply -f out/deployments/updates/
else
  kubectl apply -f $PATH_TO_DEPLOYMENT_YAML
fi

