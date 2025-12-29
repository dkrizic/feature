# Feature CLI

The `feature` CLI is a command‑line client for interacting with the Feature service.  
It lets you manage feature flags (key/value pairs) exposed by the service.

## Build

From the `cli` directory:

```bash
cd cli
go build -o feature
```

This produces a `feature` binary.

## Usage

```bash
feature [global flags] <command> [command arguments]
```

## Global Flags

These flags are defined on the root command in `main.go` and apply to all sub‑commands.

### `--log-format`

- **Env var:** `LOG_FORMAT`
- **Default:** `text`
- **Allowed values:** `text`, `json`
- **Description:** Log output format.

Examples:

```bash
feature --log-format text getall
feature --log-format json getall
```

### `--log-level`

- **Env var:** `LOG_LEVEL`
- **Default:** `info`
- **Allowed values:** `debug`, `info`, `warn`, `error`
- **Description:** Minimum log level emitted by the CLI.

Examples:

```bash
feature --log-level debug getall
feature --log-level error get my-key
```

### `--endpoint`

- **Env var:** `ENDPOINT`
- **Default:** `localhost:8000`
- **Required:** yes
- **Description:** Address of the Feature service endpoint.

Examples:

```bash
feature --endpoint localhost:8000 getall
ENDPOINT=localhost:8000 feature get my-key
```

## Commands

### `version`

Prints the Feature service name and version (using `meta.Service` and `meta.Version`).

```bash
feature --endpoint localhost:8000 version
```

### `getall`

Streams and prints all features as `key: value` pairs, one per line.

```bash
feature --endpoint localhost:8000 getall
```

Output example:

```text
feature-a: enabled
feature-b: disabled
```

### `get`

Gets a single feature by key.

```bash
feature --endpoint localhost:8000 get <key>
```

- **Arguments:**
    - `key` (string) – feature key to retrieve.

On success, prints the feature name followed by a newline.

Example:

```bash
feature --endpoint localhost:8000 get my-feature
```

### `set`

Sets or updates a feature key/value pair.

```bash
feature --endpoint localhost:8000 set <key> <value>
```

- **Arguments:**
    - `key` (string) – feature key to set.
    - `value` (string) – value to associate with the key.

Example:

```bash
feature --endpoint localhost:8000 set my-feature enabled
```

### `delete`

Deletes a feature by key.

```bash
feature --endpoint localhost:8000 delete <key>
```

- **Arguments:**
    - `key` (string) – feature key to delete.

Example:

```bash
feature --endpoint localhost:8000 delete my-feature
```

### `preset`

Pre‑sets (initializes) a feature key/value pair.

```bash
feature --endpoint localhost:8000 preset <key> <value>
```

- **Arguments:**
    - `key` (string) – feature key to pre‑set.
    - `value` (string) – value to associate with the key.

Example:

```bash
feature --endpoint localhost:8000 preset my-feature enabled
```

## Examples

```bash
# Get all features from a local service
feature --endpoint localhost:8000 getall

# Get a single feature
feature --endpoint localhost:8000 get my-feature

# Set a feature flag
feature --endpoint localhost:8000 set my-feature enabled

# Pre-set a feature
feature --endpoint localhost:8000 preset my-feature enabled

# Delete a feature
feature --endpoint localhost:8000 delete my-feature

# Use structured JSON logging at debug level
feature --endpoint localhost:8000 \
        --log-format json \
        --log-level debug \
        getall
```