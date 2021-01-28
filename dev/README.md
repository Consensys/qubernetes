## Local Developing and Testing with `qctl` 

**note:** this requires [minikube](https://minikube.sigs.k8s.io/docs/start/).

Developing and testing local quorum code requires that we can build local docker containers which minikube can access.
We will go through the steps to do this:

1. [Make sure minikube is installed and start it up](#step1)
1. [Setup your local quorum project for building local minikube containers](#step2)
2. [Build your local image for minikube](#step3)
3. [Run a Quorum network with `qctl` and a local container](#step4)
4. [Run a Mixed Quorum Network (local container nodes, and versioned nodes)](#step5)

#### <a name="step1"></a> 1. Make sure minikube is installed and start it up.
install [minikube](https://minikube.sigs.k8s.io/docs/start/).  

start minikube
```
$> minikube start --memory 6144 
```

#### <a name="step2"></a> 2. Setup your local quorum project for building local minikube containers

from the quberentes dev directory
```
$> cd qubernetes/dev
```
clone the Quorum repo you wish to work with
```
$> git clone https://github.com/ConsenSys/quorum.git
```
copy over the helper scripts and Dockerfiles
```
$> cp docker-helpers/* quorum
```
checkout the branch you wish to work with
```
$> cd quorum
$> git checkout $YOUR_BRANCH
```

#### <a name="step3"></a> 3. Build your local image for minikube

build the containers locally the script will set the docker env to the minikube docker env, `eval $(minikube docker-env)`

from the `qubernetes/dev` dir 
```
$> cd qubernetes/dev
```

run the build script
```
$> ./quorum-build-all.sh
```

check that the images have been build and are in the minikube docker env:
```
$> eval $(minikube docker-env)
$> docker images
REPOSITORY                                TAG           IMAGE ID       CREATED              SIZE
quorum-local                              latest        01dbfb31dc13   About a minute ago   1.29GB
quorum-base-local                         latest        882a062e4017   2 minutes ago        946MB
...
k8s.gcr.io/etcd                           3.4.3-0       303ce5db0e90   15 months ago        288MB
k8s.gcr.io/pause                          3.1           da86e6ba6ca1   3 years ago          742kB
gcr.io/k8s-minikube/storage-provisioner   v1.8.1        4689081edb10   3 years ago          80.8MB
```

#### <a name="step4"></a> 4. Run a Quorum network with `qctl` and a local container 

Create a Quorum network using your local container image. 

Use `qctl` with the `--qimagefull` flag set to the name of your local image, e.g. `--qimagefull=quorum-local`

open a new tab or set the docker context back to your host
```
$> eval $(minikube docker-env --unset)
```

from your `qubernetes/dev` directory.
```
$> cd qubernetes/dev
```

create a subdirectory `istanbul-local` inside of `qubernetes/dev`
```
$> mkdir istanbul-local 
```

export the global env vars required for `qctl`
```
$> export QUBE_CONFIG=$(pwd)/istanbul-local/qubernetes.generate.yaml
$> export QUBE_K8S_DIR=$(pwd)/istanbul-local/out
```

create a network with your local quorum image
```
$> qctl init --consensus=istanbul --qimagefull=quorum-local --num 3
$> qctl generate network --create
$> qctl deploy network --wait 
```

test the block number
```
$> qctl geth exec quorum-node1 'eth.blockNumber'
```

#### Look inside the generated minimal config `qubernetes.generate.yaml`
The configuration we generate above (`qubernetes/dev/istanbul-loca/qubernetes.generate.yaml`) should have the 
`Docker_Repo_Full` field set. 
When `Docker_Repo_Full` is set, the `Quorum_Version` field is ignored and the local container (`quorum-local`) is used instead.

**example** `qubernetes/dev/istanbul-loca/qubernetes.generate.yaml`:
```
genesis:
  consensus: istanbul
  Quorum_Version: 2.7.0
  Tm_Version: 0.10.6
  Chain_Id: "1000"
nodes:
- Node_UserIdent: quorum-node1
  Key_Dir: key1
  quorum:
    quorum:
      consensus: istanbul
      Quorum_Version: 2.7.0
      Docker_Repo_Full: quorum-local
    tm:
      Name: tessera
      Tm_Version: 0.10.6
  geth:
    Geth_Startup_Params: ""
- Node_UserIdent: quorum-node2
  Key_Dir: key2
  quorum:
    quorum:
      consensus: istanbul
      Quorum_Version: 2.7.0
      Docker_Repo_Full: quorum-local
    tm:
      Name: tessera
      Tm_Version: 0.10.6
  geth:
    Geth_Startup_Params: ""
- Node_UserIdent: quorum-node3
  Key_Dir: key3
  quorum:
    quorum:
      consensus: istanbul
      Quorum_Version: 2.7.0
      Docker_Repo_Full: quorum-local
    tm:
      Name: tessera
      Tm_Version: 0.10.6
  geth:
    Geth_Startup_Params: ""
```

#### <a name="step5"></a> 5. Run a Mixed Quorum Network (local container nodes, and versioned nodes)

We can modify the above `qubernetes.generate.yaml` to test a mixed net where `quorum-node1` runs 2.7.0 and 
`quorum-node2` and `quorum-node3` run our local container, by removing `Docker_Repo_Full` from `- Node_UserIdent: quorum-node1`

**example mixed net** `qubernetes/dev/istanbul-loca/qubernetes.generate.yaml`:
```
genesis:
  consensus: istanbul
  Quorum_Version: 2.7.0
  Tm_Version: 0.10.6
  Chain_Id: "1000"
nodes:
- Node_UserIdent: quorum-node1
  Key_Dir: key1
  quorum:
    quorum:
      consensus: istanbul
      Quorum_Version: 2.7.0
    tm:
      Name: tessera
      Tm_Version: 0.10.6
  geth:
    Geth_Startup_Params: ""
- Node_UserIdent: quorum-node2
  Key_Dir: key2
  quorum:
    quorum:
      consensus: istanbul
      Quorum_Version: 2.7.0
      Docker_Repo_Full: quorum-local
    tm:
      Name: tessera
      Tm_Version: 0.10.6
  geth:
    Geth_Startup_Params: ""
- Node_UserIdent: quorum-node3
  Key_Dir: key3
  quorum:
    quorum:
      consensus: istanbul
      Quorum_Version: 2.7.0
      Docker_Repo_Full: quorum-local
    tm:
      Name: tessera
      Tm_Version: 0.10.6
  geth:
    Geth_Startup_Params: ""
```

After modifying the configuration, the network will have to be generate and deployed again
```
$> qctl generate network --create
$> qctl deploy network --wait
```


### Subsequent local builds
Once you have the base quorum image build, e.g. `quorum-base-local` build when running `./quorum-build-all.sh` in step 2 
for faster subsequent builds when you make changes to the Quorum code, use: `./build-quorum.sh`
```
> ./build-quorum.sh
```

### Building multiple local images
`./build-quorum.sh MY_QUORUM_IMAGE` will build a docker image with the name set to the passed in param.
This is useful when you wish to run multiple local docker images, e.g. adding logging,testing across different branches, 
running a quorum network with multiple different quorum containers, etc. 