import re
import sys
from packaging.version import parse
from .utils import run_command, logger
from .config import INSTALLER_DIR

REQUIREMENTS_FILE = INSTALLER_DIR / "requirements.txt"
PACKAGES_DIR = INSTALLER_DIR / "packages"
PIP_WHEELS_DIR = INSTALLER_DIR / "pip_wheels"


def parse_requirements_by_version(req_path):
    """
    Parses requirements.txt with version tags.
    Returns a dict: {'1.8': ['pandas', ...], '1.9': ['odfpy', ...]}`
    """
    version_map = {}
    current_version = "base"

    if not req_path.exists():
        return version_map

    with open(req_path, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue

            # Check for version tag like "# v 1.8" or "# v1.9"
            match = re.match(r"^#\s*v\s*([\d\.]+)", line, re.IGNORECASE)
            if match:
                current_version = match.group(1)
                continue

            if line.startswith("#"):
                continue

            if current_version not in version_map:
                version_map[current_version] = []
            version_map[current_version].append(line)

    return version_map


def get_packages_for_range(start_ver, end_ver):
    """
    Returns packages strictly AFTER start_ver and UP TO end_ver.
    If start_ver is None, returns ALL packages.
    """
    all_deps = parse_requirements_by_version(REQUIREMENTS_FILE)
    selected_deps = set()

    if start_ver is None:
        for ver, deps in all_deps.items():
            for dep in deps:
                selected_deps.add(dep)
        return list(selected_deps)

    try:
        start = parse(start_ver)
        end = parse(end_ver)
    except Exception as e:
        logger.error(f"Error parsing versions: {e}")
        return []

    for ver_str, deps in all_deps.items():
        if ver_str == "base":
            continue
        try:
            current = parse(ver_str)
            if current > start and current <= end:
                for dep in deps:
                    selected_deps.add(dep)
        except:
            logger.warning(f"Could not parse version tag: {ver_str}")

    return list(selected_deps)


def download_deps(target_dir, requirements_list):
    """Downloads deps from a list of strings using standard pip."""
    if not requirements_list:
        logger.info("No packages to download.")
        return

    # Write temp req file
    temp_req = target_dir / "temp_reqs.txt"
    target_dir.mkdir(parents=True, exist_ok=True)

    with open(temp_req, "w", encoding="utf-8") as f:
        for req in requirements_list:
            f.write(req + "\n")

    # Use standard pip download
    cmd = [
        sys.executable,
        "-m",
        "pip",
        "download",
        "-r",
        str(temp_req),
        "-d",
        str(target_dir),
        "-i",
        "https://mirrors.tuna.tsinghua.edu.cn/pypi/web/simple",
    ]
    run_command(cmd)
    temp_req.unlink()


def download_full_deps():
    """Downloads everything in requirements.txt using standard pip"""
    PACKAGES_DIR.mkdir(parents=True, exist_ok=True)
    cmd = [
        sys.executable,
        "-m",
        "pip",
        "download",
        "-r",
        str(REQUIREMENTS_FILE),
        "-d",
        str(PACKAGES_DIR),
        "-i",
        "https://mirrors.tuna.tsinghua.edu.cn/pypi/web/simple",
    ]
    run_command(cmd)


def download_pip_tools():
    """Downloads pip, setuptools, wheel using standard pip."""
    PIP_WHEELS_DIR.mkdir(parents=True, exist_ok=True)
    temp_pip_req = PIP_WHEELS_DIR / "pip_tools_reqs.txt"
    with open(temp_pip_req, "w", encoding="utf-8") as f:
        f.write("pip\n")
        f.write("setuptools\n")
        f.write("wheel\n")
    
    cmd = [
        sys.executable,
        "-m",
        "pip",
        "download",
        "-r",
        str(temp_pip_req),
        "-d",
        str(PIP_WHEELS_DIR),
        "-i",
        "https://mirrors.tuna.tsinghua.edu.cn/pypi/web/simple"
    ]
    run_command(cmd)
    temp_pip_req.unlink()


def add_dep(package_name, version_tag):
    """Adds a package to requirements.txt under the specified version tag."""
    lines = []
    if REQUIREMENTS_FILE.exists():
        with open(REQUIREMENTS_FILE, "r", encoding="utf-8") as f:
            lines = f.readlines()

    # Check if tag exists
    tag_header = f"# v {version_tag}\n"
    tag_alt = f"# v{version_tag}\n"

    found_index = -1
    for i, line in enumerate(lines):
        if line.lower() == tag_header.lower() or line.lower() == tag_alt.lower():
            found_index = i
            break

    if found_index != -1:
        # Insert after the tag
        lines.insert(found_index + 1, f"{package_name}\n")
    else:
        # Append new tag and package
        if lines and not lines[-1].endswith("\n"):
            lines.append("\n")
        lines.append(f"\n{tag_header}")
        lines.append(f"{package_name}\n")

    with open(REQUIREMENTS_FILE, "w", encoding="utf-8") as f:
        f.writelines(lines)
    logger.info(f"Added {package_name} to version {version_tag}")
