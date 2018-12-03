#!/bin/bash

qctl="kubectl --namespace=quorum-test  --kubeconfig={PATH/TO/YOUR}/k8_config --insecure-skip-tls-verify "
$qctl create -f out/quorum-shared-config.yaml
$qctl create -f out/quorum-services.yaml
$qctl create -f out/quorum-keyconfigs.yaml
$qctl create -f out/quorum-deployments.yaml
