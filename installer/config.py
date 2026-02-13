import tomli
import tomli_w
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent
INSTALLER_DIR = BASE_DIR / "installer"
CONFIG_PATH = INSTALLER_DIR / "config.toml"


def load_config():
    if not CONFIG_PATH.exists():
        # Default config if not exists
        return {
            "version": "1.9",
            "product_name": "数字员工平台",
            "requirements_file": "requirements.txt",
            "nsis_script": "installer.nsi",
        }
    with open(CONFIG_PATH, "rb") as f:
        return tomli.load(f)


def save_config(config):
    with open(CONFIG_PATH, "wb") as f:
        tomli_w.dump(config, f)
