#!/bin/bash

qctl="kubectl --namespace=quorum-test  --kubeconfig={PATH/TO/YOUR}/k8_config --insecure-skip-tls-verify "

$qctl delete -f out/quorum-shared-config.yaml
$qctl delete -f out/quorum-services.yaml
$qctl delete -f out/quorum-deployments.yaml
$qctl delete -f out/quorum-keyconfigs.yaml
