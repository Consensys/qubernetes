#!/bin/bash

# Expects a name for directory prefix containing the configs to test, e.g. 7nodes, test
# the qubernetes.yaml files are expected to be in a directory with that name prefix, e.g. ${CONFIG_PREFIX}-config
# 7nodes-config, test-configs defaults to 7nodes if no {CONFIG_PREFIX} is provided.
function usage() {
  echo " ./gen-configs {CONFIG_PREFIX}"
  echo "  example: ./gen-configs 7nodes"
  echo "  expects a directory testing/{CONFIG_PREFIX}-config to exist."
  echo "  the default is '7nodes' if no CONFIG_PREFIX is passed in."
}
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

# based on quberenetes config files inside `testing/{CONFIG_PREFIX}-config/` directory:
# 1. takes the base name of each qubernetes yaml config and create a directory `testing/${CONFIG_PREFIX}-out/out.${BASE_NAME}/`
# 2. generates the kubernetes resources yaml from the config.
# 3. place the kubernetes yaml into the newly created directory, e.g. `testing/${CONFIG_PREFIX}-out/out.${BASE_NAME}/`
CONFIG_DIR=testing/${CONFIG_PREFIX}-config/
OUT_DIR=testing/${CONFIG_PREFIX}-out

echo "CONFIG $CONFIG_DIR"
# remove out dir if it exist
rm -rf out
rm -rf $OUT_DIR
mkdir -p $OUT_DIR
for CONFIG_FILE in $CONFIG_DIR*;
do
 echo "Config file: ${CONFIG_FILE}"
 NAMESPACE=$(echo $CONFIG_FILE | sed 's/.yaml//g' | sed "s|$CONFIG_DIR||g")
 ./quorum-init ${CONFIG_FILE} &&
 rm -rf $OUT_DIR/$NAMESPACE
 mv out $OUT_DIR/$NAMESPACE
done
