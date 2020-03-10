#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

function usage() {
  echo " ./test-qnet.sh {NAMESPACE}"
  echo "  if no namespace is given, the default namespace is used."
  echo "  example: ./test-qnet.sh quorum-namespace"
  echo "  example: ./test-qnet.sh"
}

# get_block_number NAMESPACE
function get_block_number() {
  NAMESPACE=$1
  kubectl exec -it $POD -c quorum /geth-helpers/geth-exec.sh "eth.blockNumber" $NAMESPACE
  BLOCK_NUM=$(kubectl exec -it $POD -c quorum /geth-helpers/geth-exec.sh "eth.blockNumber" $NAMESPACE)
  echo $BLOCK_NUM > block.tmp
  # FIXME: this is really annoying, but removes the color return from the block number.
  BLOCK_NUM=$(cat -v block.tmp | sed 's/\^M//g' | sed  's|0m||g' | sed 's|31m||g' | sed 's|\^||g' | sed 's|\[||g')
  rm block.tmp
  return $BLOCK_NUM
}

# This script will wait for a Quorum network to come up up the kubernetes cluster,
# and then test that commands can be run against the network / cluster.
# Assumes that the kubernetes resources have already been applied to a running kuberentes service, and that kubectl
# has been set to run against kubernetes services.
# ./test-qnet.sh quorum-namespace
if [[ "$#" -lt 1 ]]; then
   echo "no namespace passed in, using default namespace or namespace set in the current-context"
elif [[ "$#" -eq 1 ]]; then
  echo "setting namespace to $1"
  NAMESPACE="--namespace=$1"
  NAMESPACE_NAME=$1
else
  usage
  exit 1
fi

# make sure all the pods come up and are in a RUNNING state.
ALL_RUNNING="false"
while [[ ${ALL_RUNNING} != "true" ]]
do
  RUNNING="true"
  # PODS is POD_NAME and POD_STATUS (Running | Pending, etc) NAME%%STATUS
  # get all pods NAME%%STATUS in order to test if all pods are running yet.
  # Set up the returned pods so we can loop through them [NAME%%STATUS].
  PODS_NAME_STATUS=$(kubectl get pods $NAMESPACE | grep quorum | awk '{print $1"%%"$3}')
  # echo "PODS_NAME_STATUS: [$PODS_NAME_STATUS]"

  # if there are no pods returned, this may be because the kuberentes backend is taking a bit
  # longer to initialize the PODs, wait for a few loops:
  # TODO: terminate after several attempts with a failure code.
  if [[ -z "$PODS_NAME_STATUS" ]]; then
    printf "${RED}Pods are not up yet, wait 2 seconds before trying again.${NC}\n"
    sleep 2
    continue;
  fi

  # go through the [NAME%%STATUS] list and check that all the Pods in the list are in a RUNNING state.
  for P in ${PODS_NAME_STATUS}
  do
    POD_NAME=$(echo "$P" |  awk -F '%%' '{print $1}')
    STATUS=$(echo "$P" |  awk -F '%%' '{print $2}')
    # echo "name [$POD_NAME] : status [$STATUS]"
    if [[ ${STATUS} != "Running" ]]; then
      echo "pod $POD_NAME is ${STATUS} != RUNNING"
      RUNNING="false"
      break
    fi
  done
  # RUNNING should be set to false if any pod does not have the RUNNING status
  if [[ $RUNNING == "true" ]]; then
     ALL_RUNNING="true"
     break
  fi
  echo "Waiting for all PODs to start up."
  echo "ALL_RUNNING == ${ALL_RUNNING}"
  sleep 5
done

printf "${GREEN}OK all pods up and ready for action! ${NC}\n"
#echo "kubectl get pods $NAMESPACE | grep quorum | grep node1 | awk '{print $1}'"
POD=$(kubectl get pods $NAMESPACE | grep quorum | grep node1 | awk '{print $1}')
printf "${GREEN}Running on pod $POD ${NC}\n"

EXIT_CODE=1;
while [ ${EXIT_CODE} -ne 0 ]; do
  echo kubectl exec -it $POD -c quorum /geth-helpers/geth-exec.sh "eth.blockNumber" $NAMESPACE
  kubectl exec -it $POD -c quorum /geth-helpers/geth-exec.sh "eth.blockNumber" $NAMESPACE
  EXIT_CODE=$?
  echo "EXIT_CODE IS $EXIT_CODE"
  sleep 2
done

# BLOCK_NUM will only instantly go up if raft
# If istanbul it may take 5-10 seconds to start minting.
ISTANBUL=$(echo $NAMESPACE | grep -i istan)

## TODO: loop here, sometimes it takes a minute!
# Istanbul has a condition where it may show block 1 for a while, then jump to a higher number, so if the consesus
# is istanbul loop through until the block starts to increments to 2.
BLOCK_NUM=0
if [[ ! -z $ISTANBUL ]]; then
  printf "${RED}istanbul consensus!${NC}\n"
  CONTINUE=0
  while [[ $CONTINUE -ne 1 ]];
  do
    get_block_number $NAMESPACE
    BLOCK_NUM=$?
    if [[ $BLOCK_NUM -le 2 ]]; then
       printf "${RED}istanbul consensus, waiting for blocks to start to increase.${NC}\n"
       sleep 5
    else
       CONTINUE=1
    fi
  done
fi

# raft sleep 5 just to make sure nodes sync up.
## TODO: fails qubernetes-raft-tessera, because 1. raft doesn't sync up right away, and 2. tessera fails without
## returning an error. err creating contract Error: Non-200 status code
sleep 5

# Run some transactions:
# All PODS should now be RUNNING, and geth process is started and able to return the blockNumber.
# test a public and private transaction on a designated Node ${NODE_TO_TEST}.
NODE_TO_TEST=node1
echo "Testing a public transaction"
echo helpers/run_contracts.sh $NODE_TO_TEST pub $NAMESPACE_NAME
EXIT_CODE=1
helpers/run_contracts.sh $NODE_TO_TEST pub $NAMESPACE_NAME
EXIT_CODE=$?

if [[ $EXIT_CODE -ne 0 ]];
then
  printf "${RED}ERROR: public transaction failed!${NC}\n"
  exit 1
fi

## Test blockNumber for Public TX, should be > 1
get_block_number $NAMESPACE
BLOCK_NUM=$?
echo "BLOCK_NUM: [$BLOCK_NUM]"

## If raft the block numbers should be at least 1, if istanbul we are not sure, it is > 2.
if [[ "$BLOCK_NUM" -ge 1 ]];
then
   printf "${GREEN}SUCCESS: successfully executed public  transactions!! ${NC}\n"
else
   printf "${RED}ERROR: executing public transactions!! ${NC}\n"
   exit 1
fi

# The private transaction may take longer than the public transacion due to the transaction manager starting up and syncing
# up, especially when using tessera which takes a while to boot up and synchronize.
# try and wait for the private contract to execute without an error, e.g. in the case where
# it hasn't started up completely.
EXIT_CODE=1
MAX_TRIES=10
CT=0
# if tessera may take longer, and will not throw an error exit code
# err creating contract Error Non-200 status code: &{Status:404 Not Found StatusCode:404 Proto:HTTP/1.1 ProtoMajor:1 ProtoMinor:1
while [[ $EXIT_CODE -ne 0 ]];
do
  ((CT=CT+1))
  if [[ "$CT" -ge "$MAX_TRIES" ]];
  then
    break
  fi
  printf "${GREEN}Testing private transactions: attempt $CT out of $MAX_TRIES${NC}\n"
  echo helpers/run_contracts.sh $NODE_TO_TEST priv $NAMESPACE_NAME
  helpers/run_contracts.sh $NODE_TO_TEST priv $NAMESPACE_NAME
  EXIT_CODE=$?
  echo "exit code for private tx is [$EXIT_CODE]"
  sleep 2
  get_block_number $NAMESPACE
  BLOCK_NUM=$?
  echo "BLOCK_NUM: [$BLOCK_NUM]"
done

if [[ $EXIT_CODE -ne 0 ]];
then
  printf "${RED}ERROR: private transaction failed!${NC}\n"
  exit 1
fi

# check that the blocks increased. (TODO: only needs to be verified for Raft)
get_block_number $NAMESPACE
BLOCK_NUM=$?
echo "BLOCK_NUM: [$BLOCK_NUM]"

# If raft the block numbers should be at least 2, if istanbul we are not sure.
if [[ "$BLOCK_NUM" -ge 2 ]];
then
   printf "${GREEN}SUCCESS: successfully executed a public and private transaction!! ${NC}\n"
   sleep 2
   exit 0
else
   printf "${RED}ERROR: executing a private transactions!! ${NC}\n"
   exit 1
fi