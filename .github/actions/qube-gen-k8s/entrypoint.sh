#!/bin/sh -l

CONFIG_PREFIX=$1
echo "Running test for config prefix $CONFIG_PREFIX"
# generate the K8s resrouces using the file in this commit/github workspace
/github/workspace/testing/gen-configs.sh ${CONFIG_PREFIX}
