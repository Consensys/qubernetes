#!/bin/bash

#set -x

function usage() {
  echo " ./add_nodes_to_k8s.sh $CONSENSUS"
  echo "  example: "
  echo " ./add_nodes_to_k8s.sh raft"
  echo " ./add_nodes_to_k8s.sh istanbul"
  exit 1
}

if [[ "$#" -lt 1 ]]; then
  usage
fi

CONSENSUS=$1
CONSENSUS=$(echo $CONSENSUS | awk '{ print toupper($0) }')
echo  $CONSENSUS

if [ "$#" -eq 2 ]; then
   echo "setting namespace to $2"
   NAMESPACE="--namespace=$2"
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

QUORUM_POD_PATTERN=quorum

## apply the new  configs from the out dir
## but don't update the genesis file / config.
for f in out/*
do
	if [[ "$f" == *"genesis"* ]]; then
	   echo "skip reapplying genesis config"
	## only apply the yaml files, skipping to the configs, to avoid error output
	elif [[ "$f" == *"yaml"* ||  "$f" == *"yml"* ]]; then
	  kubectl apply -f $f
  else
	  echo "skipping $f "
	fi
done

## Deploy the new new nodes which should be inthe deployments directory. The old nodes should remain unchanged.
kubectl apply -f out/deployments

echo
kubectl get pods

if [[ $CONSENSUS == "RAFT" ]]; then

  ## Run `raft.addNode(enode)` on one connected node.
  printf "${GREEN} Enter node/pod name of cluster node to run add node on, e.g. node1: ${NC} \n "
  read POD_NAME
  POD=$(kubectl get pods $NAMESPACE | grep Running | grep $POD_NAME |  awk '{print $1}')

  PERMISSION_FILE=$(kubectl $NAMESPACE exec $POD -c quorum -- cat /etc/quorum/qdata/dd/permissioned-nodes.json)
  echo $PERMISSION_FILE
  CUR_PERMISSION_FILE=$PERMISSION_FILE

  CT=0
  # wait a max of 120 seconds (MAX_ATTEMPTS * sleep 5), if the file doesn't change, try to run the update anyways, as maybe the user too a long time
  # to enter the node in the previous step.
  MAX_ATTEMPTS=24
  while [[ "$PERMISSION_FILE" == "$CUR_PERMISSION_FILE" && "$CT" -lt "$MAX_ATTEMPTS" ]]; do
    sleep 5
    echo  "${CT} out of ${MAX_ATTEMPTS} attempts"
    ((CT=CT+1))
    CUR_PERMISSION_FILE=$(kubectl $NAMESPACE exec $POD -c quorum -- cat /etc/quorum/qdata/dd/permissioned-nodes.json)
    echo "permissioned-nodes.json: $CUR_PERMISSION_FILE"
  done

  # try to run raft.addPeer for every node in the permissioned-nodes.json file, nodes that are already in the cluster
  # will display an error, but this error is harmless.
  kubectl $NAMESPACE exec $POD -c quorum -- /etc/quorum/qdata/node-management/raft_add_all_permissioned.sh

elif [[ $CONSENSUS == "ISTANBUL" || $CONSENSUS == "IBFT" ]]; then

  printf "${GREEN} Do you wish to promote all new nodes to be istanbul validators? [Y/N] ${NC} \n"
  read RESP
  RESP=$(echo $RESP | awk '{ print toupper($0) }')

  if [[ $RESP == "Y" || $RESP == "YES" ]]; then
    ## TODO: can add a test to see when all the nodes are up and also to query one healthy node to make sure the config map has been updated.
    echo "Waiting for 30 seconds to allow the configMaps to update"
    # echo "update permissioned-nodes.json sh $QHOME/permission-nodes/permissioned-update.sh;"
    echo "Step 2: proposing IBFT validators to the network."

    PODS=$(kubectl get pods $NAMESPACE | grep $QUORUM_POD_PATTERN | grep Running | awk '{print $1}')
    # Add all nodes in $QHOME/istanbul-validator-config.toml/istanbul-validator-config.toml as validators.
    #  kubectl exec $POD -c quorum -- cat /etc/quorum/qdata/istanbul-validator-config.toml/istanbul-validator-config.toml
    for POD in $PODS; do
      kubectl $NAMESPACE exec $POD -c quorum -- sh /etc/quorum/qdata/node-management/ibft_propose_all.sh
    done
  fi

fi

