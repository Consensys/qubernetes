#!/bin/sh

PERMISSION_FILE=$QHOME/dd/permissioned-nodes.json
ENODE_URLS=$(cat $PERMISSION_FILE | jq '.[]')
RAFT_ADD_FILE=$QHOME/contracts/raft_add_$(date +%m-%d-%Y)
RAFT_ADD_LOG=$RAFT_ADD_FILE.log
RAFT_ADD_ERR=$RAFT_ADD_FILE.err
RAFT_ADD_FILE=$QHOME/contracts/raft_addded.csv

touch $RAFT_ADD_LOG
touch $RAFT_ADD_ERR
date +%m-%d-%Y-%T >> $RAFT_ADD_ERR
date +%m-%d-%Y-%T >> $RAFT_ADD_LOG

echo loop
for URL in $ENODE_URLS; do

  if echo $URL | grep -Eq $THIS_ENODE; then
    echo "skip adding self enodeID [$THIS_ENODE]"
    continue;
  fi

  RAFTID=$(PRIVATE_CONFIG=$TM_HOME/tm.ipc geth --exec "raft.addPeer($URL)" attach ipc:$QUORUM_HOME/dd/geth.ipc)

  if echo $RAFTID | grep -Eiq ERR; then
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
