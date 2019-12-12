#!/bin/ash
set -xe

for Addr in $( awk '/validators/,0' $QHOME/istanbul-validator-config.toml/istanbul-validator-config.toml | grep "0x" | sed 's/,//g; s/"//g' ); do
  echo $Addr
  $QHOME/node-management/ibft_propose.sh $Addr true
done