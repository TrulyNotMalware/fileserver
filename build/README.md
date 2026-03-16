# Build

Docker build resources for `fileServer`.

## Structure

    build/
    ├── bin/              # Compiled Linux binaries (git ignored)
    ├── config/
    │   └── config.yaml   # Config file for Docker image
    ├── Dockerfile
    └── docker-build.sh

## Files

- `Dockerfile`  
  Runtime-only image. No Go build step included.
- `docker-build.sh`  
  Builds a Linux binary and then builds the Docker image.
- `config/config.yaml`  
  Config file copied into the Docker image at build time.
- `bin/`  
  Output directory for compiled binaries. Git ignored.

## Usage

Build and create Docker image:

```bash
./build/docker-build.sh --arm64
./build/docker-build.sh --amd64
```

With custom tag:

```bash
./build/docker-build.sh --arm64 --tag file-server:latest
```

Skip binary build if already compiled:

```bash
./build/docker-build.sh --arm64 --no-binary-build
```

## Notes

- `bin/` is git ignored. Run `build-linux.sh` or `docker-build.sh` to populate it.
- Only Linux binaries can be used in the Docker image.
```