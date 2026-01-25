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
| `service.configMap.editable` | Comma-separated list of editable field names (empty = all editable) | `""` |
| `service.preset` | Pre-set key-value pairs (comma-separated, format: key=value) | `"COLOR=red,THEME=dark,BOOKING=true"` |
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
| `ui.subpath` | Subpath prefix for UI routes (e.g., `/feature` or `/app/v1`) | `""` |
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

## Field-Level Access Control

The service supports restricting which feature flags can be modified at runtime. This is useful for production environments where you want to lock down critical configuration while allowing specific flags to be toggled.

### Configuration

```yaml
service:
  configMap:
    editable: "MAINTENANCE_FLOW,DEBUG_MODE,FEATURE_X"
```

### Behavior

**When `editable` is empty (default):**
- ✅ All operations allowed: create, update, and delete any field

**When `editable` contains field names:**
- ❌ **Create**: New fields cannot be created
- ✅ **Update**: Only listed fields can be updated
- ❌ **Update**: Non-listed fields are read-only
- ❌ **Delete**: All fields are protected from deletion

### Example Configuration

```yaml
service:
  storageType: inmemory
  preset: "COLOR=red,THEME=dark,MAINTENANCE_FLOW=disabled,DEBUG_MODE=false"
  configMap:
    editable: "MAINTENANCE_FLOW,DEBUG_MODE"
```

With this configuration:
- `MAINTENANCE_FLOW` and `DEBUG_MODE` can be updated via UI/CLI/API
- `COLOR` and `THEME` are read-only (cannot be changed)
- No new fields can be created
- No fields can be deleted

### UI Behavior

When editable restrictions are active, the UI will:
- Show a warning that creating new fields is disabled
- Display editable fields with Update buttons (green background)
- Display read-only fields with disabled inputs and 🔒 indicator (orange background)
- Replace all Delete buttons with "🔒 Protected" indicators

### CLI Behavior

```bash
# List all features with editable status
$ feature-cli getall
key=COLOR value=red editable=read-only
key=MAINTENANCE_FLOW value=disabled editable=editable
key=DEBUG_MODE value=false editable=editable

# Try to update read-only field - denied
$ feature-cli set COLOR blue
Error: field 'COLOR' is not editable

# Try to create new field - denied
$ feature-cli set NEW_FIELD value
Error: creating new fields is not allowed when editable restrictions are active

# Try to delete any field - denied
$ feature-cli delete MAINTENANCE_FLOW
Error: deleting fields is not allowed when editable restrictions are active

# Update editable field - success
$ feature-cli set MAINTENANCE_FLOW enabled
```

### PreSet Bypass

The `preset` configuration always bypasses editable restrictions, allowing you to establish baseline configurations that cannot be modified later:

```yaml
service:
  preset: "CRITICAL_CONFIG=production,APP_VERSION=1.2.3"
  configMap:
    editable: ""  # Even if this were set, preset would still work
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

### Using a Subpath

To deploy the UI at a subpath (e.g., `/feature` instead of `/`):

```yaml
ui:
  subpath: /feature
  ingress:
    enabled: true
    className: nginx
    hosts:
      - host: example.com
        paths:
          - path: /feature
            pathType: Prefix
```

With this configuration:
- The UI will be accessible at `http://example.com/feature/`
- Health checks will be at `http://example.com/feature/health`
- All routes will be prefixed with `/feature`

## RBAC

When using ConfigMap storage, the chart creates:
- A ServiceAccount for the service pods
- A Role with permissions to manage ConfigMaps in the namespace
- A RoleBinding connecting the ServiceAccount to the Role

These resources are created when both `service.rbac.create` and `serviceAccount.create` are `true`.

## Image Tags

The chart uses the `appVersion` field from `Chart.yaml` as the image tag for all components. This ensures consistent versioning across the service, UI, and CLI.

### Version Management

- **Local Development**: The default values in `Chart.yaml` are `version: "0.0.0"` and `appVersion: "UNDEFINED"`, which are used when working with the chart locally.
- **CI/CD Releases**: When the chart is built and released through the CI pipeline (triggered by git tags matching `*.*.*`), both `version` and `appVersion` are automatically updated to match the git tag before packaging. For example, if you tag a release as `1.2.3`, the packaged chart will have both `version: "1.2.3"` and `appVersion: "1.2.3"`.
- This automation ensures that released charts always have proper version numbers and use the correct image tags without manual updates to `Chart.yaml`.

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
