# Simple static file server

A simple **static file service** written in Go.

`fileServer` serves files from a configured static directory over HTTP.  
It can run directly as a binary or inside Docker.

## Features

- Serve static files over HTTP
- Configurable host, port, static directory, and file permission
- YAML-based configuration
- Default config embedded in the binary
- Docker runtime image support

## Project Structure

    .
    ├── build
    │   ├── bin
    │   ├── config
    │   │   └── config.yaml
    │   ├── docker-build.sh
    │   └── Dockerfile
    ├── cmd
    ├── configs
    ├── internal
    ├── scripts
    └── go.mod

## Configuration

Default configuration:

```yaml
server:
  host: "0.0.0.0"
  port: "8080"
  static_dir: "./static"
  permission: "0644"
  mode: "create"

logging:
  level: "INFO"
```

### Config options

- `server.host`: bind address
- `server.port`: service port
- `server.static_dir`: directory to serve
- `server.permission`: permission used for created files
- `server.mode`: server mode
- `logging.level`: log level

## Run Locally

### Run with default embedded config

```bash
go run ./cmd
```

### Run with custom config file

```bash
go run ./cmd --config ./your_config/config.yaml
```

## Build

### Build Linux binaries

```bash
./scripts/build-linux.sh --amd64
./scripts/build-linux.sh --arm64
```

### Build macOS binaries

```bash
./scripts/build-macos.sh --amd64
./scripts/build-macos.sh --arm64
```

Built binaries are placed under `build/bin/`.

Examples:

- `build/bin/fileServer-linux-amd64`
- `build/bin/fileServer-linux-arm64`
- `build/bin/fileServer-darwin-amd64`
- `build/bin/fileServer-darwin-arm64`

## Docker

Docker uses:

- a binary from `build/bin/`
- a config file from `build/config/config.yaml`

### Build Docker image

For ARM64:

```bash
./build/docker-build.sh --arm64
```

For AMD64:

```bash
./build/docker-build.sh --amd64
```

### Build Docker image with custom tag

```bash
./build/docker-build.sh --arm64 --tag file-server:latest
```

### Run Docker container

```bash
docker run --rm -p 8080:8080 file-server:arm64
```

or

```bash
docker run --rm -p 8080:8080 file-server:amd64
```

## Notes

- Docker images must use a **Linux binary**
- macOS binaries are for local execution only
- The default configuration is embedded in the binary
- You can override configuration with `--config`
