# Print available commands
default:
  @just --list

# Build standard-backups binary in dist/
build-standard-backups:
  mkdir -p dist
  go build -o dist/standard-backups ./cmd/standard-backups

# Build all binaries
build: \
  build-standard-backups

# Run standard-backups
run *args:
  @go run ./cmd/standard-backups {{ args }}
