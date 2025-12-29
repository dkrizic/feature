# Feature Service

The **Feature service** is the backend for managing feature flags (key/value pairs).  
It exposes functionality that can be consumed by the `feature` CLI and other clients.

The service binary is also called `feature`, but its primary entrypoint here is the `service` subcommand (see below).

## Build

From the `service` directory:

```bash
cd service
go build -o feature
```

This produces a `feature` binary that can run the service and utility commands.

## Usage

```bash
feature [global flags] <command> [command flags]
```

The service defines **global flags** (for logging) and subcommands:

- `version` – print the service version
- `service` – start the Feature service

---

## Global Flags

These flags are defined at the root level and apply to all subcommands.

### `--log-format`

- **Env var:** `LOG_FORMAT`
- **Default:** `text`
- **Allowed values:** `text`, `json`
- **Category:** `logging`
- **Description:** Log output format for the service.

### `--log-level`

- **Env var:** `LOG_LEVEL`
- **Default:** `info`
- **Allowed values:** `debug`, `info`, `warn`, `error`
- **Category:** `logging`
- **Description:** Minimum log level emitted by the service.

Example:

```bash
feature --log-format json --log-level debug service ...
```

---

## Commands

### `version`

Prints the Feature service name and version (derived from `meta.Service` and `meta.Version`).

```bash
feature version
```

This does not start the server; it only logs the version information.

---

### `service`

Starts the Feature service process.

```bash
feature service [flags]
```

#### Flags

All of the following flags are defined on the `service` subcommand.

##### `--port`

- **Flag name:** `port`
- **Type:** integer
- **Env var:** `PORT`
- **Default:** `8080`
- **Category:** `service`
- **Description:** Port on which the Feature service will listen.

Example:

```bash
feature service --port 8000
PORT=8000 feature service
```

##### `--enable-opentelemetry`

- **Flag name:** `enable-opentelemetry`
- **Type:** boolean
- **Env var:** `ENABLE_OPENTELEMETRY`
- **Default:** `false`
- **Category:** `observability`
- **Description:** Enable OpenTelemetry tracing for the service.

Example:

```bash
feature service --enable-opentelemetry
ENABLE_OPENTELEMETRY=true feature service
```

##### `--otlp-endpoint`

- **Flag name:** `otlp-endpoint`
- **Type:** string
- **Env var:** `OTLP_ENDPOINT`
- **Default:** `localhost:4317`
- **Category:** `observability`
- **Description:** OTLP collector endpoint used when OpenTelemetry is enabled.

Example:

```bash
feature service \
  --enable-opentelemetry \
  --otlp-endpoint otel-collector:4317
```

##### `--storage-type`

- **Flag name:** `storage-type`
- **Type:** string
- **Env var:** `STORAGE_TYPE`
- **Default:** `inmemory` (from `constant.StorageTypeInMemory`)
- **Description:** Storage backend for feature data.

**Allowed values (validated in code):**

- `inmemory` – in‑memory storage (non‑persistent; data lost on restart).
- `configmap` – use a Kubernetes ConfigMap as storage.

When `storage-type` is `configmap`, **`--configmap-name` must be set**; otherwise the service will fail validation.

Example:

```bash
# In-memory storage (default)
feature service --storage-type inmemory

# ConfigMap-based storage
feature service \
  --storage-type configmap \
  --configmap-name my-feature-flags
```

##### `--configmap-name`

- **Flag name:** `configmap-name`
- **Type:** string
- **Env var:** `CONFIGMAP_NAME`
- **Description:** Name of the Kubernetes ConfigMap used when `storage-type` is `configmap`.

If `storage-type=configmap` and `configmap-name` is empty, the process will return an error:

> `configmap-name cannot be empty when storage-type is configmap`

Example:

```bash
feature service \
  --storage-type configmap \
  --configmap-name feature-flags
```

##### `--preset`

- **Flag name:** `preset`
- **Type:** string slice (`key=value` pairs)
- **Env var:** `PRESET`
- **Description:** Pre-set key/value pairs before starting the service.  
  Each value must be in the format `key=value`. Multiple values can be provided by repeating the flag.

Examples:

```bash
# CLI syntax
feature service \
  --preset featureA=enabled \
  --preset featureB=disabled

# Environment variable (depending on how urfave/cli parses string slices)
PRESET=featureA=enabled,featureB=disabled feature service
```

---

## Logging Behavior

Before any command runs, the `beforeAction` hook:

1. Reads `--log-format` and `--log-level`.
2. Validates and maps `log-level` to `slog` levels.
3. Creates either a JSON or text handler:
    - JSON: `slog.NewJSONHandler(os.Stdout, ...)`
    - Text: `slog.NewTextHandler(os.Stdout, ...)`
4. Sets a global default logger for the process with `slog.SetDefault`.

If an invalid log level is supplied, the command fails with:

> `invalid log level: <value>`

---

## Example Invocations

Start the service with default settings:

```bash
feature service
```

Start on a custom port with JSON logs and debug level:

```bash
feature \
  --log-format json \
  --log-level debug \
  service \
  --port 8000
```

Enable OpenTelemetry and point to a custom OTLP endpoint:

```bash
feature \
  --log-format json \
  --log-level info \
  service \
  --enable-opentelemetry \
  --otlp-endpoint otel-collector.observability:4317
```

Use ConfigMap storage with some pre-set features:

```bash
feature \
  --log-format text \
  --log-level info \
  service \
  --storage-type configmap \
  --configmap-name feature-flags \
  --preset featureA=enabled \
  --preset featureB=disabled
```

Print version information only:

```bash
feature version
```