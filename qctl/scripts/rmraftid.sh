#!/bin/bash
#set -xe
if [[ "$#" -lt 2 ]]; then
  echo "./removeraftid.sh RUN_ON_NODE raftId"
  exit 1
fi

NODE_NAME=$1
RAFT_ID=$2

IS_DETERM_NODE=$(qctl geth exec $NODE_NAME  "raft.cluster" | grep "raftId" | grep '"')

if [[ $IS_DETERM_NODE != "" ]]; then
  echo "is raft deterministic node"
  qctl geth exec $NODE_NAME "raft.removePeer(\"${RAFT_ID}\")"
else
  echo "is old node"
  ## try first removing from an old node
  qctl geth exec $NODE_NAME "raft.removePeer(${RAFT_ID})"
fi
