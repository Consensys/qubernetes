#!/bin/sh

set -x

# Read the permissioned-nodes.json (this will be redeployed by k8s)
# and add any new entry into the permissioned set.
PERMISSION_FILE=$QHOME/dd/permissioned-nodes.json
ENODE_URLS=$(cat $PERMISSION_FILE | jq '.[]')
RAFT_ADD_FILE=$QHOME/contracts/raft_add_$(date +%m-%d-%Y)
RAFT_ADD_LOG=$RAFT_ADD_FILE.log
RAFT_ADD_ERR=$RAFT_ADD_FILE.err
RAFT_ADD_FILE=$QHOME/contracts/raft_added.csv

touch $RAFT_ADD_LOG
touch $RAFT_ADD_ERR
date +%m-%d-%Y-%T >> $RAFT_ADD_ERR
date +%m-%d-%Y-%T >> $RAFT_ADD_LOG

echo "  Going through ENODE_URLS"
echo "  $ENODE_URLS"
echo
for URL in $ENODE_URLS; do

  # Check if the URL from the permissioned-nodes is this node, if so
  # don't add because it will cause an error.
  if echo $URL | grep -Eq $THIS_ENODE; then
    echo "skip adding self enodeID [$THIS_ENODE]"
    continue;
  fi

  RAFTID=$(PRIVATE_CONFIG=$TM_HOME/tm.ipc geth --exec "raft.addPeer($URL)" attach ipc:$QUORUM_HOME/dd/geth.ipc)

  # if the addPerr command isn't successful log the returned error and go to next ENODE_URL
  if echo $RAFTID | grep -Eiq ERR; then
    echo "RaftID Err: [$RAFTID]" >> $RAFT_ADD_ERR
    echo $RAFTID%%$URL >> $RAFT_ADD_ERR;
    continue;
  fi

  if echo $RAFTID | grep -Eq '[0-9][0-9]*'; then
    echo $RAFTID - $URL
    echo --raftjoinexisting $RAFTID
    echo "$RAFTID%%$URL" >> $RAFT_ADD_LOG;
    # holds all raft nodes added so far on this node.
    echo "$RAFTID,$URL" >> $RAFT_ADD_FILE;
  fi

done

echo | tee -a $RAFT_ADD_ERR $RAFT_ADD_LOG
echo ========================================= | tee -a $RAFT_ADD_ERR $RAFT_ADD_LOG
echo | tee -a $RAFT_ADD_ERR $RAFT_ADD_LOG
