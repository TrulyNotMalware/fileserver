# fileServer

A simple **static file service** written in Go.  
JWT (RS256) authentication is required to access files.

## Features

- Static file serving over HTTP
- JWT authentication (RS256)
- RSA key-based token signing and verification
- Access token + Refresh token flow
- Admin / Guest roles
- YAML-based configuration with environment variable override
- Docker runtime image support
- Cross-build scripts for Linux and macOS

## Project Structure

    .
    ├── build
    │   ├── bin/              # Compiled Linux binaries (git ignored)
    │   ├── keys/             # RSA key pair (git ignored) 
    │   ├── config/
    │   │   └── config.yaml   # Config for Docker image
    │   ├── docker-build.sh
    │   └── Dockerfile
    ├── cmd
    ├── configs
    ├── internal
    │   ├── auth/             # JWT middleware, handler, manager
    │   └── db/               # SQLite user and token store
    ├── keys/                 # RSA key pair (git ignored)
    ├── scripts
    └── go.mod

## Quick Start

### 1. Generate RSA key pair

```bash
./scripts/gen-keys.sh
```

Keys are saved to `build/keys/private.pem` and `build/keys/public.pem`.

### 2. Set config

```yaml
auth:
  private_key_path: "./keys/private.pem"
  access_token_ttl: "15m"
  refresh_token_ttl: "168h"
  admin_password: "your-admin-password"
  guest_password: "your-guest-password"
```

### 3. Run

```bash
go run ./cmd --config ./your_config/config.yaml
```

## Configuration

```yaml
server:
  host: "0.0.0.0"
  port: "8080"
  static_dir: "./static"
  permission: "0644"
  mode: "create"

logging:
  level: "INFO"

auth:
  private_key_path: ""
  access_token_ttl: "15m"
  refresh_token_ttl: "168h"
  admin_password: ""
  guest_password: ""
```

All values can be overridden with environment variables prefixed with `FS_`:

- `FS_AUTH_PRIVATE_KEY_PATH`
- `FS_AUTH_ADMIN_PASSWORD`
- `FS_AUTH_GUEST_PASSWORD`

## Build

### Generate RSA keys

```bash
./scripts/gen-keys.sh
```

### Build Linux binaries

```bash
./scripts/build-linux.sh --amd64
./scripts/build-linux.sh --arm64
```

Built binaries are placed under `build/bin/`.

## Docker

### Build image

```bash
./build/docker-build.sh --arm64
./build/docker-build.sh --amd64
```

### Run container

RSA private key is mounted at runtime:

```bash
docker run --rm -p 8080:8080 \
  -v "$(pwd)/keys:/app/keys" \
  file-server:arm64
```

## Roles

| Role  | Permission           |
|-------|----------------------|
| admin | Full access          |
| guest | Read-only (download) |

## Notes

- RSA private key is never baked into the Docker image
- Passwords are hashed with bcrypt at startup
- Admin and Guest accounts are seeded automatically on startup
- Only Linux binaries can be used in Docker
```