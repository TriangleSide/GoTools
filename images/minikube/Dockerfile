FROM debian:bookworm

# Base to install other components.
USER root
RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install -y ca-certificates curl

# Install the docker cli.
# https://docs.docker.com/engine/install/debian/
ARG DOCKER_CLIENT_VERSION=MUST_SET
RUN curl -Lo /tmp/docker-cli.deb https://download.docker.com/linux/debian/dists/bookworm/pool/stable/amd64/docker-ce-cli_${DOCKER_CLIENT_VERSION}-1~debian.12~bookworm_amd64.deb
RUN dpkg -i /tmp/docker-cli.deb

# Install minikube.
# https://minikube.sigs.k8s.io/docs/start/
ARG MINIKUBE_VERSION=MUST_SET
RUN curl -Lo /usr/local/bin/minikube https://storage.googleapis.com/minikube/releases/${MINIKUBE_VERSION}/minikube-linux-amd64
RUN chmod +x /usr/local/bin/minikube

# Set some setting for minikube, such as the home folder.
ENV MINIKUBE_IN_STYLE=0
ENV MINIKUBE_HOME=/

# Ensure the permissions are copied so the docker unix socket can be used.
ARG DOCKER_GID=MUST_SET
ARG USER_UID=MUST_SET
ARG USER_GID=MUST_SET
RUN groupadd -g ${DOCKER_GID} docker
RUN groupadd -g ${USER_GID} minikube
RUN useradd -u ${USER_UID} -g ${USER_GID} -G ${DOCKER_GID} minikube
USER minikube:docker

ENTRYPOINT ["minikube"]
