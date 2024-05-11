#!/bin/zsh

set -euo pipefail

if minikube start --profile=intelligence --cni=calico --driver=docker --memory=2g --cpus=2 --interactive=false --nodes=3; then
    echo "Minikube started successfully."
else
    echo "Failed to start Minikube." >&2
    exit 1
fi

minikube --profile=intelligence kubectl -- wait --for=condition=ready nodes --all --timeout=300s
minikube --profile=intelligence stop --schedule 12h
