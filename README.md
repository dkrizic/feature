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
    CLI[Command Line Interface]
    APP1[Application 1]
    APP2[Application 2]
    subgraph FR[Frontend]
      BR[Browser]
      FS[Frontend Service]
    end
    subgraph S[Service]
        API[gRPC API]
        P[Persistence]
        INM[In-Memory Storage]
        CM[ConfigMap]
        API-->P
        P-->INM
        P-->|REST|CM
    end
    BR-->|REST|FS
    FS-->|gRPC|API
    CLI-->|gRPC|API
    APP1-->|gRPC|API
    APP2-->|Mount|CM
```
## Features

* Available as OCI containers
* Multi architecture (amd64, arm64)
* gRPC API for managing feature flags
* REST API for frontend consumption
* Persistence layer with in-memory and Kubernetes ConfigMap backends
* Command Line Interface (CLI) for managing feature flags
* Designed for Kubernetes environments
* OpenTelemetry instrumentation for observability
* Configurable via environment variables and ConfigMaps
* Lightweight and easy to deploy
