## Qubernetes

A project for deployng [Quorum](https://github.com/jpmorganchase/quorum) on [Kubernetes](https://github.com/kubernetes/kubernetes).

Thanks to https://medium.com/@cryptoctl which provided 
and awesome starting point.

## Quick Start
* Set up the intial config in `qubernetes.yaml`

1. Genearte the kubernetes resource yaml files required
   for a Quorum deployment.
```
# Generate the keys, permissioned-nodes.json file
# genesis.json for the configured nodes
$> ./quorum-init

# Generate the kubernetes resources 
# necessary to support a Quorum deploy
# this will be written to the `out` dir.
$> ./qubernetes

```
2. Deploy to kubernetes

* see helper scripts `deploy.sh`

```
kubectl apply -f out/quorum-shared-config.yaml
kubectl apply -f out/quorum-services.yaml
kubectl apply -f out/quorum-keyconfigs.yaml
kubectl apply -f out/quorum-deployments.yaml
```


3. Accessing your nodes

```
local $> kubectl get pods --namespace=$YOUR_NAMESPACE 
local $> kubect exec -it $POD_ID -c quorum /bin/ash
quorum-qubernetes $> cd /etc/quorum/qdata
quorum-qubernetes $> ls 
quorum-qubernetes $> geth attach dd/geth.ipc 
> eth.blockNumber
> 0
> exit

quorum-qubernetes $> cd /etc/quorum/qdata/contracts
quorum-qubernetes $>./runscript.js public_contract.js 

# you should know see the tx go through
quorum-qubernetes $> geth attach /etc/quorum/qdata/dd/geth.ipc 
> eth.blockNumber
> 1 

# show connected peers
> admin.peers 

```


3. Deleting the deployment 

* see helper scripts `deploy.sh`

```
kubectl delete -f out/quorum-shared-config.yaml
kubectl delete -f out/quorum-services.yaml
kubectl delete -f out/quorum-keyconfigs.yaml
kubectl delete -f out/quorum-deployments.yaml
```
