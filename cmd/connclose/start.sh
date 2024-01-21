#!/bin/bash

set -ex

kind create cluster --config kind.yml
docker pull quay.io/cilium/cilium:v1.14.6
kind load docker-image quay.io/cilium/cilium:v1.14.6

helm install cilium cilium/cilium --version 1.14.6 \
   --namespace kube-system \
   --set image.pullPolicy=IfNotPresent \
   --set ipam.mode=kubernetes

cilium status --wait
