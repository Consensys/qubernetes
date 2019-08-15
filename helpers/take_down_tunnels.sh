#!/bin/bash


# helper script to take down tunnels that have been
# setup by `set_up_tunnel.sh
IPS=$(kubectl get svc | grep "quorum" | awk '{ print $3}')
for IP in $IPS
do
 sudo ifconfig lo0 -alias $IP
done

PIDS=$(ps -ef | grep "ssh -N -o StrictHostKeyChecking no -L" | awk '{print $2}')
for PID in $PIDS
do
 echo $PID
 kill $PID
done
