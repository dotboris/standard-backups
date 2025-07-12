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
  go run ./cmd/standard-backups \
    --config examples/config/etc/standard-backups/config.yaml \
    --backend-dirs examples/config/etc/standard-backups/backends.d \
    --recipe-dirs examples/config/etc/standard-backups/recipes.d \
    {{ args }}

# Run unit tests
[group("test")]
test:
  go test -v $(go list ./... | grep -v 'standard-backups/e2e')

# Run end-to-end tests
[group("test")]
e2e: build-all
  go test -v $(go list ./... | grep 'standard-backups/e2e')

# Runs all tests
[group("test")]
test-all: test e2e

# Generate mocks
[group("test")]
generate-mocks:
  mockery
