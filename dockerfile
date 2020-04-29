# BUILD
FROM golang:1.14 as builder

RUN apt update && apt install -y \
  curl gettext g++ git  

WORKDIR /workspace
RUN GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4
COPY . .
RUN go build -buildmode plugin -o SecretsFromVault.so ./SecretsFromVault.go 


# RUNTIME
FROM debian:10

RUN apt update && apt install --no-install-recommends -y \
  ca-certificates git\
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /workspace/SecretsFromVault.so /opt/kustomize/plugin/kustomize.rohde-schwarz.com/v1alpha1/secretsfromvault/SecretsFromVault.so
COPY --from=builder /go/bin/kustomize /usr/bin/kustomize

ENV XDG_CONFIG_HOME=/opt

ENTRYPOINT ["/usr/bin/kustomize", "build", "--enable_alpha_plugins"]