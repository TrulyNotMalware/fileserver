# Scripts

Utility scripts for building `fileServer`.

## Files

- `build.sh`  
  Entry point for build commands.
- `build-linux.sh`  
  Builds Linux binaries.
- `build-macos.sh`  
  Builds macOS binaries.

## Usage

Build Linux binaries:

```bash
./scripts/build-linux.sh --amd64
./scripts/build-linux.sh --arm64
```

Build macOS binaries:

```bash
./scripts/build-macos.sh --amd64
./scripts/build-macos.sh --arm64
```

Build output is written to `build/bin/`.