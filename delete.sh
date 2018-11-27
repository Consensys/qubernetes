#!/bin/bash

qctl="kubectl --namespace=quorum-test  --kubeconfig=/home/libby/.go/src/github.com/ethereum/k8-quorum/k8_config --insecure-skip-tls-verify "

$qctl delete -f out/quorum-shared-config.yaml
$qctl delete -f out/quorum-services.yaml
$qctl delete -f out/quorum-deployments.yaml
$qctl delete -f out/quorum-keyconfigs.yaml
