# Development

Before getting started, ensure the prerequisite [dependencies](./DEPENDENCIES.md) are installed.

## Running a local environment

To create a local kubernetes cluster with the application installed, run the following steps:

- `make minikube_start_cluster`
- `make helm_install_charts`

To teardown the cluster created with the steps defined above, run:

- `make minikube_delete_cluster`
