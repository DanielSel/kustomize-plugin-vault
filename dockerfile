### BUILD IMAGE ###
ARG GO_VERSION=1.14
FROM golang:${GO_VERSION} as builder
ARG KZ_VERSION=v3.8.1
ENV KZ_VERSION=$KZ_VERSION

RUN apt update && apt install -y \
  curl gettext g++ git  

WORKDIR /workspace

# Dependencies
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Upstream Kustomize - required for plugin to work
RUN git clone -b kustomize/${KZ_VERSION} https://github.com/kubernetes-sigs/kustomize.git &&\
    mkdir -p bin && cd kustomize/kustomize &&\
    CGO_ENABLED=1 go build \
      # -ldflags="-X sigs.k8s.io/kustomize/api/provenance.version=${KZ_VERSION} -X sigs.k8s.io/kustomize/api/provenance.gitCommit=$(git rev-parse HEAD) -X sigs.k8s.io/kustomize/api/provenance.buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
      -ldflags="-w -s -X sigs.k8s.io/kustomize/api/provenance.version=${KZ_VERSION} -X sigs.k8s.io/kustomize/api/provenance.gitCommit=$(git rev-parse HEAD) -X sigs.k8s.io/kustomize/api/provenance.buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
      -o ../../bin/kustomize

# Source Code
COPY main.go main.go
COPY pkg/ pkg/

# Build
RUN mkdir -p bin &&\
    CGO_ENABLED=1 go build \
      -buildmode plugin \
      -ldflags="-w -s" \
      -o bin/SecretsFromVault.so main.go 

### RUNTIME IMAGE ###
FROM bitnami/kubectl:1.17 as kubectl
FROM debian:10

RUN apt update && apt install --no-install-recommends -y \
  ca-certificates git curl gettext \
    && rm -rf /var/lib/apt/lists/*

COPY --from=kubectl /opt/bitnami/kubectl/bin/kubectl /usr/bin/kubectl
COPY --from=builder /workspace/bin/SecretsFromVault.so /opt/kustomize/plugin/kustomize.rohde-schwarz.com/v1alpha1/secretsfromvault/SecretsFromVault.so
COPY --from=builder /workspace/bin/kustomize /usr/bin/kustomize

ENV XDG_CONFIG_HOME=/opt

ENTRYPOINT ["/usr/bin/kustomize", "build", "--enable_alpha_plugins"]
