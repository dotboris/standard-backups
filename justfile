# Print available commands
default:
  @just --list

# Build a command
[group("build")]
build $cmd:
  mkdir -p dist
  go build -o dist/{{ cmd }} ./cmd/{{ cmd }}

# Build all binaries
[group("build")]
build-all: \
  (build "standard-backups") \
  (build "standard-backups-rsync-backend")

# Run standard-backups
run *args:
  @go run ./cmd/standard-backups {{ args }}

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
