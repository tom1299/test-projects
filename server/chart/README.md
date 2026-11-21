# Server Helm chart

Build and publish the server image to a registry accessible by the cluster, then
install the chart from the repository root:

```sh
docker build -t <registry>/server:<tag> ./server
docker push <registry>/server:<tag>
helm install server ./server/chart \
  --set image.repository=<registry>/server \
  --set image.tag=<tag>
```

Pushing a Git tag also builds and publishes the image as
`ghcr.io/<owner>/<repository>:<tag>` using GitHub Actions. Image names are
lowercased and tags are sanitized to meet Docker's naming rules. Ensure the
cluster can pull the package; GitHub Container Registry packages are private by
default.

By default, the chart creates one replica and a `ClusterIP` Service exposing UDP
port 6745 and TCP port 6746. Set `service.type` to `NodePort` or `LoadBalancer`
to expose the Service outside the cluster. Mixed TCP/UDP LoadBalancer support
depends on the cluster's load balancer implementation.

The Service ports can be changed with `service.udpPort` and `service.tcpPort`;
the application still listens on UDP 6745 and TCP 6746 inside the container.
Standard Kubernetes Ingress is not included because it routes HTTP(S), not the
server's raw TCP and UDP traffic.

The default `podSecurityContext` runs as UID and GID 65532 with the
`RuntimeDefault` seccomp profile. The default container `securityContext` drops
all capabilities, prevents privilege escalation, and makes the root filesystem
read-only. Override either context in a values file if your environment requires
different settings.
