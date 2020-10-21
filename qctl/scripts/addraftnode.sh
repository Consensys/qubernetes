#!/bin/bash

if [[ "$#" -lt 1 ]]; then
  echo "./addraftnode.sh NODENAME"
  exit 1
fi

NODE_NAME=$1
qctl ls config
echo
qctl add node $NODE_NAME

qctl generate network --update
ENODE_URL=$(qctl ls node --enodeurl -b $NODE_NAME)

# Get the users input whether they wish to run the update on a certain node.
printf "${GREEN} Enter node name to run raft.addPeer command on, e.g. quorum-node1 (default): ${NC} \n "
read NODE_NAME_RUN_CMD
if [[ "$NODE_NAME_RUN_CMD" == "" ]]; then
  NODE_NAME_RUN_CMD="quorum-node1"
fi

echo qctl geth exec $NODE_NAME_RUN_CMD "raft.addPeer(${ENODE_URL})"
RAFT_ADD_TO_CLUSTER_OUT=$(qctl geth exec $NODE_NAME_RUN_CMD "raft.addPeer(${ENODE_URL})")

## --qimagefull quorum-local-2.7
echo "Response from raft.addPeer: $RAFT_ADD_TO_CLUSTER_OUT"
NEW_RAFT_ID=$(echo $RAFT_ADD_TO_CLUSTER_OUT | grep "[0-9]*" | sed 's/^"[0-9]*"//g' | sed 's/^.*pods//g' | sed 's/ //g' | sed $'s/[^[:print:]\t]//g' | sed 's/\[32m//g' | sed 's/\[0m//g' | sed 's/\[31m//g' | sed 's/"//g')
echo "NEW_RAFT_ID: $NEW_RAFT_ID"

printf "${GREEN} Enter quorum image to use with the node [ quorum-local (default), quorum-local-2.7, quorum-raft-determ ] : ${NC} \n "
read QUORUM_IMAGE
echo "QUORUM_IMAGE: $QUORUM_IMAGE"

if [[ "$QUORUM_IMAGE" == "" ]]; then
  QUORUM_IMAGE="quorum-local"
fi
echo
echo

if [[ "$QUORUM_IMAGE" == "2.7" ||  "$QUORUM_IMAGE" == "2.7.0" || "$QUORUM_IMAGE" == "2.6.0" ||
      "$QUORUM_IMAGE" == "2.5.0" || "$QUORUM_IMAGE" == "2.4.0" || "$QUORUM_IMAGE" == "2.3.0" ]]; then
  echo qctl update node --gethparams '--raftjoinexisting $NEW_RAFT_ID' $NODE_NAME
  qctl update node --gethparams "--raftjoinexisting $NEW_RAFT_ID" $NODE_NAME
else
  echo qctl update node --gethparams '--raftjoinexisting $NEW_RAFT_ID' --qimagefull $QUORUM_IMAGE $NODE_NAME
  qctl update node --gethparams "--raftjoinexisting $NEW_RAFT_ID" --qimagefull $QUORUM_IMAGE $NODE_NAME
fi
qctl generate network --update
qctl deploy network

##echo qctl geth exec $NODE_NAME_RUN_CMD "'raft.addPeer($ENODE_URL)'"
##echo qctl geth exec $NODE_NAME_RUN_CMD "'raft.cluster'"
