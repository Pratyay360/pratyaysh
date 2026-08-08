# pratyaysh

An SSH-hosted interactive resume/TUI.

```
ssh -p 2222 user@pratyaysh
```

## Deploying

The app runs as a containerized SSH server. The image is published to
`ghcr.io/pratyay360/pratyaysh:latest` by GitHub Actions.

The host key lives in `.ssh/ssh` (generated automatically by the Makefile).
It must be mounted into the container at `/.ssh` — the server binds port 2222.

### Option A: Podman pod (also works on Kubernetes)

Uses `podman play kube` with a Pod manifest, so the same YAML works with
`kubectl` too.

```sh
make pod          # generate host key + render manifest + play kube
make pod-down     # stop and remove the pod
```

The rendered manifest is `kube/pratyaysh-pod.yaml`, generated from
`kube/pratyaysh-pod.yaml.tpl`. To deploy to Kubernetes:

```sh
kubectl apply -f kube/pratyaysh-pod.yaml
```

### Option B: podman-compose

`compose.yml` uses a single service; podman-compose wraps it in an
auto-named pod (`pod_pratyaysh`).

```sh
make compose          # podman-compose up -d
make compose-down     # podman-compose down
```

## Building locally

```sh
mise run build
podman build -f Containerfile -t pratyaysh .
```

Generate the host key with `ssh-keygen -f ssh` (or `make ssh-key`).
