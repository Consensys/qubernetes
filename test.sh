#!/bin/bash

# Runs the tests:
# 1. generate fresh kubernetes resource yaml files from quebernetes config files.
# 2. deploy the freshly generated configs to a K8s environment.
function usage() {
  echo " ./test.sh {CONFIG_PREFIX}"
  echo
  echo " example: ./test.sh 7nodes"
  echo "          expects a directory testing/{CONFIG_PREFIX}-config to exist."
  echo "          the default is '7nodes' if no test name is passed in."
}

# If no ${CONFIG_PREFIX} was given default to `CONFIG_PREFIX=7nodes`.
# 7nodes-config directory should exist and point to the `qubernetes/7nodes/` qubernetes config files.
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

docker run --rm -it -v $(pwd):/qubernetes quorumengineering/qubernetes testing/gen-configs.sh ${CONFIG_PREFIX}
#testing/gen-configs.sh ${CONFIG_PREFIX}
EXIT_CODE=$?
if [[ $EXIT_CODE -ne 0 ]]; then
  echo "  Error: could not generate configs."
  exit 1
fi
testing/test-k8s-resources.sh ${CONFIG_PREFIX}