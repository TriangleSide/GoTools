#!/bin/zsh

set -euo pipefail

for chart in ./charts/*(/); do
    echo "Linting $chart"
    if ! helm lint --with-subcharts --quiet --strict "$chart"; then
        echo "Helm lint failed for $chart." >&2
        exit 1
    fi
done

echo "All charts successfully linted."
