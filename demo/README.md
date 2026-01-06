# Demo

Start demo with

```
docker compose up -d
```

Then open [http://localhost:80](http://localhost:80) in your browser.```

## Overview

```mermaid
graph TD
    style OC fill:lightgreen
    U[User]
    subgraph DC[Docker Compose Network]
    T[Traefik]
    F[Frontend]
    CLI[CLI]
    S[Service]
    subgraph OC[OpenTelemetry Collector]
      GRPC[gRPC Receiver]
      PIPE[Processing Pipeline]
      EXP[OTLP Exporter]
    end
    end
    GC[Grafana Cloud]
    T-->|HTTP|F
    F-->|gRPC|S
    CLI-->|gRPC|S
    S-->|Signals|GRPC
    GRPC-.->PIPE
    PIPE-.->EXP
    U-->T
    EXP-.->GC
    T-.->|Signals|GRPC
    F-.->|Signals|GRPC
    CLI-.->|Signals|GRPC
```
