#!/bin/bash

kubectl apply -f out/01-quorum-shared-config.yaml
kubectl apply -f out/02-quorum-services.yaml
kubectl apply -f out/03-quorum-keyconfigs.yaml
kubectl apply -f out/04-quorum-deployments.yaml
