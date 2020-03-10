#!/bin/bash

if [ "$#" -lt 1 ]; then
  echo " ./run_contract.sh $NODE_NUM [pub | priv | both]"
  echo "  example: "
  echo " ./run_contract.sh node2 priv"
  exit 1
fi

# priv, pub, both
CONTRACT_TYPE="both"
if [ "$#" -gt 1 ]; then
   echo "contract to run: $2"
   CONTRACT_TYPE="$2"
fi

if [ "$#" -eq 3 ]; then
   echo "setting namespace to $3"
   NAMESPACE="--namespace=$3"
fi

POD=$(kubectl get pods $NAMESPACE | grep Running | grep $1 |  awk '{print $1}')
echo "connecting to POD [$POD]"

if [ "$CONTRACT_TYPE" == "priv" ] || [ "$CONTRACT_TYPE" == "both" ]; then
  echo "running private contract"
  kubectl $NAMESPACE exec -it $POD -c quorum /etc/quorum/qdata/contracts/runscript.sh /etc/quorum/qdata/contracts/private_contract.js
fi

if [ "$CONTRACT_TYPE" == "pub" ] || [ "$CONTRACT_TYPE" == "both" ]; then
  echo "running public contract"
  kubectl $NAMESPACE exec -it $POD -c quorum /etc/quorum/qdata/contracts/runscript.sh /etc/quorum/qdata/contracts/public_contract.js
fi
