#!/bin/bash

kubectl apply -f artifacts/examples/crd-status-subresource.yaml
kubectl apply -f artifacts/examples/controller-deployment.yaml 
