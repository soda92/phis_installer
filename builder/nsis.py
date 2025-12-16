import shutil
from .utils import run_command
from .config import INSTALLER_DIR


def find_makensis():
    """Finds makensis executable."""
    makensis_path = shutil.which("makensis")
    if not makensis_path:
        # On some systems (like the user's potentially if via scoop/windows), it might be elsewhere.
        # But we rely on path.
        raise FileNotFoundError("makensis not found in PATH")
    return makensis_path


def compile_nsis(script_name, defines=None):
    """
    Compiles an NSIS script.
    defines: dict of key-value pairs to pass as /DKey=Value
    """
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
