# Phis Installer Builder

[English](README.md) | [中文](README_zh.md)

NSIS 安装包构建工具，用于构建 Windows 安装程序并管理 Python 依赖更新。

## 功能特性
- **跨平台支持:** 在 Linux 上构建 Windows 安装程序（使用 `makensis` 和 `pip` 交叉编译）。
- **依赖管理:** 自动解析 `pyproject.toml` (使用 `uv`) 或 `requirements.txt` 中的依赖。
- **增量升级:** 通过计算版本快照之间的差异，生成仅包含变更组件的小型升级包。
- **清理构建:** 支持在构建前清理包目录，确保依赖下载是最新的。

## 前置要求
- **Go 1.18+**
- **NSIS 3.0+** (Linux 上可通过 `sudo apt install nsis` 安装)
- **Python 3** (用于 `pip download`)
- **uv** (必需，用于跨平台解析 `pyproject.toml`)

## 使用指南

### 1. 构建工具
```bash
cd builder
go build -o builder main.go
# 推荐：将二进制文件移动到根目录或添加到 PATH
mv builder ../phis-builder
cd ..
```

或者使用便捷脚本（如果你有 `fish` shell）：
```bash
./builder.fish
```

### 2. 下载静态资源
下载 `python-embed` 和 `VC_redist` 等静态资源。
```bash
./phis-builder download-resources
```

### 3. 构建完整安装包
构建包含所有依赖的完整安装程序。
```bash
./phis-builder build-installer --clean
```
默认情况下 `--clean` 为 true，即在下载前删除已有的 `build/packages/` 和 `build/pip_wheels/` 目录。

### 4. 版本快照
解析当前的 `pyproject.toml` (或 `requirements.txt`) 并保存为版本快照（例如 `versions/requirements_20.txt`）。
```bash
./phis-builder snapshot-version --version 20
```

### 5. 构建升级包
构建一个从旧版本（如 1.9）升级到当前版本（20）的升级包。
```bash
./phis-builder build-upgrade --from-ver 1.9 --to-ver 20
```
该命令将：
1.  比较 `versions/requirements_1.9.txt` 与 `versions/requirements_20.txt`（若快照缺失则直接解析当前的 `pyproject.toml`）。
2.  下载新增或更新的 whl 包到 `build/packages_upgrade_...`。
3.  生成 NSIS 升级脚本。
4.  编译升级安装包（例如 `自动化平台_升级包_1.9_至_20.exe`）。

## 配置
配置文件位于 `resources/config.toml`。
