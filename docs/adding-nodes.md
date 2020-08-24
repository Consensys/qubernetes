## Adding New Nodes 

Assuming we have created a 4 node network with a config file named `qubernetes.yaml`
```bash
 ./qube-init --action=create qubernetes.yaml
```

1. Update the config file used to create the network with the additional nodes to add.
 
   **note** make sure the consensus of the new node matches the consensus of the existing network.  
```yaml
...
nodes: 
...
# adding new node
  - Node_UserIdent: quorum-node5
    Key_Dir: key5
    quorum:
      quorum:
        # supported: (raft | istanbul)
        consensus: istanbul
        Quorum_Version: 2.6.0
      tm:
        # (tessera|constellation)
        Name: tessera
        Tm_Version: 0.10.4
```

2. Run `./qube-init --action=update qubernetes.yaml`  which will generate the Quorum and K8s resources for the new node(s), as well as update the existing 
K8s resources, such as `permissioned-nodes.json` configMap to include the new nodes. 

## Raft Node
3. If adding a raft node, run `helpers/add_nodes_to_k8s.sh raft`

## Istanbul Node
3. If adding an istanbul node, run `helpers/add_nodes_to_k8s.sh istanbul`
 