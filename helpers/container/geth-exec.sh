#!/bin/sh

# helper for connecting to geth from
# outside the container
# kubectl exec -it $POD -c quorum /geth-helpers/geth-exec.sh "admin.peers.length"

GETH_CMD="eth.blockNumber"
if [ "$#" -gt 0 ]; then
  GETH_CMD=$1
fi
geth --exec $GETH_CMD  attach $QHOME/dd/geth.ipc