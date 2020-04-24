#!/bin/ash
PRIVATE_CONFIG=$TM_HOME/tm.ipc geth --exec "loadScript(\"$1\")" attach --datadir $QUORUM_DATA_DIR ipc:$QUORUM_DATA_DIR/geth.ipc
