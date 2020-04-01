
## Configuration Details
The main configuration files are [`qubernetes.yaml`](../7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-7nodes.yaml) and [`nodes.yaml`](../7nodes/nodes-7.yaml). 
These two configuration yaml files must exist in the base directory.


### Generating Kubernetes Object yaml From Existing Quorum Resources.

This example will demo (re)generating the Quorum Kubernetes yaml (tessera/IBFT) from the 7nodes quorum resources (keys, configs, etc.). 

There are qubernetes config files in the [7nodes](../7nodes) directory for the various deployment configurations: tessera, constellation, IBFT, raft, PVC, host, etc.   
This example uses [7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-tessera-7nodes-pvc.yaml](../7nodes/istanbul-7nodes-tessera/qubernetes-istanbul-tessera-7nodes-pvc.yaml)

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
    Quorum_Version: 2.5.0
  # related to transaction manager containers
  tm:
    # (tessera|constellation)
    # container images at https://hub.docker.com/u/quorumengineering/
    Name: tessera
    Tm_Version: 0.11
    Port: 9001
    Tessera_Config_Dir: 7nodes
  # persistent storage is handled by Persistent Volume Claims (PVC) https://kubernetes.io/docs/concepts/storage/persistent-volumes/
  # test locally and on GCP
  storage:
    # PVC (Persistent_Volume_Claim - tested with GCP).
    Type: PVC
    ## when redeploying cannot be less than previous values
    Capacity: 200Mi
# generic geth related options
geth:
  Node_RPCPort: 8546
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