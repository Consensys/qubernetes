#!/bin/sh

# helper for connecting to geth from
# outside the container
# kubectl exec -it $POD -c quorum /geth-helpers/geth-attach.sh
echo "connecting to geth $QHOME"
geth attach $QHOME/dd/geth.ipc