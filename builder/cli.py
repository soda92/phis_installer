import argparse
from .config import load_config, save_config, INSTALLER_DIR
from .deps import (
    add_dep,
    download_full_deps,
    download_deps,
    get_packages_for_range,
    download_pip_tools,
)
from .nsis import compile_nsis, generate_upgrade_script
from .utils import logger
from .registry import clean_registry
from .zipapp import make_zipapp, run_zipapp


def main():
    parser = argparse.ArgumentParser(description="Installer Build Tools")
    subparsers = parser.add_subparsers(dest="command", required=True)

    # config
    cfg = load_config()

    # cmd: add-dep
    p_add = subparsers.add_parser("add-dep", help="Add a dependency")
    p_add.add_argument("package", help="Package name (e.g. 'pandas>=1.0')")
    p_add.add_argument(
        "--tag", default=cfg.get("version"), help="Version tag to add under"
    )

    # cmd: download-deps
    p_dl = subparsers.add_parser("download-deps", help="Download all dependencies")
    p_dl.add_argument(
        "--diff",
        nargs=2,
        metavar=("FROM", "TO"),
        help="Download only diff between versions",
    )

    # cmd: build-installer
    p_build = subparsers.add_parser("build-installer", help="Build full installer")
    p_build.add_argument(
        "--no-download", action="store_true", help="Skip downloading deps"
    )

    # cmd: build-upgrade
    p_up = subparsers.add_parser("build-upgrade", help="Build upgrade package")
    p_up.add_argument("--from-ver", required=True, help="Upgrade from version")
    p_up.add_argument("--to-ver", default=cfg.get("version"), help="Upgrade to version")

    # cmd: set-version
    p_ver = subparsers.add_parser("set-version", help="Update project version")
    p_ver.add_argument("version", help="New version string")

    # cmd: clean-registry
    subparsers.add_parser("clean-registry", help="Clean registry keys (Windows only)")

    # cmd: make-zipapp
    subparsers.add_parser("make-zipapp", help="Create test zipapp")

    # cmd: run-zipapp
    subparsers.add_parser("run-zipapp", help="Run test zipapp")

    args = parser.parse_args()

    if args.command == "add-dep":
        add_dep(args.package, args.tag)

    elif args.command == "download-deps":
        download_pip_tools()
        if args.diff:
            from_v, to_v = args.diff
            logger.info(f"Downloading diff {from_v} -> {to_v}")
            pkgs = get_packages_for_range(from_v, to_v)
            if pkgs:
                logger.info(f"Packages: {pkgs}")
                download_deps(
                    INSTALLER_DIR / f"packages_upgrade_{from_v}_to_{to_v}", pkgs
                )
            else:
                logger.info("No new packages to download.")
        else:
            logger.info("Downloading all dependencies...")
            download_full_deps()

    elif args.command == "build-installer":
        if not args.no_download:
            download_pip_tools()
            download_full_deps()

        compile_nsis(cfg["nsis_script"])

    elif args.command == "build-upgrade":
        from_v = args.from_ver
        to_v = args.to_ver

        # 1. Download Diff
        pkgs = get_packages_for_range(from_v, to_v)
        dl_dir = INSTALLER_DIR / f"packages_upgrade_{from_v}_to_{to_v}"

        req_file = INSTALLER_DIR / f"requirements_upgrade_{from_v}_to_{to_v}.txt"
        with open(req_file, "w", encoding="utf-8") as f:
            for p in pkgs:
                f.write(p + "\n")

        if pkgs:
            logger.info(f"Downloading {len(pkgs)} packages for upgrade...")
            download_deps(dl_dir, pkgs)
        else:
            logger.info("No new packages. Creating empty upgrade.")
            dl_dir.mkdir(parents=True, exist_ok=True)

        # 2. Generate NSI
        tpl_path = INSTALLER_DIR / "upgrade_template.nsi"
        nsi_name = generate_upgrade_script(from_v, to_v, tpl_path)

        # 3. Compile
        compile_nsis(nsi_name)

    elif args.command == "set-version":
        cfg["version"] = args.version
        save_config(cfg)
        logger.info(f"Version updated to {args.version}")

    elif args.command == "clean-registry":
        clean_registry()

    elif args.command == "make-zipapp":
        make_zipapp()

    elif args.command == "run-zipapp":
        run_zipapp()

    return 0
