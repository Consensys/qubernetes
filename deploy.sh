#!/bin/bash

#qctl="kubectl --namespace=quorum-test  --kubeconfig=/home/libby/.go/src/github.com/ethereum/k8-quorum/k8_config --insecure-skip-tls-verify "
#qctl="kubectl  --kubeconfig=/home/libby/Workspace.JPMC/kubernetes/k8s-katas/config/k8s_config --insecure-skip-tls-verify "
qctl='kubectl --namespace=quorum-test  --kubeconfig=/home/libby/.go/src/github.com/ethereum/k8-quorum/k8_config_east2 --insecure-skip-tls-verify '
$qctl create -f out/quorum-shared-config.yaml
$qctl create -f out/quorum-services.yaml
$qctl create -f out/quorum-keyconfigs.yaml
$qctl create -f out/quorum-deployments.yaml
