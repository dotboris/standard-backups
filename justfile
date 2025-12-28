# Print available commands
default:
  @just --list

# Build a command
[group("build")]
build $cmd:
  mkdir -p dist
  go build -o dist/{{ cmd }} ./cmd/{{ cmd }}

# Build all backends
[group("build")]
build-backends: \
  (build "standard-backups-rsync-backend") \
  (build "standard-backups-restic-backend")

# Build all binaries
[group("build")]
build-all: \
  build-backends \
  (build "standard-backups")

# Run standard-backups
run *args: build-backends
  go run ./cmd/standard-backups {{ args }}

# Run standard-backups using config in `example/config/`
run-example *args: build-backends
  #!/usr/bin/env bash
  XDG_DATA_DIRS="$PWD/examples/config/share" \
  go run ./cmd/standard-backups \
    --config examples/config/etc/standard-backups/config.yaml \
    {{ args }}

# Run unit tests
[group("test")]
test:
  gotestsum \
    --format standard-verbose \
    --packages $(go list ./... | grep -v 'standard-backups/e2e')

# Run end-to-end tests
[group("test")]
e2e filter="": build-all
  gotestsum \
    --format standard-verbose \
    --packages $(go list ./... | grep 'standard-backups/e2e') \
    {{ if filter == "" {""} else { "-- -run " + filter } }}

# Runs all tests
[group("test")]
test-all: test e2e

# Generate mocks
[group("test")]
generate-mocks:
  mockery

# Run static code checks
[group("checks")]
lint:
  golangci-lint run

# Apply automated fixes for static code checks
[group("checks")]
lint-fix:
  golangci-lint run --fix

# Format code
[group("checks")]
fmt:
  golangci-lint fmt

# Check if code is formatted
[group("checks")]
fmt-check:
  golangci-lint fmt -d
