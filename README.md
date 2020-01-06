## Qubernetes

A project for deploying [Quorum](https://github.com/jpmorganchase/quorum) on [Kubernetes](https://github.com/kubernetes/kubernetes).

## Install 
```shell
$> brew install ruby

# check ruby version > 2.6
$> ruby --version
   ruby 2.6.3
$> gem install colorize
```

## Quickstart with minikube

This section shows how to deploy a 7 node quorum network to minikube. Once the pods are deployed if
you wish to interact with them see [Accessing Quorum and Transaction Manager Containers on K8s](#accessing-quorum-and-transaction-manager-containers-on-k8s).

*  Install [minikube](https://kubernetes.io/docs/setup/minikube/) for your distro.
```shell
$> brew cask install minikube
$> minikube start --memory 6144

# you should be able to ssh into minikube
$> minikube ssh

# update your kubectl command see: https://stackoverflow.com/questions/55417410/kubernetes-create-deployment-unexpected-schemaerror 
$> rm /usr/local/bin/kubectl
$> brew link --overwrite kubernetes-cli

# version tested with
$> kubectl version
Client Version: version.Info{Major:"1", Minor:"14", GitVersion:"v1.14.3", GitCommit:"5e53fd6bc17c0dec8434817e69b04a25d8ae0ff0", GitTreeState:"clean", BuildDate:"2019-06-07T09:55:27Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"14", GitVersion:"v1.14.2", GitCommit:"66049e3b21efe110454d67df4fa62b08ea79a19b", GitTreeState:"clean", BuildDate:"2019-05-16T16:14:56Z", GoVersion:"go1.12.5", Compiler:"gc", Platform:"linux/amd64"}

# generate k8s files
$> ./qubernetes
# by default the k8s namespace is set to quorum-test, set this in the kubectl global config
$> kubectl config set-context $(kubectl config current-context) --namespace=quorum-test

# deploy quorum to your minikube
$> kubectl apply -f out

# you should now see your pods running.
$> kubectl get pods

# delete minikube k8s deployment
$> kubectl delete -f out

# when you wish to shutdown minikube
$> minikube stop
$> minikube delete
```

### Deleting Kubernetes Deployment

1. Delete the kubernetes resources:
```shell
$> kubectl delete -f PATH/TO/K8S-YAML-DIR
```
2. Delete the files on the host machine, by default under `/var/lib/docker/geth-storage`.
```
$> minikube ssh
$> sudo su
$> rm -r /var/lib/docker/geth-storage
```

## 7 Nodes Examples
[quorum-examples 7nodes](https://github.com/jpmorganchase/quorum-examples/tree/master/examples/7nodes) has been ported to k8s resources.
There are k8s resource files in this repo's 7nodes directory for deploying quorum on kubernetes with 
tessera or constellation as the transaction manager, and raft or istanbul as the consensus engines.  

**note** when deleting remove the `/var/lib/docker/geth-storage` from the host (see above).
* [istanbul tessera k8s resource yaml](7nodes/istanbul-7nodes-tessera/k8s-yaml)
```shell
$> kubectl apply -f 7nodes/istanbul-7nodes-tessera/k8s-yaml
$> kubectl delete -f 7nodes/istanbul-7nodes-tessera/k8s-yaml
```
* [istanbul constellation k8s resource yaml](7nodes/istanbul-7nodes-constellation/k8s-yaml)
```shell
$> kubectl apply -f 7nodes/istanbul-7nodes-constellation/k8s-yaml
$> kubectl delete -f 7nodes/istanbul-7nodes-constellation/k8s-yaml
```
* [raft tessera k8s resource yaml](7nodes/raft-7nodes-tessera/k8s-yaml)
```shell
$> kubectl apply -f 7nodes/raft-7nodes-tessera/k8s-yaml
$> kubectl delete -f 7nodes/raft-7nodes-tessera/k8s-yaml
```
* [raft constellation k8s resource yaml](7nodes/raft-7nodes-constellation/k8s-yaml)
```shell
$> kubectl apply -f 7nodes/raft-7nodes-constellation/k8s-yaml
$> kubectl delete -f 7nodes/raft-7nodes-tessera/k8s-yaml
```

### Accessing Quorum and Transaction Manager Containers on K8s

```shell
export YOUR_NAMESPACE="quorum-test"

$> kubectl get pods --namespace=$YOUR_NAMESPACE
# additionally you can set the default namespace for k8s so you don't have to add the --namespace flag.
$> kubectl config set-context $(kubectl config current-context) --namespace=$YOUR_NAMESPACE
$> kubectl get pods
NAME                                       READY   STATUS    RESTARTS   AGE
quorum-node1-deployment-57b6588b6b-5tqdr   1/2     Running   1          40s
quorum-node2-deployment-5f776b479c-f7kxs   2/2     Running   2          40s

$> POD_NAME=$(kubectl get pods | grep node1 | awk '{print $1}')
$> kubectl  exec -it $POD_NAME -c quorum /bin/ash

# now you are inside the quorum container
/> geth attach /etc/quorum/qdata/dd/geth.ipc
> eth.blockNumber
> 0

/> cd /etc/quorum/qdata/contracts
/> ./runscript.sh public_contract.js
/> ./runscript.sh private_contract.js

# you should know see the txs go through
# note: if you are running istanbul the blocknumber will automatic increment at a steady interval.
/> geth attach /etc/quorum/qdata/dd/geth.ipc
> eth.blockNumber
> 2

# show connected peers
> admin.peers 

```

## Configuration Details
The main configuration files are [`qubernetes.yaml`](7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-7nodes.yaml) and [`nodes.yaml`](7nodes/nodes-7.yaml). 
These two configuration yaml files must exist in the base directory.

By default `qubernetes.yaml` is symlinked to [`qubernetes-istanbul-7nodes.yaml`](7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-7nodes.yaml)
```
$> ls -la qubernetes.yaml
qubernetes.yaml -> 7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-7nodes.yaml 
```
But can be changed to point to other yaml config files. 

Let's change it to generate the **raft** 7nodes with tessera: 
```shell
$> rm qubernetes.yaml
$> ln -s 7nodes/raft-7nodes-tessera/qubernetes-tessera.yaml qubernetes.yaml
# generate the resource yaml in the ./out dir
$> ./qubernetes
$> kubectl apply -f out
$> kubectl config set-context $(kubectl config current-context) --namespace=quorum-test
$> kubectl get pods

# now you can delete 
$> kubectl delete -f out
# and remove dir /var/lib/docker/geth-storage from the host.
```

## Generating Quroum K8s Resources From Custom Configs

This section describes how to deploy a more customized k8s deployment with a varying number of quorum and transaction 
manager nodes, including generating the appropriate genesis config, required keys, services, etc. 

### Install Prerequisites
* [`bootnode`](https://github.com/ethereum/go-ethereum/tree/master/cmd/bootnode) (geth) for generating keys. 
   ```
     # what you should see if installed.
     $> bootnode
     Fatal: Use -nodekey or -nodekeyhex to specify a private key
   ```
   
   If you have geth source on your machine: 
   ```
    $> cd go-ethereum 
    go-ethereum $> make all
    # or place this in your .bash_profile or equivalent file
    $> export PATH="~/go/src/github.com/ethereum/go-ethereum/build/bin:$PATH"
   ```
* [nodejs](https://nodejs.org/en/download/) Istanbul only.
  ```
   # tested with version 10.15
   $> node --version
   v10.15.
   ```
* web3 `$> npm web3` Istanbul only.

* [constellation-node](https://github.com/jpmorganchase/constellation)
  ```
  $> brew install berkeley-db leveldb libsodium
  $> brew install haskell-stack
  $> git clone https://github.com/jpmorganchase/constellation.git WHATEVER/DIRECTORY
  $> cd constellation
  constellation $> stack setup
  constellation $> stack install
  ```

* [istanbul-tools](https://github.com/jpmorganchase/istanbul-tools) Istanbul only.
  ```
   # install
   $> go get github.com/jpmorganchase/istanbul-tools/cmd/istanbul  
   ```
### Generating Your Own K8s Resources

Once the install prerequisites are on your machine, the k8s resources can now be generated to run an arbitrary number of nodes.
   
1. There are example `qubernetes.yaml` configs in [`examples/config`](examples/config). 
   Let's run the 8nodes example: 
```
 $> rm qubernetes.yaml
 $> cp examples/config/qubernetes-istanbul-generate-8nodes.yaml qubernetes.yaml
 
 # remove your old out dir, this will be recreated.
 $> rm -r out
```  
The most basic thing to modify in `qubernetes.yaml` is the number of nodes you wish to deploy: 
```yaml
# number of nodes to deploy
nodes:
  number: 8
```

2. Run `./quorum-init` to generate the necessary quorum keys (**note**: this requires the geth `bootnode` command to be on your path),
 genesis.json, permissioned-nodes.json, etc. needed for the quorum deployment.
  
 These resources will be written to (and read from) the directories specified in the `qubernetes.yaml` the default [`qubernetes.yaml`](config/qubernetes.yaml)
 is configured to write theses to the `./out/config` directory.
 ```yaml
 Key_Dir_Base: out/config 
 Permissioned_Nodes_File: out/config/permissioned-nodes.json
 Genesis_File: out/config/genesis.json
 ```
 
 After the quorum  resources have been generate, the necessary k8s resources will be created in the `out` directory:
```shell
# if you want to check out the generated quorum resources
$> ls out/config

$> rm -r out
# Generate the keys, permissioned-nodes.json file
# genesis.json for the configured nodes
$> ./quorum-init
$> ls out
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

## Docker bootstrap container

To avoid installing all the prerequisites, you can use a docker container with all prerequisites already installed.

Usage:
```
docker run -ti jpmorganchase/qubernetes
./quorum-init
# point kubectl to a correct cluster
# for example for k8s on gcloud do `gcloud init` and paste config command from "connect" button in UI
kubectl apply -f out
```

If you have qubernetes checked out and with custom configs, you can mount it to a container adding `-v $(pwd):/qubernetes` to your `docker run` command

## Thanks! And Additional Resources 
Thanks to [Maximilian Meister blog and code](https://medium.com/@cryptoctl) which provided and awesome starting point!
and is a good read to undestand the different components.
