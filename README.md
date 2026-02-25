# Phis Installer Builder

[English](README.md) | [中文](README_zh.md)

Tool for building NSIS installers and managing Python dependency upgrades.

## Features
- **Cross-Platform:** Build Windows installers on Linux (using `makensis` and `pip` cross-compilation).
- **Dependency Management:** Automatically resolve dependencies from `pyproject.toml` (using `uv`) or `requirements.txt`.
- **Upgrade Packages:** Generate small upgrade installers containing only changed components by calculating diffs between version snapshots.
- **Clean Builds:** Ensure fresh downloads by cleaning package directories before building.

## Prerequisites
- **Go 1.18+**
- **NSIS 3.0+** (`sudo apt install nsis` on Linux)
- **Python 3** (for `pip download`)
- **uv** (Required for resolving `pyproject.toml` cross-platform)

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
./phis-builder build-installer --clean
```
By default, `--clean` is true, which removes existing `build/packages/` and `build/pip_wheels/` before downloading.

### 4. Snapshot Version
Resolve current `pyproject.toml` (or `requirements.txt`) and save as a version snapshot (e.g., `versions/requirements_20.txt`).
```bash
./phis-builder snapshot-version --version 20
```

### 5. Build Upgrade Package
Build an installer that upgrades from an older version (e.g., 1.9) to the current version (20).
```bash
./phis-builder build-upgrade --from-ver 1.9 --to-ver 20
```
This will:
1.  Compare `versions/requirements_1.9.txt` vs `versions/requirements_20.txt` (or resolve current `pyproject.toml` if snapshot is missing).
2.  Download missing/upgraded wheels to `build/packages_upgrade_...`.
3.  Generate an NSIS script.
4.  Compile the upgrade installer (e.g., `自动化平台_升级包_1.9_至_20.exe`).

## Configuration
Configuration is loaded from `resources/config.toml`.
