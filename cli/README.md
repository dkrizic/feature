# Feature CLI

A command-line interface for interacting with the Feature service. The Feature CLI allows you to manage feature flags as key/value pairs, providing a simple way to get, set, delete, and query feature flags from the command line.

## Build

To build the Feature CLI from source:

```bash
cd cli
go build -o feature
```

This will create a `feature` binary in the current directory.

## Usage

The general usage pattern for the Feature CLI is:

```bash
feature [global flags] <command> [command arguments]
```

## Global Flags

Global flags can be set via command-line options or environment variables:

- `--log-format` (env: `LOG_FORMAT`)
  - Log output format
  - Values: `text` | `json`
  - Default: `text`

- `--log-level` (env: `LOG_LEVEL`)
  - Log verbosity level
  - Values: `debug` | `info` | `warn` | `error`
  - Default: `info`

- `--endpoint` (env: `ENDPOINT`)
  - Feature service endpoint address
  - Default: `localhost:8000`

## Commands

### version

Prints the service name and version information.

```bash
feature version
```

### getall

Streams and prints all features from the service. Each feature is printed on a separate line in the format `key: value`.

```bash
feature getall
```

### get

Fetches a single feature by its key. When successful, prints the feature key.

```bash
feature get <key>
```

**Arguments:**
- `key` - The feature key to retrieve

### set

Sets or updates a feature with the specified key and value.

```bash
feature set <key> <value>
```

**Arguments:**
- `key` - The feature key to set
- `value` - The value to assign to the feature

### delete

Deletes a feature by its key.

```bash
feature delete <key>
```

**Arguments:**
- `key` - The feature key to delete

### preset

Pre-sets or initializes a feature with the specified key and value. This is used for initial feature flag setup.

```bash
feature preset <key> <value>
```

**Arguments:**
- `key` - The feature key to preset
- `value` - The value to assign to the feature

## Examples

### Using the default endpoint

```bash
# Print version information
feature version

# Get all features
feature getall

# Get a specific feature
feature get my-feature-flag

# Set a feature flag
feature set my-feature-flag enabled

# Delete a feature flag
feature delete my-feature-flag

# Pre-set a feature flag
feature preset default-theme dark
```

### Using a custom endpoint

```bash
# Connect to a remote Feature service
feature --endpoint feature-service.example.com:8000 getall

# Set endpoint via environment variable
export ENDPOINT=feature-service.example.com:8000
feature getall
feature get my-feature-flag
```

### Using JSON logging

```bash
# Enable JSON formatted logs with debug level
feature --log-format json --log-level debug get my-feature-flag

# Or use environment variables
export LOG_FORMAT=json
export LOG_LEVEL=debug
feature get my-feature-flag
```

### Combined examples

```bash
# Set a feature on a remote service with debug logging
feature --endpoint prod-features:8000 --log-level debug set new-ui-enabled true

# Get all features from staging environment
ENDPOINT=staging-features:8000 feature getall
```
