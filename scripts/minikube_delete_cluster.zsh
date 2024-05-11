#!/bin/zsh

set -euo pipefail

if ! minikube delete --profile=intelligence >/dev/null 2>&1; then
    echo "Minikube failed to delete the cluster." >&2
    exit 1
fi

echo "Deleted minikube cluster."
