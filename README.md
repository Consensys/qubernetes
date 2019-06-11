## Qubernetes

A project for deploying [Quorum](https://github.com/jpmorganchase/quorum) on [Kubernetes](https://github.com/kubernetes/kubernetes).

## Install 
```shell
$> brew install ruby
$> gem install colorize
```
* To test locally install [minikube](https://kubernetes.io/docs/setup/minikube/) for your distro.
```shell
$> brew cast install minikube
$> minikube start --memory 6144

# when you wish to shutdown
$> minikube stop
$> minikube delete
```
## Configuration 
The main configuration files are [`qubernetes.yaml`](qubernetes.yaml) and [`nodes.yaml`](nodes.yaml). 
`qubernetes.yaml` can generate `nodes.yaml`. These two configuration yaml files must exist in the base directory.

By default `qubernetes.yaml` is symlinked to [7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-7nodes.yaml](7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-7nodes.yaml), but it can be changed
to point to the desired configurations, e.g. to regenerate the [istanbul-7nodes-tessera/k8s-yaml](7nodes/istanbul-7nodes-tessera/k8s-yaml) 
use [7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-7nodes.yaml](7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-7nodes.yaml): 
```shell
$> ln -s 7nodes/istanbul-7nodes/qubernetes-istanbul-7nodes.yaml qubernetes.yaml
$> ln -s 7nodes/nodes.yaml nodes.yaml
# generate the resource yaml in the ./out dir
$> ./qubernetes
$> kubectl apply -f out
```

## [7 Nodes Example](https://github.com/jpmorganchase/quorum-examples/tree/master/examples/7nodes)
quorum-exmaples 7nodes has been ported to k8s resources:

* [istanbul tessera k8s resource yaml](7nodes/istanbul-7nodes-tessera/k8s-yaml)
```shell
$> kubectl apply -f 7nodes/istanbul-7nodes-tessera/k8s-yaml
```
* [istanbul constellation k8s resource yaml](7nodes/istanbul-7nodes-constellation/k8s-yaml)
```shell
$> kubectl apply -f 7nodes/istanbul-7nodes-constellation/k8s-yaml
```
* [raft tessera k8s resource yaml](7nodes/raft-7nodes-tessera/k8s-yaml)
```shell
$> kubectl apply -f 7nodes/raft-7nodes-tessera/k8s-yaml
```
* [raft constellation k8s resource yaml](7nodes/raft-7nodes-constellation/k8s-yaml)
```shell
$> kubectl apply -f 7nodes/raft-7nodes-constellation/k8s-yaml
```
### Deleting
1. Delete the kubernetes resources:
```shell
$> kubeclt delete -f 7nodes/istanbul-7nodes-tessera/k8s-yaml
```
2. Delete the files on the host machine, by default under `/var/lib/docker/geth-storage`.


## Generating Quroum K8s Resources From Configs 

### Install Prerequisites
* [`bootnode`](https://github.com/ethereum/go-ethereum/tree/master/cmd/bootnode) (geth) for generating keys. 
   If you have geth source on your machine
   ```
    $> cd go-ethereum 
    go-ethereum $> make all
    # or place this in your .bash_profile or equivalent file
    $> export PATH="~/go/src/github.com/ethereum/go-ethereum/build/bin:$PATH"
   ```
* [nodejs](https://nodejs.org/en/download/) Istanbul only.
* [istanbul-tools](https://github.com/jpmorganchase/istanbul-tools) Istanbul only.
   
1. Set `qubernetes.yaml` in this directory to the desired configuration, there are some example configs in [`examples/config`](examples/config).
create a symlink `ln -s examples/config/qubernetes-istanbul-generate-8nodes.yaml qubernetes.yaml` if you wish to use it, or cp it to this direction
`cp examples/config/qubernetes-istanbul-generate-8nodes.yaml qubernetes.yaml`.
The most basic thing to modify in `qubernetes.yaml` is the number of nodes you wish to deploy: 
```yaml
# number of nodes to deploy
nodes:
  number: 8
```

2. Run `./quorum-init` to generate the necessary keys (**note**: this requires the geth `bootnode` command to be on your path,
),
 genesis.json, and permissioned-nodes.json needed for the quorum deployment. 
These resources will be written to the directory specified in the [`qubernetes.yaml`](qubernetes.yaml)
and generate the necessary k8s resources in the `out` directory:
```shell
$> rm -r out
# Generate the keys, permissioned-nodes.json file
# genesis.json for the configured nodes
$> ./quorum-init
```
* These resources will be written to (and read from) the directories specified in the `qubernetes.yaml` the default [`qubernetes.yaml`](config/qubernetes.yaml)
is configured to write theses to the `./out/config` directory.
```yaml
Key_Dir_Base: out/config 
Permissioned_Nodes_File: out/config/permissioned-nodes.json
Genesis_File: out/config/genesis.json
```

3. (Re)generate the Kubernetes resource yaml for the deployment. By default these will be generated to the `./out` directory.
This command can be run multiple times and be idempotent as long as the underlying Quorum resources do not change.

```shell
# Generate the kubernetes resources necessary to support a Quorum deploy
# this will be written to the `out` dir.
$> ./qubernetes

```
4. Deploy to your kubernetes cluster

```shell
# apply all the generated .yaml files that are in the ./out directory.
$> kubectl apply -f out
```


4. Accessing the quorum container: 

```shell
local $> kubectl get pods --namespace=$YOUR_NAMESPACE
local $> kubectl config set-context $(kubectl config current-context) --namespace=$YOUR_NAMESPACE 
local $> kubectl  exec -it $POD_ID -c quorum /bin/ash
quorum-qubernetes $> geth attach /etc/quorum/qdata/dd/geth.ipc
> eth.blockNumber
> 0

quorum-qubernetes $> cd /etc/quorum/qdata/contracts
quorum-qubernetes $> ./runscript.sh public_contract.js
quorum-qubernetes $> ./runscript.sh private_contract.js

# you should know see the txs go through
# note: if you are running istanbul the blocknumber will automatic increment at a steady interval.
quorum-qubernetes $> geth attach /etc/quorum/qdata/dd/geth.ipc
> eth.blockNumber
> 2

# show connected peers
> admin.peers 

```


5. Deleting the deployment 

```shell

$> kubectl delete -f out
# make sure to delete the data directory on the base box
# e.g. minikube
$> minikube ssh
$> sudo su
$> rm -r /var/lib/docker/geth-storage
```
* Delete the files on the host machine, by default under `/var/lib/docker/geth-storage`.

## Thanks! And Additional Resources 
Thanks to [Maximilian Meister blog and code](https://medium.com/@cryptoctl) which provided and awesome starting point!
and is a good read to undestand the different components.
