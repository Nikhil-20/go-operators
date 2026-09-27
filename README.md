# go-operators

A collection of Kubernetes operators and tooling written in Go using [Kubebuilder](https://book.kubebuilder.io/) and [client-go](https://github.com/kubernetes/client-go).

## Projects

| Project | Description |
|---|---|
| [db-operator](db-operator/) | Kubebuilder operator that manages a `Database` custom resource, provisioning a PostgreSQL StatefulSet, Service and Secret from the CR spec (image, replicas, storage, credentials). |
| [hello-world-operator](hello-world-operator/) | Kubebuilder operator that manages a `HelloWorld` custom resource, used as a minimal example/scaffold for building new operators. |
| [k8s-watcher](k8s-watcher/) | Standalone controller (no CRD) that uses `client-go` informers to watch Pod add/delete events across configured namespaces and logs them. |

Each project is a self-contained Go module with its own `go.mod`, `Dockerfile`, and Kubernetes manifests under `config/` (or plain YAML for `k8s-watcher`). See each project's own README for setup and deployment instructions.

## Prerequisites

- Go 1.24+
- Docker
- kubectl
- Access to a Kubernetes cluster (e.g. kind/minikube for local development)

## Repository layout

```
db-operator/            # Database CR operator (Kubebuilder)
hello-world-operator/   # HelloWorld CR operator (Kubebuilder)
k8s-watcher/            # Pod-watching controller (client-go informers)
```
