#!/bin/zsh

set -euo pipefail

if ! command -v helm >/dev/null 2>&1; then
    echo "Helm is not installed. Please install it to proceed." >&2
    exit 1
fi

version_greater_equal() {
    [[ "$(print -l $1 $2 | sort -V | head -n 1)" == "$2" ]]
}

HELM_VERSION_REGEX='^v[0-9]+\.[0-9]+\.[0-9]+$'
HELM_REQUIRED_VERSION="v3.14.2"
HELM_INSTALLED_VERSION=$(helm version --short | sed 's/[+-].*//' | grep -E "$HELM_VERSION_REGEX")

if [ -z "$HELM_INSTALLED_VERSION" ]; then
    echo "Could not find Helm version." >&2
    exit 1
fi

if version_greater_equal "$HELM_INSTALLED_VERSION" "$HELM_REQUIRED_VERSION"; then
    echo "Helm version is sufficient ($HELM_INSTALLED_VERSION)."
else
    echo "Helm version is too low. Required: $HELM_REQUIRED_VERSION, but found: $HELM_INSTALLED_VERSION." >&2
fi
