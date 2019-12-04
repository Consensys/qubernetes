#!/bin/bash

## apply a new deployment file based
## on the enode_id of the currently
## running pod(s).
function usage() {
  echo " ./redeploy.sh $POD_PATTERN_OPTIONAL"
  echo "  example: "
  echo " ./redeploy.sh node1"
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

PODS=$(kubectl get pod $NAMESPACE | grep $POD_PATTERN |  awk '{print $1}')
for POD in $PODS; do
  ENODE_ID=$(kubectl exec $POD -c quorum env | grep THIS_ENODE | awk -F= '{print $2}')
  DEPLOYMENT_FILE=$(grep -ls  $ENODE_ID out/deployments/{*,.*})
  echo "Redeploying enodeID $ENODE_ID, Deployment File: $DEPLOYMENT_FILE"
  kubectl apply -f $DEPLOYMENT_FILE
done
