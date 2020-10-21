#!/bin/bash

# Deletion requires:
# * removing the raftId from the cluster
# * removing the PVC and Deployment associated with the nodes.
# * removing the keys / enode id, TODO: we should be able to only regenerate the enodeid.
#  $>  ./deletenode.sh $QUBE_K8S_DIR quorum-node4
if [[ "$#" -lt 1 ]]; then
  #echo "./deletenode.sh QUBE_K8S_DIR NODENAME "
  echo "./deletenode.sh NODE_NAME"
  exit 1
fi

if [[  -z $QUBE_K8S_DIR ]]; then
  echo "please set $QUBE_K8S_DIR to you K8s out directory."
  echo "  export QUBE_K8S_DIR=/PATH/TO/K8s/out"
  echo
fi
echo "using QUBE_K8S_DIR $QUBE_K8S_DIR"

NODE_NAME=$1
# node < 2.7
TYPE="old"

## check if the node is a new deterministic node
IS_DETERM_NODE=$(qctl geth exec $NODE_NAME  "raft.cluster" | grep "raftId" | grep '"')
if [[ $IS_DETERM_NODE != "" ]]; then
  echo "is deterministic node"
  TYPE="deterministic"
else
  echo "old node"
fi

## Obtain the RAFT_ID associated to the node from the cluster.
RAFT_ID=$(qctl geth exec $NODE_NAME "raft.cluster" | grep -A 4 $NODE_NAME | grep raftId | sed "s/raftId://g" | sed "s/\"//g" | sed 's/,//g' | sed 's/ //g')

if [[ $TYPE == "old" ]]; then
  echo "remove 2.7 or before node from cluster"
  ## remove hidden chars
  RAFT_ID=$(echo $RAFT_ID | sed $'s/[^[:print:]\t]//g' | sed 's/\[32m//g' | sed 's/\[0m//g' | sed 's/\[//g' | sed 's/31m//g' | sed 's/\]//g')
  echo "raft id [$RAFT_ID]"
  qctl geth exec $NODE_NAME "raft.removePeer(${RAFT_ID})"
else
  echo "remove deterministic raft node from cluster"
  ## remove hidden chars
  RAFT_ID=$(echo $RAFT_ID | sed $'s/[^[:print:]\t]//g' | sed 's/\[32m//g' | sed 's/\[0m//g')
  echo "raft id [$RAFT_ID]"
  qctl geth exec $NODE_NAME "raft.removePeer(\"${RAFT_ID}\")"
fi

## remove from the K8s cluster
echo qctl delete node --hard $NODE_NAME
qctl delete node --hard $NODE_NAME

#rm $K8S_DIR/config/key-${NODE_NAME}/*
#rmdir $K8S_DIR/config/key-${NODE_NAME}

# Generate a new enode id
#cd $K8S_DIR/config/key-${NODE_NAME}/
#bootnode -genkey nodekey
#bootnode  -nodekeyhex $(cat nodekey) -writeaddress > enode

#kubectl delete deployment ${NODE_NAME}-deployment
#kubectl delete pvc ${NODE_NAME}-pvc
