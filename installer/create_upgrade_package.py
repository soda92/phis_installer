import subprocess
import sys
import argparse
from pathlib import Path
import shutil


def parse_requirements(filepath: Path) -> set:
    """Parses a requirements.txt file into a set of package lines, ignoring comments and empty lines."""
    if not filepath.exists():
        print(f"错误: 找不到需求文件 {filepath}", file=sys.stderr)
        sys.exit(1)

    packages = set()
    with open(filepath, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith("#"):
                packages.add(line)
    return packages


def main():
    """Main function to create the differential upgrade package."""
    parser = argparse.ArgumentParser(
        description="Create a differential upgrade package for Python dependencies."
    )
    parser.add_argument("from_version", help="The version to upgrade from (e.g., 1.7)")
    parser.add_argument("to_version", help="The version to upgrade to (e.g., 1.8)")
    args = parser.parse_args()

    from_ver = args.from_version
    to_ver = args.to_version

    base_dir = Path(__file__).parent
    old_reqs_path = base_dir / f"requirements_{from_ver}.txt"
    new_reqs_path = base_dir / f"requirements_{to_ver}.txt"
    upgrade_reqs_path = base_dir / f"requirements_upgrade_{from_ver}_to_{to_ver}.txt"
    upgrade_packages_dir = base_dir / f"packages_upgrade_{from_ver}_to_{to_ver}"

    print(f"--- 正在为 {from_ver} -> {to_ver} 创建差异升级包 ---")

    if not new_reqs_path.exists():
        print(f"错误: 找不到目标版本需求文件 '{new_reqs_path.name}'。", file=sys.stderr)
        sys.exit(1)

    if not old_reqs_path.exists():
        print(f"错误: 找不到起始版本需求文件 '{old_reqs_path.name}'。", file=sys.stderr)
        sys.exit(1)

    # 1. Find the difference between the two requirement files
    print("正在比较新旧需求文件...")
    old_packages = parse_requirements(old_reqs_path)
    new_packages = parse_requirements(new_reqs_path)

    diff_packages = new_packages - old_packages

    if not diff_packages:
        print("未发现新包。无需操作。")
        upgrade_packages_dir.mkdir(exist_ok=True)
        upgrade_reqs_path.write_text("", encoding="utf-8")
        sys.exit(0)

    print(f"发现 {len(diff_packages)} 个新增/更新的包:")
    for pkg in sorted(list(diff_packages)):
        print(f"  - {pkg}")

    # 2. Write the differential requirements to a new file
    with open(upgrade_reqs_path, "w", encoding="utf-8") as f:
        for pkg in sorted(list(diff_packages)):
            f.write(f"{pkg}\n")
    print(f"差异化需求已写入 '{upgrade_reqs_path.name}'")

    # 3. Download only the new packages into the upgrade directory
    print(f"正在下载差异包到 '{upgrade_packages_dir.name}'...")
    if upgrade_packages_dir.exists():
        shutil.rmtree(upgrade_packages_dir)
    upgrade_packages_dir.mkdir(exist_ok=True)

    command = [
        sys.executable,
        "-m",
        "pip",
        "download",
        "-r",
        str(upgrade_reqs_path),
        "-d",
        str(upgrade_packages_dir),
        "-i",
        "https://mirrors.tuna.tsinghua.edu.cn/pypi/web/simple",
    ]

    subprocess.run(command, check=True, text=True, encoding="utf-8")

    print("--- 差异包创建完成 ---")


if __name__ == "__main__":
    main()
