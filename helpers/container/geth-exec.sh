#!/bin/sh

# helper for connecting to geth from
# outside the container
# kubectl exec -it $POD -c quorum -- /geth-helpers/geth-exec.sh "admin.peers.length"

GETH_CMD="eth.blockNumber"
if [ "$#" -gt 0 ]; then
  GETH_CMD=$1
fi
# see: https://github.com/ethereum/go-ethereum/pull/17281
# https://github.com/ethereum/go-ethereum/issues/16905
# to avoid warning being returned
# "WARN [02-20|00:21:04.382] Sanitizing cache to Go's GC limits  provided=1024 updated=663"
geth --exec $GETH_CMD  --cache=16 attach --datadir $QUORUM_DATA_DIR $QUORUM_DATA_DIR/geth.ipc