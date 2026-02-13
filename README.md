# Phis Installer Builder

[English](README.md) | [中文](README_zh.md)

Tool for building NSIS installers and managing Python dependency upgrades.

## Features
- **Cross-Platform:** Build Windows installers on Linux (using `makensis` and `pip` cross-compilation).
- **Dependency Management:** Automatically calculate diffs between versions and download only necessary packages.
- **Upgrade Packages:** Generate small upgrade installers containing only changed components.

## Prerequisites
- **Go 1.18+**
- **NSIS 3.0+** (`sudo apt install nsis` on Linux)
- **Python 3** (for `pip download`)
- **uv** (Recommended, for accurate cross-platform dependency resolution)

## Usage

### 1. Build the Tool
```bash
cd builder
go build -o builder main.go
# Recommended: move binary to root or add to PATH
mv builder ../phis-builder
cd ..
```

Or use the convenience script (if you have `fish` shell):
```bash
./builder.fish
```

### 2. Download Static Resources
Download static resources like `python-embed` and `VC_redist`.
```bash
./phis-builder download-resources
```

### 3. Build Full Installer
Build the full installer with all dependencies.
```bash
./phis-builder build-installer
```
The output will be in the `build/` directory.

### 4. Snapshot Version
Save current `requirements.txt` as a version snapshot (e.g., `versions/requirements_20.txt`).
```bash
./phis-builder snapshot-version --version 20
```

### 5. Build Upgrade Package
Build an installer that upgrades from an older version (e.g., 1.9) to the current version (20).
```bash
./phis-builder build-upgrade --from-ver 1.9 --to-ver 20
```
This will:
1.  Compare `versions/requirements_1.9.txt` vs `versions/requirements_20.txt`.
2.  Download missing/upgraded wheels to `build/packages_upgrade_...`.
3.  Generate an NSIS script.
4.  Compile the upgrade installer (e.g., `数字员工平台_升级包_1.9_至_20.exe`).

## Configuration
Configuration is loaded from `resources/config.toml`.
