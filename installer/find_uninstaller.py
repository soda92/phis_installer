import winreg
import logging


def find_python_uninstall_entry():
    uninstall_key_path = r"SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall"
    try:
        # Open the Uninstall key
        with winreg.OpenKey(
            winreg.HKEY_LOCAL_MACHINE, uninstall_key_path
        ) as uninstall_key:
            i = 0
            while True:
                try:
                    # Enumerate subkeys (each subkey represents an installed program)
                    subkey_name = winreg.EnumKey(uninstall_key, i)
                    current_program_key_path = f"{uninstall_key_path}\\{subkey_name}"

                    with winreg.OpenKey(
                        winreg.HKEY_LOCAL_MACHINE, current_program_key_path
                    ) as program_key:
                        try:
                            # Get the DisplayName value
                            display_name, _ = winreg.QueryValueEx(
                                program_key, "DisplayName"
                            )
                            if "Python" in display_name:
                                logging.info("Found Python Uninstall Entry:")
                                logging.info(f"  Display Name: {display_name}")
                                # Attempt to get other relevant values if they exist
                                try:
                                    display_version, _ = winreg.QueryValueEx(
                                        program_key, "DisplayVersion"
                                    )
                                    logging.info(f"  Version: {display_version}")
                                except FileNotFoundError:
                                    pass
                                try:
                                    uninstall_string, _ = winreg.QueryValueEx(
                                        program_key, "UninstallString"
                                    )
                                    logging.info(
                                        f"  Uninstall String: {uninstall_string}"
                                    )
                                except FileNotFoundError:
                                    pass
                                logging.info(
                                    f"  Registry Key: HKLM\\{current_program_key_path}"
                                )
                                logging.info(f"{...}")
                        except FileNotFoundError:
                            # DisplayName not found for this subkey, skip
                            pass
                    i += 1
                except OSError:
                    # No more subkeys to enumerate
                    break
    except FileNotFoundError:
        logging.info(f"Registry key not found: HKLM\\{uninstall_key_path}")
    except Exception as e:
        logging.exception(f"An error occurred: {e}")


if __name__ == "__main__":
    find_python_uninstall_entry()
