## Qubernetes Command Line Tool: qctl 

A command line tool for creating, running and interacting with a K8s Quorum network. 
 
The commands intentionally try to do only one thing (Unix philosophy). They can be added to scripts to automate complete 
network creation / deletion, etc. 

### Install
```
> go get https://github.com/ConsenSys/qubernetes/qctl 
```

```
> qctl --help

NAME:
   qctl - command line tool for managing qubernetes network. Yay!

USAGE:
   qctl [global options] command [command options] [arguments...]

COMMANDS:
   log, logs      Show logs for [quorum, tessera, constellation], running on a specific pod
   init           creates a base qubernetes.yaml file which can be used to create a Quorum network.
   generate       options for generating base config / resources
   delete         options for deleting networks / resources
   deploy         options for deploying networks / resources to K8s
   geth           options for interacting with geth
   list, ls, get  options for listing resources
   add, ls, get   options for adding resources
   connect, c     connect to nodes / pods
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --namespace value, -n value, --ns value  The k8s namespace for the quorum network (default: "default") [$QUORUM_NAMESPACE]
   --help, -h                               show help (default: false)
``` 
Config and network variables, can be set in environment variables or passed in via commnad line flags.  

### Initialize A Quorum Network / Config 
```
$> qctl init 
Your Qubernetes config has been generated see:

  /Users/libby/go/src/github.com/libby/qctl/qubernetes.generate.yaml

The Quorum network values are:

  num nodes = 4
  consensus = istanbul
  quorumVersion = 2.6.0
  tmVersion = 0.10.4
  transacationManger = tessera
  chainId = 1000

To set this as your default config for future commands, run:

**********************************************************************************************

  $> export QUBE_CONFIG=/Users/libby/go/src/github.com/libby/qctl/qubernetes.generate.yaml

  $> qctl generate network

************************************************************************************************

$> export QUBE_CONFIG=/Users/libby/Workspace.Quorum/qubernetes-priv/qctl/qubernetes.generate.yaml
$> qctl generate network --create
...
  To enable future commands, e.g. qctl create network, qctl delete network, to use this network
  config, set the QUBE_K8S_DIR environment variable to the out directory that has just been generated
  by running:

*****************************************************************************************
  $> export QUBE_K8S_DIR=/Users/libby/go/src/github.com/libby/qctl/out
  $> qctl create network
*****************************************************************************************

$>  export QUBE_K8S_DIR=/Users/libby/Workspace.Quorum/qubernetes-priv/qctl/out

```

### Deploying The Qubernetes K8s Network
This requires a running K8s network, either local (kind, minikube, docker on desktop) or cloud provider (GKE, EKS, Azure),
or other managed K8s runtime.

```
$>  export QUBE_K8S_DIR=/Users/libby/Workspace.Quorum/qubernetes-priv/qctl/out
$> qctl create network

$> qctl delete network
``` 


### Modify The Network

```
## display the current config that is being used to generate the network
$> qctl ls config
$> qctl ls config --long

$> qctl add node --name=quorum-node5

## Generate the additional config
$> qctl generate network --update

## Deploy the new node to K8s
$> qctl  network
```

## Interacting With A Running Quorum Network

### Attach to geth on a specific node
```
> qctl geth attach node1
```

### Execute a geth command on a specific node
```
> qctl geth exec node1 'eth.blockNumber'
```

### Follow Quorum Logs
```
> qctl logs -f node1 quorum
```

### Follow Tessera Logs
```
> qctl logs -f node1 tessera 
```
