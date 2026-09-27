# k8s-watcher

A lightweight Kubernetes controller that watches Pod events (add/delete) across one or more namespaces and logs them. It uses `client-go` informers to efficiently stream Pod changes from the Kubernetes API server without polling.

## How it works

- On startup, `main.go` loads `config.yaml` (via [Viper](https://github.com/spf13/viper)) to determine which namespaces to watch.
- It builds a Kubernetes clientset (in-cluster config when running inside a Pod, or your local kubeconfig when running outside a cluster).
- For each configured namespace, [`watcher.WatchPod`](watcher/pod-watcher.go) starts a `SharedInformerFactory` scoped to that namespace and registers event handlers that log when a Pod is **added** or **deleted**.
- The process runs until it receives `SIGINT`/`SIGTERM`, at which point all informers are stopped gracefully.

## Configuration

Namespaces and resources to watch are configured in [`config.yaml`](config.yaml):

```yaml
watch:
  resources:
    - pods
  namespaces:
    - default
    - kube-system
notifier:
  type: stdout
```

- `watch.namespaces` — list of namespaces to watch for Pod deployment/deletion events.
- `watch.resources` — resource types to watch (currently only `pods` is implemented).
- `notifier.type` — how events are reported (currently events are printed to stdout).

## Running locally

Requires a valid kubeconfig with access to the target cluster.

```bash
go run . 
```

Make sure `config.yaml` is present in the working directory, then update `watch.namespaces` to the namespaces you want to observe.

## Running in-cluster

The [`Dockerfile`](Dockerfile) builds a minimal Alpine-based image running as a non-root `watcher` user.

```bash
docker build -t k8s-watcher:latest .
```

[`k8s-watcher-deployment.yaml`](k8s-watcher-deployment.yaml) deploys the watcher along with:

- A `ServiceAccount` (`pod-watcher-sa`, defined in [`sa.yaml`](sa.yaml)) used by the Deployment.
- A `Role`/`RoleBinding` granting the permissions needed to list/watch Pods.
- A `Deployment` (`k8s-watcher-deployment`) running the `k8s-watcher` image.
- A `ConfigMap` (`k8s-watcher-config`) mounted into the container as `/app/config.yaml`, supplying the namespaces/resources to watch.

Apply it to your cluster:

```bash
kubectl apply -f sa.yaml
kubectl apply -f k8s-watcher-deployment.yaml
```

## Project layout

```
main.go              # entrypoint: loads config, builds clientset, starts watchers
config/config.go      # config.yaml loading via Viper
pkg/kube/             # Kubernetes clientset construction (in-cluster / kubeconfig)
watcher/pod-watcher.go # per-namespace Pod informer + event handlers
config.yaml           # default watch configuration
sa.yaml               # ServiceAccount used by the deployment
k8s-watcher-deployment.yaml # RBAC, Deployment and ConfigMap manifests
Dockerfile            # multi-stage build producing the k8s-watcher image
```
