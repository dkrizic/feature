# Feature Helm Chart

This Helm chart deploys the Feature flag service along with its UI and CLI components to a Kubernetes cluster.

## Components

The chart includes three main components:

- **Service**: The main gRPC feature flag service (port 8000)
- **UI**: HTMX-based web interface (port 80)
- **CLI**: Command-line interface container

## Installation

### Add the Helm Repository

```bash
helm repo add feature https://dkrizic.github.io/feature
helm repo update
```

### Install the Chart

```bash
helm install my-feature feature/feature
```

### Install with Custom Values

```bash
helm install my-feature feature/feature -f custom-values.yaml
```

## Configuration

The following table lists the main configurable parameters of the Feature chart and their default values.

### Service Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.enabled` | Enable the feature service deployment | `true` |
| `service.replicaCount` | Number of service replicas (must be 1 for inmemory storage) | `1` |
| `service.image.repository` | Service image repository | `ghcr.io/dkrizic/feature/feature` |
| `service.port` | Service gRPC port (container port) | `8000` |
| `service.service.port` | Kubernetes Service port (the port the Service listens on) | `80` |
| `service.storageType` | Storage backend type (`inmemory` or `configmap`) | `inmemory` |
| `service.configMap.name` | ConfigMap name (only for configmap storage) | `""` |
| `service.rbac.create` | Create RBAC resources for ConfigMap access | `true` |
| `service.resources` | CPU/Memory resource requests/limits | `{}` |
| `service.livenessProbe` | Liveness probe configuration | `grpc on http port` |
| `service.readinessProbe` | Readiness probe configuration | `grpc on http port` |

### UI Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ui.enabled` | Enable the UI deployment | `true` |
| `ui.replicaCount` | Number of UI replicas | `1` |
| `ui.image.repository` | UI image repository | `ghcr.io/dkrizic/feature/feature-ui` |
| `ui.endpoint` | Feature service endpoint (defaults to service name) | `""` |
| `ui.ingress.enabled` | Enable Ingress for UI | `false` |
| `ui.ingress.className` | Ingress class name | `""` |
| `ui.ingress.annotations` | Ingress annotations | `{}` |
| `ui.ingress.hosts` | Ingress hosts configuration | See values.yaml |
| `ui.httpRoute.enabled` | Enable Gateway API HTTPRoute | `false` |
| `ui.httpRoute.parentRefs` | Gateway references | See values.yaml |
| `ui.httpRoute.hostnames` | HTTPRoute hostnames | See values.yaml |
| `ui.resources` | CPU/Memory resource requests/limits | `{}` |

### CLI Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `cli.enabled` | Enable the CLI deployment | `true` |
| `cli.replicaCount` | Number of CLI replicas | `1` |
| `cli.image.repository` | CLI image repository | `ghcr.io/dkrizic/feature/feature-cli` |
| `cli.endpoint` | Feature service endpoint (defaults to service name) | `""` |
| `cli.resources` | CPU/Memory resource requests/limits | `{}` |

### Common Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nameOverride` | Override chart name | `""` |
| `fullnameOverride` | Override full name | `""` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.annotations` | Service account annotations | `{}` |
| `serviceAccount.automount` | Automount service account token | `true` |
| `podAnnotations` | Pod annotations | `{}` |
| `podLabels` | Pod labels | `{}` |
| `podSecurityContext` | Pod security context | `{}` |
| `securityContext` | Container security context | `{}` |
| `imagePullSecrets` | Image pull secrets | `[]` |

## Storage Types

The service supports two storage backends:

### In-Memory Storage (Default)

```yaml
service:
  storageType: inmemory
  replicaCount: 1  # Must be 1 for in-memory storage
```

### ConfigMap Storage

```yaml
service:
  storageType: configmap
  replicaCount: 3  # Can scale horizontally
  configMap:
    name: feature-flags
  rbac:
    create: true  # Required for ConfigMap access
```

## Endpoint Configuration

The UI and CLI components need to communicate with the feature service. By default, both components automatically use the feature service name (the Kubernetes Service resource name) as their endpoint. This allows the chart to work out of the box without additional configuration.

### Default Behavior

When `ui.endpoint` and `cli.endpoint` are not specified (or set to empty string), the chart automatically configures them to use the feature service name:

```yaml
# No explicit endpoint configuration needed
ui:
  enabled: true

cli:
  enabled: true
```

The environment variable `ENDPOINT` is automatically set in both UI and CLI containers via ConfigMaps loaded with `envFrom`.

### Custom Endpoint

You can override the default endpoint if you need to point the UI or CLI to a different service:

```yaml
ui:
  endpoint: "my-custom-service:8000"

cli:
  endpoint: "external-feature-service.example.com:8000"
```

This is useful when:
- Using an external feature service
- Connecting to a service in a different namespace
- Using a custom DNS name or load balancer

## Exposing the UI

### Using Ingress

```yaml
ui:
  ingress:
    enabled: true
    className: nginx
    hosts:
      - host: feature.example.com
        paths:
          - path: /
            pathType: Prefix
```

### Using Gateway API HTTPRoute

```yaml
ui:
  httpRoute:
    enabled: true
    parentRefs:
      - name: my-gateway
        sectionName: http
    hostnames:
      - feature.example.com
    rules:
      - matches:
          - path:
              type: PathPrefix
              value: /
```

## RBAC

When using ConfigMap storage, the chart creates:
- A ServiceAccount for the service pods
- A Role with permissions to manage ConfigMaps in the namespace
- A RoleBinding connecting the ServiceAccount to the Role

These resources are created when both `service.rbac.create` and `serviceAccount.create` are `true`.

## Image Tags

The chart uses the `appVersion` field from `Chart.yaml` as the image tag for all components. This ensures consistent versioning across the service, UI, and CLI.

## Uninstallation

```bash
helm uninstall my-feature
```

## Development

To test the chart locally:

```bash
# Lint the chart
helm lint charts/feature

# Render templates
helm template my-feature charts/feature

# Install from local files
helm install my-feature charts/feature
```
