#!/bin/bash

QIPS=$(kubectl get services | grep quorum | awk '{print $3}')
for IP in $QIPS
do
 echo "geth attach http://$IP:8545"
done
