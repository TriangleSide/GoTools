#!/bin/zsh

set -euo pipefail

if ! command -v minikube >/dev/null 2>&1; then
    echo "Minikube is not installed. Please install it to proceed." >&2
    exit 1
fi

version_greater_equal() {
    [[ "$(print -l $1 $2 | sort -V | head -n 1)" == "$2" ]]
}

MINIKUBE_VERSION_REGEX='^v[0-9]+\.[0-9]+\.[0-9]+$'
MINIKUBE_MINIMUM_VERSION=v1.33.0
MINIKUBE_CURRENT_VERSION=$(minikube version | awk '/minikube version:/ {print $3}' | grep -E "$MINIKUBE_VERSION_REGEX")

if [[ -z "$MINIKUBE_CURRENT_VERSION" ]]; then
    echo "Could not find Minikube version." >&2
    exit 1
fi

if version_greater_equal "$MINIKUBE_CURRENT_VERSION" "$MINIKUBE_MINIMUM_VERSION"; then
    echo "Minikube version is sufficient ($MINIKUBE_CURRENT_VERSION)."
else
    echo "Minikube version is too low. Required: $MINIKUBE_MINIMUM_VERSION, but found: $MINIKUBE_CURRENT_VERSION."
fi
