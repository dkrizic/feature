# Feature

Simple feature flag service designed for Kubernetes.

Related components:

- [UI README](./ui/README.md)
- [CLI README](./cli/README.md)
- [Service README](./service/README.md)
- [Helm Chart README](./charts/feature/README.md)
- [Demo README](./demo/README.md)

## Overview

```mermaid
graph TD
    subgraph W1[Workload 1]
      APP1[Application 1]
    end
    subgraph W2[Workload 2]
      CM[ConfigMap]
      APP2[Application 2]
    end
    subgraph FR[Frontend]
      BR[Browser]
      FS[Frontend Service]
      CLI[Command Line Interface]
    end
    subgraph S[Service]
        API[gRPC API]
        P[Persistence]
        INM[In-Memory Storage]
        CM[ConfigMap]
        API-->P
        P-->|Update|INM
        P-->|Update|CM
    end
    BR-->|REST|FS
    FS-->|Update via gRPC|API
    CLI-->|Update via gRPC|API
    APP1-->|Read via gRPC|API
    APP2-->|Mount|CM
    API-->|Restart|APP2
    
```
## Features

* Available as OCI containers
* Multi architecture (amd64, arm64)
* gRPC API for managing feature flags
* REST API for frontend consumption
* Persistence layer with in-memory and Kubernetes ConfigMap backends
* Command Line Interface (CLI) for managing feature flags
* **Multi-application support** for managing feature flags across multiple applications
* **Field-level access control** with editable field restrictions
* Workload restart functionality for Deployments, StatefulSets, and DaemonSets
* Designed for Kubernetes environments
* OpenTelemetry instrumentation for observability
* Configurable via environment variables and ConfigMaps
* Lightweight and easy to deploy

## Multi-Application Support

The feature service supports managing feature flags for multiple applications from a single service instance. Each application can have its own:

- Namespace
- Storage type (in-memory or ConfigMap)
- ConfigMap configuration
- Workload restart settings
- Editable field restrictions

### Configuration

**Helm Chart:**
```yaml
service:
  applications:
    - name: yasm-frontend
      namespace: frontend
      storageType: configmap
      configMap:
        name: yasm-frontend
        preset: BANNER=Hello
        editable: BANNER
      workload:
        enabled: true
        type: deployment
        name: yasm-frontend
    - name: yasm-backend
      namespace: backend
      storageType: configmap
      configMap:
        name: yasm-backend
        preset: AUTH_ENABLED=true,BACKGROUND=blue
        editable: BACKGROUND
      workload:
        enabled: true
        type: deployment
        name: yasm-backend
  defaultApplication: yasm-frontend
```

### CLI Usage

```bash
# List all configured applications
feature-cli applications

# Get all features for a specific application
feature-cli -a yasm-frontend getall

# Set a feature value for an application
feature-cli -a yasm-backend set BACKGROUND green

# Get a feature value (uses default application if -a not specified)
feature-cli get BANNER

# Delete a feature
feature-cli -a yasm-frontend delete DEBUG_MODE
```

The `-a` or `--application` flag can be used with all feature management commands. If not specified, the default application (first in the list or set via `defaultApplication`) is used.

### Environment Variables

For multi-application mode, set:

```bash
APPLICATIONS=app1,app2,app3
DEFAULT_APPLICATION=app1

# Configuration for app1
APP1_NAMESPACE=namespace1
APP1_STORAGE_TYPE=configmap
APP1_CONFIGMAP_NAME=app1-config
APP1_PRESET=KEY1=value1,KEY2=value2
APP1_EDITABLE=KEY1
APP1_RESTART_ENABLED=true
APP1_RESTART_TYPE=deployment
APP1_RESTART_NAME=app1-deployment
```

Replace hyphens with underscores in application names for environment variable prefixes (e.g., `yasm-frontend` becomes `YASM_FRONTEND_`).

### Legacy Single-Application Mode

The service still supports the legacy single-application mode for backward compatibility. If `applications` is not set in the Helm chart or `APPLICATIONS` environment variable is not set, the service operates in single-application mode using the legacy configuration.

## Field-Level Access Control

The service supports restricting which feature flags can be modified at runtime through the `EDITABLE` configuration.

### Configuration

**Helm Chart:**
```yaml
service:
  configMap:
    editable: "MAINTENANCE_FLOW,DEBUG_MODE"
```

**Environment Variable:**
```bash
EDITABLE=MAINTENANCE_FLOW,DEBUG_MODE
```

### Behavior

**When `EDITABLE=""` (default - no restrictions):**
- ✅ Create new fields
- ✅ Update any field
- ✅ Delete any field

**When `EDITABLE="FIELD1,FIELD2"` (restrictions active):**
- ❌ Creating new fields is **not allowed**
- ✅ Update FIELD1 and FIELD2 only
- ❌ Update other fields (read-only)
- ❌ Delete **any** field (all protected)

### Use Cases

This feature is useful for:
- **Production environments**: Lock down critical feature flags while allowing specific flags to be toggled
- **Multi-tenant scenarios**: Provide different access levels for different teams
- **Configuration management**: Prevent accidental deletion or creation of flags in controlled environments
- **PreSet initialization**: Use `PreSet` to establish immutable baseline configurations

## ConfigMap Usage and Workload Restarts

This service uses ConfigMaps for configuration management. Understanding how ConfigMaps are mounted is important for knowing when workload restarts are necessary.

### ConfigMap with `envFrom` (Restart Required)

When a ConfigMap is mounted using `envFrom` to populate environment variables:

```yaml
envFrom:
  - configMapRef:
      name: app-config
```

Environment variables are set once when the pod starts. If you update the ConfigMap (e.g., changing `LOG_LEVEL=debug` to `LOG_LEVEL=info`), the running pods continue using the old values because environment variables cannot be updated in a running process. **You must restart the deployment for pods to pick up the new environment variable values.**

This is the method used by the feature service for injecting configuration values.

### ConfigMap as Volume Mount (Auto-Update)

When a ConfigMap is mounted as a volume:

```yaml
volumeMounts:
  - name: config
    mountPath: /etc/config
volumes:
  - name: config
    configMap:
      name: app-config
```

Kubernetes automatically updates the files in the mounted volume when the ConfigMap changes (typically within 60-90 seconds due to kubelet sync period). Applications that read configuration files dynamically (e.g., nginx reloading config, apps watching file changes) can pick up updates without pod restart.

### Why the Workload Restart Feature Matters

The workload restart button in the UI is essential for the first scenario - when using `envFrom` (as this service does), updating feature flags in the ConfigMap requires restarting pods to apply changes. This feature provides a convenient UI button to trigger that restart instead of requiring kubectl access.

Use the **Workload Management** section in the UI to restart deployments, statefulsets, or daemonsets after updating their ConfigMap configuration.
