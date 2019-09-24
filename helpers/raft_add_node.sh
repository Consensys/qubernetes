#!/bin/sh

ENODE_URL=$1
PRIVATE_CONFIG=$TM_HOME/tm.ipc geth --exec "raft.addPeer(\"$ENODE_URL\")" attach ipc:$QUORUM_HOME/dd/geth.ipc;
