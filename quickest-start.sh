#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

usage() {
  echo "  ./quickest-start.sh "
  echo "  ./quickest-start.sh NUM"
  echo "   If no number is passed in, deploys the 7node tessera IBFT network"
  echo "   If a number is passed in then a network with that number of nodes will be created."
  echo "   Requires Docker to be installed and running"
}

NUM_NODES=0
if [[ $# -gt 0 ]];
then
  if [[ $1 -eq "help" ]]; then
    usage
  else
    NUM_NODES=$1
    echo "Creating a $NUM_NODES network."
    echo
  fi
else
  echo
  echo "Deploying 7nodes 7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc/ "
  echo
fi


## make sure docker is installed
docker ps > /dev/null
EXIT_CODE=$?

if [[ EXIT_CODE -ne 0 ]];
then
  printf "${RED}Error: docker is not running, please start docker before running this script.${NC}\n"
  usage
fi

## make sure kind is installed
kind version
EXIT_CODE=$?

if [[ EXIT_CODE -ne 0 ]];
then
  # try to install kind
  echo "Kind is not installed, going to install it for MacOS"
  brew install kind
fi

function wait_for_running_pods() {
 # make sure all the pods come up and are in a RUNNING state.
  CT=0
  MAX_ATTEMPTS=11
  ALL_RUNNING="false"
  while [[ ${ALL_RUNNING} != "true" && $CT -lt $MAX_ATTEMPTS ]]
  do
    ((CT=CT+1))
    echo "Attempt $CT"
    printf "${GREEN}Waiting for all PODs to be in the RUNNING state.${NC} \n"
    RUNNING="true"
    # PODS is POD_NAME and POD_STATUS (Running | Pending, etc) NAME%%STATUS
    # get all pods NAME%%STATUS in order to test if all pods are running yet.
    # Set up the returned pods so we can loop through them [NAME%%STATUS].
    PODS_NAME_STATUS=$(kubectl get pods $NAMESPACE | grep quorum | awk '{print $1"%%"$3}')
    # echo "PODS_NAME_STATUS: [$PODS_NAME_STATUS]"
    RES=$?
    if [[ $RES -gt 0 ]];
    then
      printf "${RED}Issue applying pods, exiting.${NC}\n"
      exit 1
    fi
    # if there are no pods returned, this may be because the kuberentes backend is taking a bit
    # longer to initialize the PODs, wait for a few loops:
    # TODO: terminate after several attempts with a failure code.
    if [[ -z "$PODS_NAME_STATUS" ]]; then
      printf "${RED}Pods are not up yet, wait 10 seconds before trying again.${NC}\n"
      sleep 10
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
    echo "ALL_RUNNING == ${ALL_RUNNING}"
    echo
    sleep 10
  done

  if [[ $CT -ge $MAX_ATTEMPTS ]];
  then
    printf "${RED}ISSUE: The pods are taking a long time to get to the RUNNING state${NC}"
    echo " Potential issue: docker does not have enough memory to run the desireed network. "
    echo " Increase Docker Engine's memory and try again. "
    exit 1
  fi
}

echo "Removing any existing kind quickest-qube cluster"
kind delete cluster --name quickest-qube
echo "Starting kind cluster"
echo
kind create cluster --name quickest-qube
echo
echo "kind cluster created"


if [[ $NUM_NODES -gt 0 ]];
then
  cat qubernetes.yaml | sed "s/number:.*/number: $NUM_NODES/g" > quickest-start.yaml
  echo docker run --rm -it -v $(pwd):/qubernetes quorumengineering/qubernetes ./qube-init quickest-start.yaml
  docker run --rm -it -v $(pwd):/qubernetes quorumengineering/qubernetes ./qube-init quickest-start.yaml
  kubectl apply -f out -f out/deployments > /dev/null
else
  echo "Deploying 7nodes with Privacy manager tessera and consensus IBFT"
  kubectl apply -f 7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc/ > /dev/null
fi

echo "Waiting for pods to come up"
wait_for_running_pods
echo
printf "${GREEN} SUCCESS! Pods are up and running!${NC}\n"
kubectl get pods
echo
echo "==============================================================="
echo
echo "To connect to a node once the pods have been started, run:"
printf "${GREEN}$> ./connect.sh node1 quorum ${NC} \n"
echo
echo "Quorum resources are under \$QHOME on the pod."
printf "${GREEN}$> cd \$QHOME${NC} \n"
echo
echo "To run some transcations from inside the quorum container:"
printf "${GREEN}$> cd \$QHOME/contracts${NC} \n"
printf "${GREEN}$> ./runscript.sh public_contract.js${NC}\n"
printf "${GREEN}$> ./runscript.sh private_contract.js${NC}\n"

echo
