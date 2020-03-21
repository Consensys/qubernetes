#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

function usage() {
  echo " ./test-k8s-resources.sh {CONFIG_PREFIX}"
  echo
  echo " example: ./test-k8s-resources.sh 7nodes"
  echo "          expects a directory testing/{CONFIG_PREFIX}-config to exist."
  echo "          the default is '7nodes' if no test name is passed in."
}

# If no CONFIG_PREFIX was passed in, default to 'CONFIG_PREXIX=7nodes'.
# `testing/7nodes-config` directory should exist with sym links to the qubernetes/7nodes/ qubernetes config files.
if [[ $# -eq 0 ]];
then
  CONFIG_PREFIX=7nodes
elif [[ $# -eq 1 ]];
then
  CONFIG_PREFIX=$1
else
  usage
  exit 1
fi

# TODO: allow various backend kubernetes clusters, e.g. minikube, kind, gke.
function restart_k8s_cluster() {
# minikube delete
# sleep 3
# minikube delete
# minikube start --memory 6144
# When running kind, make sure docker has enough memory or it may fail.
# https://kind.sigs.k8s.io/docs/user/known-issues/
  kind delete cluster
  kind create cluster
}

function run_test_qnet() {
 EXIT_CODE=1
 # the kuberentes resources have been applied at this point, test-qnet.sh will query the namespace and test the quorum deployments.
 testing/test-qnet.sh $NAMESPACE &&
 EXIT_CODE=$?
 # if there is an error exit code wait for some manual checking
 if [[ $EXIT_CODE -ne 0 ]];
 then
   # Wait here to allow for manual checking / testing.
   ((FAILURES=FAILURES+1))
   echo "Hit any key to continue"
   read NEXT
 else
   ((SUCCESS=SUCCESS+1))
 fi
}

# Test will be run against the out put directories under the base output directory. These are directories that hold
# the kubernetes API resource files.
OUT_DIR_BASE=testing/${CONFIG_PREFIX}-out

restart_k8s_cluster

TOTAL_TEST_NUM=$(ls -1 $OUT_DIR_BASE | wc -l | sed 's| ||g')
printf "${GREEN} Testing ${TOTAL_TEST_NUM} test networks. ${NC}\n"

# now go through the examples and test k8s-yaml set of resources
# that were in the directory.
CT=0
SUCCESS=0
FAILURES=0
for OUT_DIR in $OUT_DIR_BASE/*;
do
 ((CT=CT+1))
 printf "${GREEN} Running Testing $CT out of ${TOTAL_TEST_NUM} ${OUT_DIR}${NC}\n"
 printf "${GREEN}Total Successful networks: ${SUCCESS}${NC}\n"
 printf "${RED}Total Failed networks: ${FAILURES}${NC}\n"
 # remove $OUT_DIR_BASE from the string to create the namespace
 echo $OUT_DIR | sed "s|$OUT_DIR_BASE||g" | sed 's|/||g'
 NAMESPACE=$(echo $OUT_DIR | sed "s|$OUT_DIR_BASE||g" | sed 's|/||g')
 echo $NAMESPACE
 CUR_OUT_DIR=${OUT_DIR_BASE}/$NAMESPACE

 # Create namespace for this run
 kubectl delete namespace $NAMESPACE
 kubectl create namespace $NAMESPACE &&

 # test if /deployments directory exists in the case where the kubernetes resources were generated with separate
 # deployment files.
 if [[ -d $CUR_OUT_DIR/deployments ]];
 then
   printf "${GREEN}kubectl apply -f $CUR_OUT_DIR -f $CUR_OUT_DIR/deployments --namespace=$NAMESPACE ${NC}\n"
   kubectl apply -f $CUR_OUT_DIR -f $CUR_OUT_DIR/deployments --namespace=$NAMESPACE > /dev/null
 else
   printf "${GREEN}kubectl apply -f $CUR_OUT_DIR --namespace=$NAMESPACE ${NC}\n"
   kubectl apply -f $CUR_OUT_DIR --namespace=$NAMESPACE > /dev/null
 fi

## Depending on the service used it may take a while for the PODS to appear
EXIT_CODE=1
while [[ $EXIT_CODE -ne 0 ]];
do
  # Test the now deployed quorum network
  EXIT_CODE=1
  echo testing/test-qnet.sh $NAMESPACE
  testing/test-qnet.sh $NAMESPACE &&
  EXIT_CODE=$?
  # if there is an error exit code wait for some manual checking
  if [[ $EXIT_CODE -ne 0 ]];
  then
    # Wait here to allow for manual checking / testing.
    echo "Issue running test: Hit 1 to run txs again any key to record failure"
    read RES
    # run again?
    if [[ $RES -eq 1 ]];
    then
      echo "Trying again..."
    else
      break
    fi
  fi
done

# record result of running the test
if [[ $EXIT_CODE -ne 0 ]];
then
  ((FAILURES=FAILURES+1))
else
  ((SUCCESS=SUCCESS+1))
fi
 kubectl delete namespace $NAMESPACE
 HOST_STORAGE=$(echo $NAMESPACE | grep host)
 if [[ ! -z "$HOST_STORAGE" ]];
 then
   printf "${RED}Deleting kubernetes deployment because this test run used host storage${NC}\n"
   restart_k8s_cluster
   echo "restarted cluster"
 fi
done

printf "${GREEN}Total Successful networks: ${SUCCESS}${NC}\n"
printf "${RED}Total Failed networks: ${FAILURES}${NC}\n"