#!/bin/bash
# ./addraftnode.sh quorum-node4 // interactive
# ./addraftnode.sh quorum-node4 quorum-node1 2.7.0 // run with released version
# ./addraftnode.sh quorum-node4 quorum-node1 quorum-raft-determ // run with local docker image.

if [[ "$#" -lt 1 ]]; then
  echo "./addraftnode.sh NODE_NAME"
  echo "./addraftnode.sh NODE_NAME NODE_NAME_RUN_RAFT_ADD QUORUM_IMAGE"
  exit 1
fi

NODE_NAME=$1

# Node to run raft.addpeer(enode) on, this will have a different effect if it is a <= 2.7 node vs a deterministic raft node.
NODE_NAME_RUN_RAFT_ADD=""
if [[ "$#" -gt 1 ]]; then
  NODE_NAME_RUN_RAFT_ADD=$2
fi
QUORUM_IMAGE=""
if [[ "$#" -eq 3 ]]; then
  QUORUM_IMAGE=$3
fi

qctl ls config
echo
qctl add node $NODE_NAME

# Generate the quorum resources first, as we need the enodeURL to try and add the node to the cluster.
qctl generate network --update
ENODE_URL=$(qctl ls node --enodeurl -b $NODE_NAME)

# Get the users input whether they wish to run the update on a certain node.
if [[ $NODE_NAME_RUN_RAFT_ADD == "" ]]; then
  printf "${GREEN} Enter node name to run raft.addPeer command on, e.g. quorum-node1 (default): ${NC} \n "
  read NODE_NAME_RUN_RAFT_ADD
  if [[ "$NODE_NAME_RUN_RAFT_ADD" == "" ]]; then
    NODE_NAME_RUN_RAFT_ADD="quorum-node1"
  fi
fi

# get the quorum image for the new node that we are deploying.
#  --qimagefull quorum-local-2.7
if [[ $QUORUM_IMAGE == "" ]]; then
  printf "${GREEN} Enter quorum image to use with the node [ quorum-local (default), quorum-local-2.7, quorum-raft-determ ] : ${NC} \n "
  read QUORUM_IMAGE
  echo "QUORUM_IMAGE: $QUORUM_IMAGE"

  if [[ "$QUORUM_IMAGE" == "" ]]; then
    QUORUM_IMAGE="quorum-local"
  fi
  echo
  echo
fi

# Try to get the RAFT_ID by adding the new node to the cluster.
echo qctl geth exec $NODE_NAME_RUN_RAFT_ADD "raft.addPeer(${ENODE_URL})"
RAFT_ADD_TO_CLUSTER_OUT=$(qctl geth exec $NODE_NAME_RUN_RAFT_ADD "raft.addPeer(${ENODE_URL})")
ALREADY_ADDED=$(echo $RAFT_ADD_TO_CLUSTER_OUT | grep "Error: node with this enode has already")

# if an error is returned when trying to get the RAFT_ID exit.
if [[ "$ALREADY_ADDED" != "" ]]; then
  echo
  echo "Error while trying to get the raft id"
  echo "Node [$NODE_NAME] with enode [$ENODE_URL]"
  echo "already in the cluster."
  echo
  echo  "> qctl ls nodes -b --name --enode"
  qctl ls nodes -b --name --enode
  exit 1
fi
echo "Response from raft.addPeer: $RAFT_ADD_TO_CLUSTER_OUT"
NEW_RAFT_ID=$(echo $RAFT_ADD_TO_CLUSTER_OUT | grep "[0-9]*" | sed 's/^"[0-9]*"//g' | sed 's/^.*pods//g' | sed 's/ //g' | sed $'s/[^[:print:]\t]//g' | sed 's/\[32m//g' | sed 's/\[0m//g' | sed 's/\[31m//g' | sed 's/"//g')
echo "NEW_RAFT_ID: $NEW_RAFT_ID"

if [[ "$QUORUM_IMAGE" == "2.7" ||  "$QUORUM_IMAGE" == "2.7.0" || "$QUORUM_IMAGE" == "2.6.0" ||
      "$QUORUM_IMAGE" == "2.5.0" || "$QUORUM_IMAGE" == "2.4.0" || "$QUORUM_IMAGE" == "2.3.0" ]]; then
  echo qctl update node --gethparams '--raftjoinexisting $NEW_RAFT_ID' $NODE_NAME
  qctl update node --gethparams "--raftjoinexisting $NEW_RAFT_ID" $NODE_NAME
else
  echo qctl update node --gethparams '--raftjoinexisting $NEW_RAFT_ID' --qimagefull $QUORUM_IMAGE $NODE_NAME
  qctl update node --gethparams "--raftjoinexisting $NEW_RAFT_ID" --qimagefull $QUORUM_IMAGE $NODE_NAME
fi
qctl generate network --update
qctl deploy network --wait

##echo qctl geth exec $NODE_NAME_RUN_CMD "'raft.addPeer($ENODE_URL)'"
##echo qctl geth exec $NODE_NAME_RUN_CMD "'raft.cluster'"
