# pratyaysh

The app runs as a containerized SSH server. The image is published to
`ghcr.io/pratyay360/pratyaysh:latest` by GitHub Actions.
The host key lives in `.ssh/ssh` (generated automatically ).
It must be mounted into the container at `/.ssh` — the server binds port 2222.

## Building locally

```sh
mise run build
podman build -f Containerfile -t pratyaysh .
```

Generate the host key with `ssh-keygen -f ssh` (or `make ssh-key`).
