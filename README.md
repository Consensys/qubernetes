## Qubernetes

A project for deploying [Quorum](https://github.com/jpmorganchase/quorum) on [Kubernetes](https://github.com/kubernetes/kubernetes),
including: 

* Kubernetes resource yaml for [quorum-examples 7nodes](https://github.com/jpmorganchase/quorum-examples/tree/master/examples/7nodes)
  This assumes you have a running Kubernetes cluster.  
  **see** [7 node example](#7-nodes-examples).
* Generating the necessary Quorum resource (keys, configs, etc.) and Kubernetes API resource yamls for a new N node quorum 
  deployment using a minimal config [`qubernetes.yaml`](qubernetes.yaml).    
  **see** [Generating Quorum k8s Resources From Custom Configs](#generating-quroum-k8s-resources-from-custom-configs). 
* Generating Kuberenetes API resource yaml from already existing Quorum setup: keys, config, etc.   
  **see** [Generating Kubernetes Object yaml From Existing Quorum Resources](#generating-kubernetes-object-yaml-from-existing-quorum-resources) 
* Quickstart for running on minikube. This is the recommended way to run intial tests and to get familiar with the 
  project and Kubernetes, as some Cloud provider Kubernetes services or clusters may vary from vanilla Kubernetes.
  This is also recommended get started way if you do not have a running Kubernetes cluster yet.   
  **see** [Quickstart with minikube](#quickstart-with-minikube).  

## Install 
```shell
$> brew install ruby

# check ruby version > 2.6
$> ruby --version
   ruby 2.6.3
$> gem install colorize
```

## 7 Nodes Examples
[quorum-examples 7nodes](https://github.com/jpmorganchase/quorum-examples/tree/master/examples/7nodes) has been ported to k8s resources.
There are k8s resource files in the qubernetes repo's [7nodes](7nodes) directory for deploying quorum on kubernetes with 
tessera or constellation as the transaction manager, and raft or istanbul as the consensus engines.  

This assume you have a running k8s cluster which you can connect to via `kubectl`.

There are two sets of configs generated, one that uses `HostPath` storage, and one that uses `PVC`(Persistent Volume Claims) storage.
The recommend storage option is `PVC` as this will be automatically deleted when the deployment is deleted.

**note** when using `HostPath` storage, when the deployment is deleted `kubectl delete -f YOUR_DEPLOYMENT_YAML/` YOU MUST ALSO remove 
the `/var/lib/docker/geth-storage` directory from the host.
PVC storage will be deleted by Kubernetes when you delete the deployment (`kubectl delete -f YOUR_DEPLOYMENT_YAML/`).

Examples below are for deploying 7nodes using PVC (Persistent Volume Claims):

* [istanbul tessera k8s resource yaml (PVC)](7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc)
```shell
$> kubectl apply -f 7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc
$> kubectl delete -f 7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc
```
* [istanbul constellation k8s resource yaml (PVC)](7nodes/istanbul-7nodes-constellation/k8s-yaml-pvc)
```shell
$> kubectl apply -f 7nodes/istanbul-7nodes-constellation/k8s-yaml-pvc
$> kubectl delete -f 7nodes/istanbul-7nodes-constellation/k8s-yaml-pvc
```
* [raft tessera k8s resource yaml (PVC)](7nodes/raft-7nodes-tessera/k8s-yaml-pvc)
```shell
$> kubectl apply -f 7nodes/raft-7nodes-tessera/k8s-yaml-pvc
$> kubectl delete -f 7nodes/raft-7nodes-tessera/k8s-yaml-pvc
```
* [raft constellation k8s resource yaml (PVC)](7nodes/raft-7nodes-constellation/k8s-yaml-pvc)
```shell
$> kubectl apply -f 7nodes/raft-7nodes-constellation/k8s-yaml-pvc
$> kubectl delete -f 7nodes/raft-7nodes-constellation/k8s-yaml-pvc
```

## Quickstart with minikube

This section demonstrates how to deploy a 7 node quorum network to [minikube](https://github.com/kubernetes/minikube) a local
kubernetes cluster. 

Once the pods are deployed, if you wish to interact with them see [Accessing Quorum and Transaction Manager Containers on K8s](#accessing-quorum-and-transaction-manager-containers-on-k8s).

### Install [minikube](https://kubernetes.io/docs/setup/minikube/) for your distro.
```shell
$> brew install minikube
```

### Once Minikube is install: Start local minikube cluster and deploy 7nodes example. 

For this example we will deploy [istanbul tessera k8s resource yaml (PVC)](7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc).

```shell

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

# deploy istanbul/tessera 7node network to local minikube cluster in the default namespace. 
$> kubectl apply -f 7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc
persistentvolumeclaim/quorum-node1-quorum created
persistentvolumeclaim/quorum-node1-tm-pvc created
persistentvolumeclaim/quorum-node1-log-pvc created
persistentvolumeclaim/quorum-node2-quorum created
persistentvolumeclaim/quorum-node2-tm-pvc created
persistentvolumeclaim/quorum-node2-log-pvc created
persistentvolumeclaim/quorum-node3-quorum created
persistentvolumeclaim/quorum-node3-tm-pvc created
persistentvolumeclaim/quorum-node3-log-pvc created
...

# you should now see your pods running.
$> kubectl get pods
NAME                                       READY   STATUS    RESTARTS   AGE
quorum-node1-deployment-67766d6d54-vm4ms   2/2     Running   0          63s
quorum-node2-deployment-dcf7d9557-ttqk6    2/2     Running   0          63s
quorum-node3-deployment-cf64579d7-hhlrw    2/2     Running   0          63s
quorum-node4-deployment-7667977997-r5z62   2/2     Running   0          63s
quorum-node5-deployment-bd8f859bb-qcxb4    2/2     Running   0          62s
quorum-node6-deployment-787d95bdb7-2k9tc   2/2     Running   0          62s
quorum-node7-deployment-89c58598d-5rbbl    2/2     Running   0          62s

# delete minikube k8s deployment
$> kubectl delete -f 7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc

```

### Shutdown minikube
```shell
$> minikube stop
$> minikube delete
```

### Deleting Kubernetes Deployment

1. Delete the kubernetes resources:
```shell
$> kubectl delete -f PATH/TO/K8S-YAML-DIR/
```
1a. **If using HostPath storage** Delete the files on the host machine, by default under `/var/lib/docker/geth-storage`.
example for minikube.
```
$> minikube ssh
$> sudo su
$> rm -r /var/lib/docker/geth-storage
```

## Accessing Quorum and Transaction Manager Containers on K8s

**note** this assumes that the quorum deployment was deployed to the `default` namespace or that namespace that the 
Quorum cluster was deployed to is set as the default, 
e.g. `kubectl config set-context $(kubectl config current-context) --namespace=$YOUR_NAMESPACE`

```shell
$> kubectl get pods
NAME                                       READY   STATUS    RESTARTS   AGE
quorum-node1-deployment-57b6588b6b-5tqdr   1/2     Running   1          40s
quorum-node2-deployment-5f776b479c-f7kxs   2/2     Running   2          40s

$> POD_NAME=$(kubectl get pods | grep node1 | awk '{print $1}')
$> kubectl  exec -it $POD_NAME -c quorum /bin/ash

# now you are inside the quorum container

> geth attach $QHOME/dd/geth.ipc
> eth.blockNumber
> 0

/> cd $QHOME/contracts
/> ./runscript.sh public_contract.js
/> ./runscript.sh private_contract.js

# you should now see the transactions go through
# note: if you are running IBFT (Istanbul BFT consensus) the blockNumber will automaticly increment at a steady 
(configured) time interval.

\> geth attach $QHOME/dd/geth.ipc
> eth.blockNumber
> 2

# show connected peers
> admin.peers.length
6

```

## Configuration Details
The main configuration files are [`qubernetes.yaml`](7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-7nodes.yaml) and [`nodes.yaml`](7nodes/nodes-7.yaml). 
These two configuration yaml files must exist in the base directory.


### Generating Kubernetes Object yaml From Existing Quorum Resources.

This example will demo (re)generating the Quorum Kubernetes yaml (tessera/IBFT) from the 7nodes quorum resources (keys, configs, etc.). 

There are qubernetes config files in the [7nodes](7nodes) directory for the various deployment configurations: tessera, constellation, IBFT, raft, PVC, host, etc.   
This example uses [7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-tessera-7nodes-pvc.yaml](7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-tessera-7nodes-pvc.yaml)

```yaml
nodes:
  number: 7
service:
  # NodePort | ClusterIP
  type: ClusterIP
quorum:
  # supported: raft | istanbul
  consensus: istanbul
  # base quorum data dir as set inside each container.
  Node_DataDir: /etc/quorum/qdata
  # This is where all the keys are store, and/or where they are generated, as in the case of quorum-keygen.
  # Either full or relative paths on the machine generating the config
  Key_Dir_Base: 7nodes
  Permissioned_Nodes_File: 7nodes/permissioned-nodes.json
  Genesis_File: 7nodes/istanbul-genesis.json
  # related to quorum containers
  quorum:
    Raft_Port: 50401
    # container images at https://hub.docker.com/u/quorumengineering/
    Quorum_Version: 2.2.3
  # related to transaction manager containers
  tm:
    # (tessera|constellation)
    # container images at https://hub.docker.com/u/quorumengineering/
    Name: tessera
    Tm_Version: 0.9.2
    Port: 9001
    Tessera_Config_Dir: 7nodes
  # for persistent storage can be host or Persistent Volume Claim.
  # The data dir is persisted here
  storage:
    # Host (requires hostPath) || PVC (Persistent_Volume_Claim - tested with GCP).
    Type: PVC
    ## when redeploying cannot be less than previous values
    Capacity: 200Mi
# generic geth related options
geth:
  Node_WSPort: 8546
  NodeP2P_ListenAddr: 21000
  network:
    # network id (1: mainnet, 3: ropsten, 4: rinkeby ... )
    id: 10
    # public (true|false) is it a public network?
    public: false
  # general verbosity of geth [1..5]
  verbosity: 9
```

Replace the `qubernetes.yaml` in the qubernetes base directory with [7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-tessera-7nodes-pvc.yaml](7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-tessera-7nodes-pvc.yaml), e.g. create a symlink to `qubernetes-istanbul-tessera-7nodes-pvc.yaml`
 
```
$> pwd
~qubernetes
$> rm qubernetes.yaml

## Create the symlink
$> ln -s 7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-tessera-7nodes-pvc.yaml qubernetes.yaml

$> ls -la qubernetes.yaml
qubernetes.yaml -> 7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-tessera-7nodes-pvc.yaml
 
```

Inside the `qubernetes.yaml` config file, the paths of the existing resources are set:
```yaml
  Key_Dir_Base: 7nodes
  Permissioned_Nodes_File: 7nodes/permissioned-nodes.json
  Genesis_File: 7nodes/istanbul-genesis.json
```
There are keys already generated in the 7nodes directory in sub directories: key1, key2 ... key7. These will be used when
generating the the kubernetes resource yaml.
```shell
$> ls 7nodes/key
key1/  key2/  key3/  key4/  key5/  key6/  key7/
```

The `permissioned-nodes.json` and `istanbul-genesis.json` already exist as well, and are also inside the 7nodes directory.
```shell
$> ls 7nodes/permissioned-nodes.json
7nodes/permissioned-nodes.json

$> ls 7nodes/istanbul-genesis.json
7nodes/istanbul-genesis.json
```

Finally, there needs to be a file `nodes.yaml`, in the qubernetes base directory, which specifies the key directory for each node.
nodes.yaml  
```shell
nodes:

- member:
    Node_UserIdent: quorum-node1
    Key_Dir: key1

- member:
    Node_UserIdent: quorum-node2
    Key_Dir: key2
...
```

When generating a fresh deployment, the `node.yaml` file will be generated, but since this example is creating 
Kubernetes API resources from existing resources this file needs to be present.

Create a symlink to link `7nodes/nodes-7.yaml` file to `nodes.yaml` in the base directory. 
```shell
$> ln -s 7nodes/nodes-7.yaml nodes.yaml
```

All set! Now we can run the `./qubernetes` command to (re) generate the kubernetes yaml, by default this command will create 
a directory `out` and place the generated file in the `out` directory:

```shell
 $> ./qubernetes
   
     Success!
   
     Quorum Kubernetes resource files have been generated in the `out/` directory.
   
     To deploy to kubernetes run:
   
     $> kubectl apply -f out
```

The Kubernetes resource yaml will now be inside the `out` directory, which should be the same as the files in the
`7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc` directory.
  
```shell
$> ls out
00-quorum-persistent-volumes.yaml 01-quorum-genesis.yaml            02-quorum-shared-config.yaml      03-quorum-services.yaml           04-quorum-keyconfigs.yaml         05-quorum-deployments.yaml

$> ls 7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc
00-quorum-persistent-volumes.yaml 01-quorum-genesis.yaml            02-quorum-shared-config.yaml      03-quorum-services.yaml           04-quorum-keyconfigs.yaml         05-quorum-deployments.yaml
```

To deploy the regenerated Kubernetes resource files to a running cluster:
```
$> kubectl apply -f out
persistentvolumeclaim/quorum-node1-quorum created
persistentvolumeclaim/quorum-node1-tm-pvc created
persistentvolumeclaim/quorum-node1-log-pvc created
persistentvolumeclaim/quorum-node2-quorum created
persistentvolumeclaim/quorum-node2-tm-pvc created
persistentvolumeclaim/quorum-node2-log-pvc created
persistentvolumeclaim/quorum-node3-quorum created
persistentvolumeclaim/quorum-node3-tm-pvc created
...
# when finished you can delete the cluster.
$> kubectl delete -f out
```

## Generating Quroum K8s Resources From Custom Configs

This section describes how to deploy a more customized kubernetes deployment with a varying number of quorum and transaction 
manager nodes, including generating the appropriate genesis config, required keys, services, etc. 

In order to do this, various tools need to be install, if you have Docker then you are all set! Use the [Docker Bootstrap Container](#docker-bootstrap-container)
if you do not wish to install Docker, follow the instruction in [Install Prerequisites without Docker](#install-prerequisites-without-docker).

Once you have the prerequists set up see [Generating Your Own K8s Resources](#generating-your-own-k8s-resources) for more 
information about configuring a custom deployment.

### Docker Bootstrap Container

To avoid installing all the prerequisites, you can use a docker container with all prerequisites already installed.

Usage:
```
docker run -ti quorumengineering/qubernetes
./quorum-init
# point kubectl to a correct cluster
# for example for k8s on gcloud do `gcloud init` and paste config command from "connect" button in UI
kubectl apply -f out
```
**note**: `qubernetes.yaml` is not added to the the docker container, as this file will change between various deployments.
It can by included by mounting a directory containing the desired `qubernetes.yaml` files. For example, if you have qubernetes 
checked out and with custom configs, you can mount it to a container adding `-v $(pwd):/qubernetes` to your `docker run` command:

```
$> docker run -v $(pwd):/qubernetes -ti quorumengineering/qubernetes
```

### Install Prerequisites without Docker 
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
 $> ln -s examples/config/qubernetes-istanbul-generate-8nodes.yaml qubernetes.yaml
 
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
 
 ```shell
 
 ## in this case, an out directory exists, so select `1`.
 $> ./quorum-init
 The 'out' directory already exist.
 Please select the action you wish to take:

 [1] Delete the 'out' directory and generate new resources.
 [2] Update / add nodes that don't already exist.
 [3] Cancel.
 
 ..
 
 Creating all new resources.
 
   Generating keys...
 INFO [01-14|17:05:09.402] Maximum peer count                       ETH=25 LES=0 total=25
 INFO [01-14|17:05:11.302] Maximum peer count                       ETH=25 LES=0 total=25
 INFO [01-14|17:05:13.160] Maximum peer count                       ETH=25 LES=0 total=25
```

 After the quorum  resources have been generated, the necessary k8s resources will be created in the `out` directory:
```shell
# if you want to check out the generated quorum resources
$> ls out/config
genesis.json                   key2                           key5                           key8                           tessera-config-9.0.json
istanbul-validator-config.toml key3                           key6                           nodes.yaml                     tessera-config-enhanced.json
key1                           key4                           key7                           permissioned-nodes.json        tessera-config.json

$> ls out
00-quorum-persistent-volumes.yaml 02-quorum-shared-config.yaml      04-quorum-keyconfigs.yaml         config
01-quorum-genesis.yaml            03-quorum-services.yaml           05-quorum-deployments.yaml

# deploy the resources
$> kubectl apply -f out
```

3. Once the Quorum resources have been generated, the `./quberetes` command can be run to generator variations of the Kubernetes
Resources, e.g. `ClusterIP` vs `NodePort`. The `./qubernetes` command can be run multiple times and is idempotent as long as the 
underlying Quorum resources do not change.

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
```
* **If using `hostStorage`** Delete the files on the host machine, by default under `/var/lib/docker/geth-storage`.

## Thanks! And Additional Resources 
Thanks to [Maximilian Meister blog and code](https://medium.com/@cryptoctl) which provided and awesome starting point!
and is a good read to undestand the different components.
