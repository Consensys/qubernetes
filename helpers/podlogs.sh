#!/bin/bash

#!/bin/bash

if [ "$#" -lt 2 ]; then
  echo " ./podlogs.sh $NODE_NUM $CONTAINER"
  echo "  example: "
  echo " ./podlogs.sh node2 quorum"
  exit 1
fi

POD=$(kubectl get pods | grep Running | grep $1 |  awk '{print $1}')
echo "Following logs for POD [$POD]"
kubectl logs -f $POD $2