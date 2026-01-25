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

##### `--editable`

- **Flag name:** `editable`
- **Type:** string
- **Env var:** `EDITABLE`
- **Default:** `""` (empty - all fields editable)
- **Category:** `service`
- **Description:** Comma-separated list of field names that can be edited. When set, enforces field-level access control:
  - Only listed fields can be updated
  - New fields cannot be created
  - No fields can be deleted (all fields protected)
  - Non-listed fields become read-only
  - Empty value means all operations are allowed

**Behavior:**

When `EDITABLE=""` (default):
- ✅ Create new fields
- ✅ Update any field
- ✅ Delete any field

When `EDITABLE="FIELD1,FIELD2"`:
- ❌ Creating new fields blocked
- ✅ Update FIELD1 and FIELD2 only
- ❌ Update other fields (read-only)
- ❌ Delete any field (all protected)

**Note:** The `--preset` flag always bypasses editable restrictions, allowing baseline configuration to be established.

Examples:

```bash
# Allow editing only MAINTENANCE_FLOW and DEBUG_MODE
feature service --editable "MAINTENANCE_FLOW,DEBUG_MODE"

# Environment variable
EDITABLE=MAINTENANCE_FLOW,DEBUG_MODE feature service

# Combined with preset (preset always works regardless of editable)
feature service \
  --preset COLOR=red \
  --preset THEME=dark \
  --preset MAINTENANCE_FLOW=disabled \
  --editable "MAINTENANCE_FLOW"
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

Use field-level access control to restrict editing:

```bash
# Set up baseline configuration with only MAINTENANCE_FLOW editable
feature service \
  --preset COLOR=red \
  --preset THEME=dark \
  --preset MAINTENANCE_FLOW=disabled \
  --preset DEBUG_MODE=false \
  --editable "MAINTENANCE_FLOW,DEBUG_MODE"

# In this configuration:
# - COLOR and THEME cannot be modified (read-only)
# - MAINTENANCE_FLOW and DEBUG_MODE can be updated
# - No new fields can be created
# - No fields can be deleted
```

Print version information only:

```bash
feature version
```

---

## Field-Level Access Control

The `--editable` flag enables fine-grained control over which feature flags can be modified at runtime.

### Use Cases

1. **Production Environments**: Lock down critical configuration values while allowing operational toggles to be changed
2. **Multi-Tenant Scenarios**: Provide different access levels for different teams
3. **Configuration Management**: Prevent accidental deletion or creation of flags
4. **Immutable Baselines**: Use `--preset` to establish configurations that cannot be changed later

### API Behavior

When editable restrictions are active:

- **GetAll** RPC: Returns an `editable: bool` field for each feature flag indicating whether it can be modified
- **Set** RPC: Returns `PermissionDenied` error for:
  - Non-editable fields: `"field 'FIELD_NAME' is not editable"`
  - New field creation: `"creating new fields is not allowed when editable restrictions are active"`
- **Delete** RPC: Returns `PermissionDenied` error: `"deleting fields is not allowed when editable restrictions are active"`
- **PreSet** RPC: Always succeeds regardless of editable configuration (for initial setup)

### Example Error Messages

```bash
# Trying to update a read-only field
$ grpcurl -d '{"key":"COLOR","value":"blue"}' localhost:8000 feature.v1.Feature/Set
ERROR:
  Code: PermissionDenied
  Message: field 'COLOR' is not editable

# Trying to create a new field
$ grpcurl -d '{"key":"NEW","value":"value"}' localhost:8000 feature.v1.Feature/Set
ERROR:
  Code: PermissionDenied
  Message: creating new fields is not allowed when editable restrictions are active

# Trying to delete a field
$ grpcurl -d '{"name":"MAINTENANCE_FLOW"}' localhost:8000 feature.v1.Feature/Delete
ERROR:
  Code: PermissionDenied
  Message: deleting fields is not allowed when editable restrictions are active
```