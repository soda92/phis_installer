import argparse
import sys
from .config import load_config, save_config, INSTALLER_DIR
from .deps import add_dep, download_full_deps, download_deps, get_packages_for_range, download_pip_tools
from .nsis import compile_nsis, generate_upgrade_script
from .utils import logger

def main():
    parser = argparse.ArgumentParser(description="Installer Build Tools")
    subparsers = parser.add_subparsers(dest="command", required=True)

    # config
    cfg = load_config()

    # cmd: add-dep
    p_add = subparsers.add_parser("add-dep", help="Add a dependency")
    p_add.add_argument("package", help="Package name (e.g. 'pandas>=1.0')")
    p_add.add_argument("--tag", default=cfg.get("version"), help="Version tag to add under")

    # cmd: download-deps
    p_dl = subparsers.add_parser("download-deps", help="Download all dependencies")
    p_dl.add_argument("--diff", nargs=2, metavar=("FROM", "TO"), help="Download only diff between versions")

    # cmd: build-installer
    p_build = subparsers.add_parser("build-installer", help="Build full installer")
    p_build.add_argument("--no-download", action="store_true", help="Skip downloading deps")

    # cmd: build-upgrade
    p_up = subparsers.add_parser("build-upgrade", help="Build upgrade package")
    p_up.add_argument("--from-ver", required=True, help="Upgrade from version")
    p_up.add_argument("--to-ver", default=cfg.get("version"), help="Upgrade to version")

    # cmd: set-version
    p_ver = subparsers.add_parser("set-version", help="Update project version")
    p_ver.add_argument("version", help="New version string")

    args = parser.parse_args()

    if args.command == "add-dep":
        add_dep(args.package, args.tag)

    elif args.command == "download-deps":
        download_pip_tools() # Always ensure pip tools are there
        if args.diff:
            from_v, to_v = args.diff
            logger.info(f"Downloading diff {from_v} -> {to_v}")
            pkgs = get_packages_for_range(from_v, to_v)
            if pkgs:
                logger.info(f"Packages: {pkgs}")
                download_deps(INSTALLER_DIR / f"packages_upgrade_{from_v}_to_{to_v}", pkgs)
            else:
                logger.info("No new packages to download.")
        else:
            logger.info("Downloading all dependencies...")
            download_full_deps()

    elif args.command == "build-installer":
        if not args.no_download:
            download_pip_tools()
            download_full_deps()
        
        # We need to ensure the NSI has the correct version. 
        # Passing /DPRODUCT_VERSION=... overrides the !define in script? 
        # NSIS: Command line defines override script defines if !ifdef checks are used, or just globally.
        # But standard `!define` will warn about redefinition.
        # Best to just compile. The script reads version from hardcode.
        # Let's update the script version before compile if it differs?
        # Or just pass the define and expect the user to update the script to support it.
        # For now, let's just compile existing script.
        
        compile_nsis(cfg["nsis_script"])

    elif args.command == "build-upgrade":
        from_v = args.from_ver
        to_v = args.to_ver
        
        # 1. Download Diff
        pkgs = get_packages_for_range(from_v, to_v)
        dl_dir = INSTALLER_DIR / f"packages_upgrade_{from_v}_to_{to_v}"
        
        # Also create requirements file for the upgrade
        req_file = INSTALLER_DIR / f"requirements_upgrade_{from_v}_to_{to_v}.txt"
        with open(req_file, "w", encoding="utf-8") as f:
            for p in pkgs:
                f.write(p + "\n")
        
        if pkgs:
            logger.info(f"Downloading {len(pkgs)} packages for upgrade...")
            download_deps(dl_dir, pkgs)
        else:
            logger.info("No new packages. Creating empty upgrade.")
            dl_dir.mkdir(parents=True, exist_ok=True) # Ensure dir exists

        # 2. Generate NSI
        tpl_path = INSTALLER_DIR / "upgrade_template.nsi"
        nsi_name = generate_upgrade_script(from_v, to_v, tpl_path)
        
        # 3. Compile
        compile_nsis(nsi_name)

    elif args.command == "set-version":
        cfg["version"] = args.version
        save_config(cfg)
        logger.info(f"Version updated to {args.version}")

    return 0
