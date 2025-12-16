import shutil
from .utils import run_command
from .config import INSTALLER_DIR
from pathlib import Path
from .utils import logger
import re


def find_makensis():
    """Finds makensis executable."""
    makensis_path = shutil.which("makensis")
    if not makensis_path:
        # On some systems (like the user's potentially if via scoop/windows), it might be elsewhere.
        # But we rely on path.
        raise FileNotFoundError("makensis not found in PATH")
    return makensis_path


def find_nsis_root(makensis_path):
    """
    Deduces NSIS root directory from makensis executable path.
    Handles standard installs and some scoop variations.
    """
    path = Path(makensis_path).resolve()

    # Check for Scoop shim
    # Scoop shims are usually in .../scoop/shims/makensis.exe
    # The real path is in .../scoop/shims/makensis.shim (text file)
    if "scoop" in str(path).lower() and "shims" in str(path).lower():
        shim_file = path.with_suffix(".shim")
        if shim_file.exists():
            try:
                content = shim_file.read_text(encoding="utf-8", errors="ignore")
                # Format is usually 'path = "C:\..."'
                match = re.search(r'path\s*=\s*"(.*)"', content)
                if match:
                    real_path_str = match.group(1)
                    path = Path(real_path_str).resolve()
                    logger.info(f"Resolved Scoop shim to: {path}")
            except Exception as e:
                logger.warning(f"Failed to parse shim file {shim_file}: {e}")

    parent = path.parent
    if parent.name.lower() == "bin":
        return parent.parent
    return parent


def install_plugin():
    """Copies nsisunz.dll to NSIS plugins directory."""
    try:
        makensis = find_makensis()
    except FileNotFoundError:
        logger.warning("Could not find makensis, skipping plugin install.")
        return

    nsis_root = find_nsis_root(makensis)

    # Target directory: Plugins/x86-unicode (standard for NSIS 3 Unicode)
    # The PS1 script used 'Plugins/x86-unicode'.
    # Note: On Linux, NSIS plugins might be in /usr/share/nsis/Plugins/...

    plugin_src = INSTALLER_DIR / "nsisunz.dll"
    if not plugin_src.exists():
        logger.warning(f"Plugin {plugin_src} not found. Skipping.")
        return

    # Try to find the plugins dir
    possible_dirs = [
        nsis_root / "Plugins" / "x86-unicode",
        nsis_root / "Plugins",
        # common linux paths if nsis_root is /usr/bin/.. -> /usr/share/nsis
        Path("/usr/share/nsis/Plugins/x86-unicode"),
        Path("/usr/share/nsis/Plugins"),
    ]

    dest_dir = None
    for d in possible_dirs:
        if d.exists() and d.is_dir():
            dest_dir = d
            break

    if dest_dir:
        dest_path = dest_dir / "nsisunz.dll"
        logger.info(f"Copying plugin to {dest_path}")
        try:
            shutil.copy2(plugin_src, dest_path)
        except PermissionError:
            logger.warning(
                f"Permission denied copying to {dest_path}. Run as admin/sudo if needed."
            )
        except Exception as e:
            logger.warning(f"Failed to copy plugin: {e}")
    else:
        logger.warning("Could not determine NSIS Plugins directory.")


def compile_nsis(script_name, defines=None):
    """
    Compiles an NSIS script.
    defines: dict of key-value pairs to pass as /DKey=Value
    """
    # Ensure plugin is present
    install_plugin()

    script_path = INSTALLER_DIR / script_name
    if not script_path.exists():
        raise FileNotFoundError(f"NSIS script not found: {script_path}")

    makensis = find_makensis()

    # We will just pass the script to makensis.
    # If encoding issues arise, we can handle them, but makensis v3+ handles UTF-8 BOM or UTF-16LE/BE.
    # We'll assume the script is in a compatible encoding (UTF-8 or UTF-8 BOM).
    # The previous script converted to UTF-16BE. Let's try to match that if we want maximum safety,
    # or just try compiling the source if it is UTF-8 (NSIS 3 supports UTF-8).

    # Let's generate a temporary UTF-16BE file just like the old script did, to be safe.
    utf16_script = script_path.with_suffix(".utf16be.nsi")
    content = script_path.read_text(encoding="utf-8")

    with open(utf16_script, "w", encoding="utf-16-be") as f:
        f.write("\ufeff")  # BOM
        f.write(content)

    cmd = [makensis]
    if defines:
        for k, v in defines.items():
            cmd.append(f"/D{k}={v}")

    cmd.append("/V2")
    cmd.append(str(utf16_script))

    run_command(cmd)

    # Check if we should delete the temp file.
    # Maybe keep it for debugging or delete it.
    # utf16_script.unlink()


def generate_upgrade_script(from_ver, to_ver, template_path):
    """Generates an upgrade NSI script from template."""
    content = template_path.read_text(encoding="utf-8")
    content = content.replace("%%FROM_VERSION%%", from_ver)
    content = content.replace("%%TO_VERSION%%", to_ver)

    dest = INSTALLER_DIR / f"upgrade_{from_ver}_to_{to_ver}.nsi"
    dest.write_text(content, encoding="utf-8")
    return dest.name
