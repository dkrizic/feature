# Feature

Simple feature flag service designed for Kubernetes.

Related components:

- [UI README](./ui/README.md)
- [CLI README](./cli/README.md)
- [Service README](./service/README.md)

```mermaid
graph TD
    UI[User Interface]
    CLI[Command Line Interface]
    APP1[Application 1]
    APP2[Application 2]
    subgraph S[Service]
        API[gRPC API]
        P[Persistence]
        INM[In-Memory Storage]
        CM[ConfigMap]
        API-->P
        P-->INM
        P-->CM
    end
    UI-->API
    CLI-->API
    APP1-->API
    APP2-->|Mount|CM
```
