# Feature

Simple feature flag service designed for Kubernetes.

Related components:

- [UI README](./ui/README.md)
- [CLI README](./cli/README.md)
- [Service README](./service/README.md)
- [Helm Chart README](./charts/feature/README.md)

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

