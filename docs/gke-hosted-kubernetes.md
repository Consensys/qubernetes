## GKE (Google Kubernetes Engine) 

Running on GKE inside GCP (Google Cloud Platform)

### Docker Container

The GCP SDK is already installed inside the docker container `docker run -ti quorumengineering/qubernetes`
 
### Prerequisites

1. GCP account: https://cloud.google.com/free/
2. GCP SDK: https://cloud.google.com/sdk/docs/quickstart-macos

### Enable Kubernetes Engine

* [Creating GKE Cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/creating-a-cluster)
* [Enabling Kubernetes Engine](https://console.cloud.google.com/apis/library/container.googleapis.com?q=kubernetes%20engine) 

### Creating a Kubernetes Cluster with GCP SDK 
```
## select an existing project and zone.
$> gcloud init  

$> gcloud config list project

$> gcloud container clusters list

# all lowercase for name
$> gcloud container clusters create q1
...
NAME       LOCATION    MASTER_VERSION  MASTER_IP     MACHINE_TYPE   NODE_VERSION   NUM_NODES  STATUS
q1         us-east1-b  1.12.8-gke.10   34.73.114.36  n1-standard-1  1.12.8-gke.10  3          RUNNING

# or with a default region
$> gcloud container clusters create q1 --region=us-east1 

# or with default zone
$> gcloud container clusters create q1 --zone=us-east1-b 
NAME  LOCATION    MASTER_VERSION  MASTER_IP     MACHINE_TYPE   NODE_VERSION    NUM_NODES  STATUS
q1    us-east1-b  1.13.11-gke.14  35.231.41.38  n1-standard-1  1.13.11-gke.14  3          RUNNING

# delete cluster (include zone or region)
$> gcloud container clusters delete q1 --zone=us-east1-b
```

### Setting up kubectl 
```
> gcloud container clusters list
NAME  LOCATION    MASTER_VERSION  MASTER_IP     MACHINE_TYPE   NODE_VERSION    NUM_NODES  STATUS
q1    us-east1-b  1.13.11-gke.14  35.231.41.38  n1-standard-1  1.13.11-gke.14  3          RUNNING

# this will set the context and access enbabling the kubectl cmd to work with the GKE cluster
> gcloud container clusters get-credentials q1 --zone=us-east1-b

> kubectl get pods
No resources found.

# deploy a quorum network to the GKE cluster
> kubectl apply -f 7nodes/istanbul-7nodes-tessera/k8s-yaml-pvc

```

### Troubleshooting
    
1. pods are crashing right away `Init:CrashLoopBackOff`.

* Are persistent volume mounts (PVM) being used? This issue might be due to `hostPath` as the storage backend,
and redeploying the cluster. Try using PVM instead.
  
### Revert to Previous gcloud Versions
```
$> gcloud components update --version 261.0.0
```

### Status
```
$> gcloud container clusters list
NAME  LOCATION    MASTER_VERSION  MASTER_IP     MACHINE_TYPE   NODE_VERSION    NUM_NODES  STATUS
q1    us-east1-b  1.13.11-gke.14  35.231.41.38  n1-standard-1  1.13.11-gke.14  3          RUNNING

$> gcloud container clusters describe q1 --zone=us-east1-b
```

### Delete `hostpath`

* When using `hostpath` as the storage, delete the cluster and recreate it before deploying another cluster - this ensures 
that the underlying storage has been deleted as well.
* `gcloud container clusters delete $NAME`
