#!/bin/zsh

set -euo pipefail

if [ "$#" -ne 1 ]; then
    echo "Error: Exactly one argument is required." >&2
    exit 1
fi

allowed_environment_values=("local" "dev" "prod")
if [[ ! " ${allowed_environment_values[*]} " =~ " $1 " ]]; then
    echo "Invalid argument. Allowed arguments are: [${allowed_environment_values[*]}]." >&2
    exit 1
fi

echo "Applying helm charts for the $1 environment."

charts_to_install=("cni" "cni-default-policy")
helm_upgrade_extra_options=(--install --atomic --cleanup-on-fail --wait --timeout 5m0s --qps 5 --history-max 3)
helm_test_extra_options=()

for chart in "${charts_to_install[@]}"; do
    echo "Installing the $chart chart."
    chart_dir="./charts/$chart"

    set -x

    if ! helm dependency update "$chart_dir"; then
        set +x
        echo "Helm dependency failed for chart $chart." >&2
        exit 1
    fi

    if ! helm upgrade "$chart" "$chart_dir" -f "$chart_dir/values.yaml" -f "$chart_dir/values-$1.yaml" $(<"$chart_dir/helm_options.txt") "${(@)helm_upgrade_extra_options}"; then
        set +x
        echo "Helm install failed for chart $chart." >&2
        exit 1
    fi

    set +x
done

echo "All charts have been successfully installed."
