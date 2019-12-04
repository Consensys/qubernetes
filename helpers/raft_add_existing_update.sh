#!/bin/bash
#set -ex
#echo "enode://c57f7113aa5b05392d1e33c0dee1795f3d5855fb84c442f27488bd64be2b8571ae7b131737ba50b8e6946de2dda0e60c380cdc270c750d7e88e9582e70274be7@10.11.242.133:30303?discport=0&raftport=50401" | awk -F {Print $1}
POD=$1
#POD=quorum-node1-deployment-785b45b775-xl9rs
RAFT_ID_ENODE_URL_CSV=$(kubectl exec -it $POD -c quorum cat /etc/quorum/qdata/contracts/raft_added.csv)

mkdir -p out/deployments/updates

# 5,"enode://6c46988942f8446c6fa6ae459d6369b01e44f5b4c7c6be6ef14bc14139305a1093fffad46be0482547d44c0353adef45b562458d6c0fe4cde95c3b0b0b8e2f12@10.11.240.120:30303?discport=0&raftport=50401"
for line in $RAFT_ID_ENODE_URL_CSV; do
  echo "$line"
  RAFT_ID=$(echo $line | awk -F , '{print $1}')
  echo
  echo " RaftID ${RAFT_ID}"
  if [ ! -z $RAFT_ID ]; then
    ENODE_URL=$(echo $line | awk -F , '{print $2}')
    echo " Enode_URL $ENODE_URL"
    ENODE_ID=$(echo $ENODE_URL | awk -F // '{print $2}' | awk -F @ '{print $1}')
    echo " ENODE_ID [$ENODE_ID]"
    echo
    DEPLOYMENT_FILE=$(grep -ls  $ENODE_ID out/deployments/{*,.*})
    echo "Deployment File: $DEPLOYMENT_FILE"
    UPDATE_FILE=out/deployments/updates/${RAFT_ID}_raft_join_deployment.yaml
    echo "UPDATEFILE: $UPDATE_FILE"
    sed "s/--raft .* --raftport /--raft --raftjoinexisting $RAFT_ID --raftport /g" $DEPLOYMENT_FILE > $UPDATE_FILE
  fi
done
