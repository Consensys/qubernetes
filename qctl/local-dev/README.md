## Developing with qctl and a local minikube environment

This is for Quorum developers who wish to build and test quorum networks locally.  With this approach, it is possible
to easily test quorum networks with different versions, e.g. 3 nodes running quorum 2.7; with 1 node running your local changes.

#### 1) Checkout Quorum locally 
```
$> git clone https://github.com/ConsenSys/quorum.git
$> cd quorum
# note the directory where you checkout Quorum
$> pwd
PATH/TO/QUORUM
```
#### 2) Copy the Dockerfiles from this directory to the Quorum directory 
This step is optional and only an optimization that breaks apart the base Dockerfile, to make subesequent docker builds
faster.
```
$> cp Dockerfile.gethbase PATH/TO/QUORUM
$> cp Dockerfile.gethonly PATH/TO/QUORUM
```

#### 3) Startup Minikube  
```
$> minikube start --memory 8192
```
#### 4) Once Minikube is start, set the docker-env to the minikube docker-env
**note**: This needs to be done in the same tab where you intend to build the container, and where you have the Quorum project.
```
$> cd PATH/TO/QUORUM
$> eval $(minikube docker-env)
``````


#### 5) build both containers so minikube can access them
**note**: make sure you are running this in the same tab that you ran `eval $(minikube docker-env)` in.
```
$> cd PATH/TO/QUORUM
$> docker build -t quorum-base-local -f Dockerfile.gethbase . && docker build -t quorum-local -f Dockerfile.gethonly .

# your docker should now be set to minikubes docker env
$> docker images

REPOSITORY                                TAG                 IMAGE ID            CREATED             SIZE
quorum-local                              latest              d785e7bf7ac4        21 seconds ago      1.27GB
quorum-base-local                         latest              e5ca3e234149        42 seconds ago      936MB
...
k8s.gcr.io/kube-proxy                     v1.17.2             cba2a99699bd        8 months ago        116MB
k8s.gcr.io/kube-apiserver                 v1.17.2             41ef50a5f06a        8 months ago        171MB
```

The quorum-base-local only needs to be build once, after that when you make changes to your go-ethereum / quorum code, you only need to build
quorum-local for the changes to be picked up.
```
$> docker build -t quorum-local -f Dockerfile.gethonly .
```

## Create and Deploy a Mixed Version Quorum Network 

* 3 - Quorum 2.7.0 nodes
* 1 - Quorum 2.6.0 node
* 1 - Quorum quorum-local node (as described in the previous steps)

Initalize the network with 3 - Quorum 2.7.0 nodes
```
> qctl init --qversion=2.7.0 --num=3
> qctl ls nodes --quorumversion
```
Before starting the network for the first time, we'll also add a 2.6.0 Quorum node to the mix.
```
> qctl add node --name=quorum-node4  --qversion=2.6.0
> qctl ls nodes --quorumversion
config currently has 4 nodes

     [quorum-node1] unique name
     [quorum-node1] quorumVersion: [2.7.0]


     [quorum-node2] unique name
     [quorum-node2] quorumVersion: [2.7.0]


     [quorum-node3] unique name
     [quorum-node3] quorumVersion: [2.7.0]


     [quorum-node4] unique name
     [quorum-node4] quorumVersion: [2.6.0]
```

Now we can generate the resources and deploy the network
```
> qctl generate network --create
> qctl deploy network --wait
```
connect to node1 and you should see Quorum Version: 2.7.0
```
> qctl connect quorum-node1
INFO[0000] trying to connect pods [quorum-node1-deployment-d89bb46bf-csnv8]
/ # geth version
Geth
Version: 1.9.7-stable
Git Commit: 6005360c90b636ba0fdc5a18ab308b3df2aa289f
Git Commit Date: 20200715
Quorum Version: 2.7.0
Architecture: amd64
Protocol Versions: [64 63]
Network Id: 1337
Go Version: go1.13.13
Operating System: linux
GOPATH=
GOROOT=/usr/local/go
```
connect to node4 and you should see Quorum Version: 2.6.0
```
> qctl connect quorum-node4
/ # geth version
Geth
Version: 1.9.7-stable
Git Commit: 9339be03f9119ee488b05cf087d103da7e68f053
Git Commit Date: 20200504
Quorum Version: 2.6.0
Architecture: amd64
Protocol Versions: [64 63]
Network Id: 1337
Go Version: go1.13.10
Operating System: linux
GOPATH=
GOROOT=/usr/local/go
```

Add Your local node (quorum-local) that was built in the previous steps, e.g. `docker build -t quorum-local -f Dockerfile.gethonly` 
```
> qctl add node --name=quorum-node4 --qimagefull=quorum-local
```
list all the nodes and notice that the `quorumImage` field is set on quorum-node5 
`[quorum-node5] quorumImage: [quorum-local]` this will be used when deploying the node.  
**note**: `quorumVersion` is ignored if `quorumImage` is set. 
```
> qctl ls nodes --all
...
  [quorum-node5] unique name
     [quorum-node5] keydir: [key-quorum-node5]
     [quorum-node5] quorumVersion: [2.7.0]
     [quorum-node5] txManager: [tessera]
     [quorum-node5] tmVersion: [0.10.4]
     [quorum-node5] quorumImage: [quorum-local]
     [quorum-node5] geth params: []
...
```
generate the resources/config/keys for the new node and deploy it to the running network
```
> qctl generate network --update
> qctl deploy network --wait
> qctl test contract quorum-node5
> qctl geth exec quorum-node5 'eth.blockNumber'
```
connect to node5 (this will be whatever version your local build was set to) 
```
> qctl connect quorum-node5
/go # geth version
Geth
Version: 1.9.7-stable
Git Commit: 762ee9f60d945db2c93dac22dc8f7f2f469c8943
Git Commit Date: 20200924
Quorum Version: 2.7.0
Architecture: amd64
Protocol Versions: [64 63]
Network Id: 1337
Go Version: go1.13.15
Operating System: linux
GOPATH=/go
GOROOT=/usr/local/goo
```

You should now have a mixed network running with: 
* 3 - Quorum 2.7.0 nodes
* 1 - Quorum 2.6.0 node
* 1 - Quorum quorum-local node (as described in the previous steps)