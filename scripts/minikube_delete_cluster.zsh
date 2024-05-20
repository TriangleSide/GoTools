#!/bin/zsh

set -euo pipefail

if ! minikube delete --profile=intelligence >/dev/null 2>&1; then
    echo "Failed to delete this project's minikube cluster." >&2
    exit 1
fi

echo "Deleted the minikube cluster for this project."
