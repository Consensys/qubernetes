#!/bin/bash

# set up tunnels for a minikube quorum deployment, so
# that applications, such as cakeshop, can connected
# to the transaction manager without mapping proxied
# IP addresses, e.g. minikube sets up proxy IPs, but
# these will be different from the IPs used by the internal
# cluster, mainly in curl http://$IP:9001/partyinfo.
IPS=$(kubectl get svc | grep "quorum" | awk '{ print $3}')
MINI_IP=$(minikube ip)
for IP in $IPS
do
 echo $IP
 sudo ifconfig lo0 alias $IP
 echo "setting up port forwarding mini $MINI_IP"
 # don't prompt for 'The authenticity of host' / known_host on first connection.
 ssh -N  -o "StrictHostKeyChecking no" -L $IP:9001:$IP:9001 -i ~/.minikube/machines/minikube/id_rsa docker@$MINI_IP  &
 ssh -N  -o "StrictHostKeyChecking no" -L $IP:8545:$IP:8545 -i ~/.minikube/machines/minikube/id_rsa docker@$MINI_IP  &
done

for IP in $IPS
do
 echo "curl http://$IP:9001/partyinfo"
done
