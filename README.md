## Qubernetes

A project for deploying [Quorum](https://github.com/jpmorganchase/quorum) on [Kubernetes](https://github.com/kubernetes/kubernetes).


## Configuration 
The main configuration files are [`qubernetes.yaml`](config/qubernetes.yaml) and [`nodes.yaml`](nodes.yaml). 
`qubernetes.yaml` can generate `nodes.yaml`. These two configuration yaml files must exist in the base directory.

By default `qubernetes.yaml` is symlinked to [config/qubernetes.yaml](config/qubernetes.yaml), but it can be changed
to point to the desired configurations, e.g. to regenerate the [istanbul-7nodes-tessera/k8s-yaml][7nodes/istanbul-7nodes-tessera/k8s-yaml] 
use [7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-7nodes.yaml](7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-7nodes.yaml): 
```shell
$> ln -s 7nodes/istanbul-7nodes/qubernetes-istanbul-7nodes.yaml qubernetes.yaml
$> ln -s 7nodes/nodes.yaml nodes.yaml
# generate the resource yaml in the ./out dir
$> ./qubernetes
```

## [7 Nodes Example](https://github.com/jpmorganchase/quorum-examples/tree/master/examples/7nodes)
quorum-exmaples 7nodes has been ported to k8 resources:

* [istanbul tessera k8s resource yaml](7nodes/istanbul-7nodes-tessera/k8s-yaml)
```shell
$> kubeclt apply -f 7nodes/istanbul-7nodes-tessera/k8s-yaml
```
* [istanbul constellation k8s resource yaml](7nodes/istanbul-7nodes-constellation/k8s-yaml)
```shell
$> kubeclt apply -f 7nodes/istanbul-7nodes-constellation/k8s-yaml
```
* [raft tessera k8s resource yaml](7nodes/raft-7nodes-tessera/k8s-yaml)
```shell
$> kubeclt apply -f 7nodes/raft-7nodes-tessera/k8s-yaml
```
* [raft constellation k8s resource yaml](7nodes/raft-7nodes-constellation/k8s-yaml)
```shell
$> kubeclt apply -f 7nodes/raft-7nodes-constellation/k8s-yaml
```
### Deleting
1. Delete the kubernetes resources:
```shell
$> kubeclt delete -f 7nodes/istanbul-7nodes-tessera/k8s-yaml
```
2. Delete the files on the host machine, by default under `/var/lib/docker/geth-storage`.


## Generating Quroum K8s Resources From Configs 

1. Set `qubernetes.yaml` in this directory to the desired confufation, there is a base config in [`config/qubernetes.yaml`](config/qubernetes.yaml).
create a symlink `ln -s config/qubernetes.yaml .` if you wish to use it, or cp it to this direction.
The most basic thing to modify in `qubernetes.yaml` is the number of nodes you wish to deploy: 
```yaml
# number of nodes to deploy
nodes:
  number: 10
```

2. Run `./quorum-init` to generate the necessary keys, genesis.json, and permissioned-nodes.json needed for the quorum deployment. 
These resouces will be written to the directory specified in the [`qubernetes.yaml`](qubernetes.yaml):
```shell
# Generate the keys, permissioned-nodes.json file
# genesis.json for the configured nodes
$> ./quorum-init
```
* These resouces will be written to (and read from) the directories specified in the `qubernetes.yaml` the default [`qubernets.yaml`](config/qubernetes.yaml)
is configured to write theses to the `./out/config` directory.
```yaml
Key_Dir_Base: out/config 
Permissioned_Nodes_File: out/config/permissioned-nodes.json
Genesis_File: out/config/genesis.json
```

3. Generate the Kubernetes resource yaml for the deployment. By default these will be generated to the `./out` directory.

```shell
# Generate the kubernetes resources necessary to support a Quorum deploy
# this will be written to the `out` dir.
$> ./qubernetes

```
4. Deploy to your kubernetes cluster

```shell
# apply all the generated .yaml files that are in the ./out directory.
$> kubectl apply -f out/
```


4. Accessing the quorum container: 

```shell
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


5. Deleting the deployment 

```shell

$> kubeclt delete -f out/ 
```
* Delete the files on the host machine, by default under `/var/lib/docker/geth-storage`.

## Thanks! And Additional Resources 
Thanks to [Maximilian Meister blog and code](https://medium.com/@cryptoctl) which provided and awesome starting point!
and is a good read to undestand the different components.
