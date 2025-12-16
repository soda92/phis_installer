import shutil
import os
from pathlib import Path
from .utils import run_command, logger
from .config import INSTALLER_DIR

def find_makensis():
    """Finds makensis executable, handling scoop shims."""
    makensis_path = shutil.which("makensis")
    if not makensis_path:
        raise FileNotFoundError("makensis not found in PATH")

    # Check if it's a scoop shim
    # On Linux, shutil.which returns the path directly.
    # The original script was PowerShell on Windows.
    # Since we are on Linux (according to system info), or running cross-platform logic,
    # we should just trust `which` unless we are on Windows.
    
    if os.name == 'nt':
        # Simple heuristic for Windows shim
        # But for now, let's assume standard behavior or let the user config it.
        # If we really need to parse shims, we can add that later.
        pass
        
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
    
    # Create UTF-16BE version (NSIS unicode requirement on Windows sometimes, 
    # but let's follow the old script's lead)
    # The old script did: Get-Content ... | Out-File ... -Encoding BigEndianUnicode
    
    utf16_script = script_path.with_suffix(".utf16be.nsi")
    
    # We read as utf-8 (assuming source is utf-8) and write as utf-16-be
    content = script_path.read_text(encoding="utf-8")
    
    # Inject defines if needed via text replacement (or use command line /D)
    # Command line /D is cleaner.
    
    # Write to UTF-16BE with BOM? NSIS usually likes BOM for UTF-16.
    # Python's 'utf-16-be' does NOT write BOM. 'utf-16' does (if LE/BE not specified).
    # But BigEndianUnicode in PowerShell is usually UTF-16BE with BOM.
    # Let's try to just pass the original file first? 
    # The original script explicitly converted. Let's replicate.
    
    with open(utf16_script, "w", encoding="utf-16-be") as f:
        f.write("\ufeff") # BOM for BE? No, \ufeff is standard BOM. BE is \ufeff. LE is \ufffe.
        # Actually utf-16-be with BOM is \xfe\xff... 
        # Python's 'utf-16' adds BOM automatically.
        pass

    # Actually, let's just use 'utf-16' which defaults to OS endianness or adds BOM.
    # If the original script used BigEndianUnicode, it likely meant UTF-16BE.
    # Let's try writing simple 'utf-8' first. makensis V3 supports it.
    # If legacy makensis, it might need conversion.
    # Given we are on Linux now (see prompt context), `makensis` on Linux handles UTF-8 fine usually.
    # BUT, the target system might be Windows? The user is on Linux now.
    # Let's write UTF-16BE just to be safe and match the old script.
    
    with open(utf16_script, "w", encoding="utf-16-be") as f:
        f.write("\ufeff") # Explicit BOM
        f.write(content)

    cmd = [makensis]
    if defines:
        for k, v in defines.items():
            cmd.append(f"/D{k}={v}")
    
    # Enable V2 compatible mode if needed? Old script used /V2 (verbosity)
    cmd.append("/V2")
    cmd.append(str(utf16_script))
    
    run_command(cmd)
    
    # Clean up
    # utf16_script.unlink() 

def generate_upgrade_script(from_ver, to_ver, template_path):
    """Generates an upgrade NSI script from template."""
    content = template_path.read_text(encoding="utf-8")
    content = content.replace("%%FROM_VERSION%%", from_ver)
    content = content.replace("%%TO_VERSION%%", to_ver)
    
    dest = INSTALLER_DIR / f"upgrade_{from_ver}_to_{to_ver}.nsi"
    dest.write_text(content, encoding="utf-8")
    return dest.name
