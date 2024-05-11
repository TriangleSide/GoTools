#!/bin/zsh

set -euo pipefail

if ! command -v minikube >/dev/null 2>&1; then
    echo "Minikube is not installed. Please install it to proceed." >&2
    exit 1
fi

echo "Minikube is installed."

MINIKUBE_MINIMUM_VERSION=v1.33.0
MINIKUBE_CURRENT_VERSION=$(minikube version | grep 'minikube version:' | cut -d' ' -f3);

if [ "$MINIKUBE_CURRENT_VERSION" = "" ]; then
    echo "Could not parse minikube version." >&2
    exit 1
fi

if [ "$(printf '%s\n%s' "$MINIKUBE_MINIMUM_VERSION" "$MINIKUBE_CURRENT_VERSION" | sort -V | head -n1)" = "$MINIKUBE_MINIMUM_VERSION" ]; then
    echo "Minikube version $MINIKUBE_CURRENT_VERSION is sufficient (>= $MINIKUBE_MINIMUM_VERSION)."
else
    echo "Minikube version $MINIKUBE_CURRENT_VERSION is not sufficient. Please upgrade to at least $MINIKUBE_MINIMUM_VERSION." >&2
    exit 1
fi
