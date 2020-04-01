## 7 Nodes Examples

[quorum-examples 7nodes](https://github.com/jpmorganchase/quorum-examples/tree/master/examples/7nodes) has been ported 
to k8s resources.  There are k8s resource files in the qubernetes repo's [7nodes](7nodes) directory for deploying 
quorum on kubernetes with tessera or constellation as the transaction manager, and raft or istanbul as the consensus engines.  

This assume you have a running k8s cluster which you can connect to via `kubectl`.

For presistent storage `PVC`(Persistent Volume Claims) are used/recommended, `PVC` will be automatically created when 
the deployment is created and deleted when the deployment is deleted (`kubectl delete -f YOUR_DEPLOYMENT_YAML/`).


**note** [HostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath) persistent storage is no longer supported 
after commit [536e1084e362cb3db87003c36f1fdffaa4f9da64](commit/536e1084e362cb3db87003c36f1fdffaa4f9da64) Wed Mar 11 17:03:21 2020 -0400

Examples below are for deploying 7nodes using PVC (Persistent Volume Claims):

* [istanbul tessera k8s resource yaml (PVC)](../7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc)
```shell
$> kubectl apply -f 7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc
$> kubectl delete -f 7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc
```
* [istanbul constellation k8s resource yaml (PVC)](../7nodes/istanbul-7nodes-constellation/k8s-yaml-pvc)
```shell
$> kubectl apply -f 7nodes/istanbul-7nodes-constellation/k8s-yaml-pvc
$> kubectl delete -f 7nodes/istanbul-7nodes-constellation/k8s-yaml-pvc
```
* [raft tessera k8s resource yaml (PVC)](../7nodes/raft-7nodes-tessera/k8s-yaml-pvc)
```shell
$> kubectl apply -f 7nodes/raft-7nodes-tessera/k8s-yaml-pvc
$> kubectl delete -f 7nodes/raft-7nodes-tessera/k8s-yaml-pvc
```
* [raft constellation k8s resource yaml (PVC)](../7nodes/raft-7nodes-constellation/k8s-yaml-pvc)
```shell
$> kubectl apply -f 7nodes/raft-7nodes-constellation/k8s-yaml-pvc
$> kubectl delete -f 7nodes/raft-7nodes-constellation/k8s-yaml-pvc
```
