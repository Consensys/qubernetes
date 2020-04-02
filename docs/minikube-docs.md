## Quickstart With Minikube

This section demonstrates how to deploy a 7 node quorum network to [minikube](https://github.com/kubernetes/minikube) a local
kubernetes cluster. 

Once the pods are deployed, if you wish to interact with them see [Accessing Quorum and Transaction Manager Containers on K8s](..#accessing-nodes-on-k8s).

### Install [minikube](https://kubernetes.io/docs/setup/minikube/) for your distro.
```shell
$> brew install minikube
```

### Once Minikube is install: Start local minikube cluster and deploy 7nodes example. 

For this example we will deploy [istanbul tessera k8s resource yaml (PVC)](../7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc).

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