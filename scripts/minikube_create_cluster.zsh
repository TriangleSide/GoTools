#!/bin/zsh

set -euo pipefail

minikube start \
    --profile=intelligence \
    --driver=docker \
    --memory=2g \
    --cpus=2 \
    --interactive=false \
    --nodes=3 \
    --cni=false \
    --network-plugin=cni \
    --extra-config=kubeadm.pod-network-cidr=192.168.0.0/16 \
    --subnet=172.16.0.0/24

if [ $? -ne 0 ]; then
    echo "Failed to start Minikube." >&2
    exit 1
fi

echo "Minikube started successfully."

if ! minikube --profile=intelligence stop --schedule 12h; then
    echo "Failed to set the schedule stop for Minikube." >&2
    exit 1
fi
