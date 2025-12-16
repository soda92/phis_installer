import sys
from .utils import logger

def clean_registry():
    """
    Removes the registry keys for the product.
    Mimics:
    Remove-Item -Path "HKLM:\\Software\\数字员工平台" -Recurse -Force
    Remove-Item -Path "HKLM:\\Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\数字员工平台" -Recurse -Force
    """
    if sys.platform != "win32":
        logger.warning("Registry cleanup is only supported on Windows.")
        return

    try:
        import winreg
    except ImportError:
        logger.error("winreg module not available.")
        return

    keys_to_delete = [
        (winreg.HKEY_LOCAL_MACHINE, r"Software\\数字员工平台"),
        (winreg.HKEY_LOCAL_MACHINE, r"Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\数字员工平台"),
    ]

    for root, path in keys_to_delete:
        logger.info(f"Deleting registry key: {path}")
        try:
            _delete_key_recursive(root, path)
            logger.info(f"Successfully deleted {path}")
        except FileNotFoundError:
            logger.info(f"Key not found: {path}")
        except PermissionError:
            logger.error(f"Permission denied deleting {path}. Run as Administrator.")
        except Exception as e:
            logger.error(f"Failed to delete {path}: {e}")

def _delete_key_recursive(root, path):
    import winreg
    try:
        open_key = winreg.OpenKey(root, path, 0, winreg.KEY_ALL_ACCESS)
    except FileNotFoundError:
        return

    info = winreg.QueryInfoKey(open_key)
    # Delete subkeys first
    for _ in range(info[0]):
        # EnumKey returns the name of the subkey.
        # We delete the first subkey repeatedly until none are left.
        # But wait, index changes.
        # It's safer to delete by name, but EnumKey takes index.
        # We can just always query index 0?
        # No, if we delete index 0, the next one becomes 0.
        # But recursive delete needs to open subkey.
        
        # Simpler approach:
        # Recursively delete subkeys.
        while True:
            try:
                subkey_name = winreg.EnumKey(open_key, 0)
                _delete_key_recursive(root, f"{path}\\{subkey_name}")
            except OSError:
                # No more subkeys
                break
    
    winreg.CloseKey(open_key)
    winreg.DeleteKey(root, path)
