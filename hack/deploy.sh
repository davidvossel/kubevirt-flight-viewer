#!/bin/bash

kubectl apply -f artifacts/examples/kubevirtflightviewer.kubevirt.io_inflightoperations.yaml
kubectl apply -f artifacts/examples/controller-deployment.yaml 
