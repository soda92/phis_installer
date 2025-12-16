import sys
from .utils import run_command, logger
from .config import INSTALLER_DIR

ZIPAPP_SOURCE = INSTALLER_DIR / "test_zipapp"
ZIPAPP_OUTPUT = INSTALLER_DIR / "test_zipapp.pyz"

def make_zipapp():
    """Creates the zipapp from test_zipapp folder."""
    logger.info(f"Creating zipapp from {ZIPAPP_SOURCE} to {ZIPAPP_OUTPUT}")
    
    if not ZIPAPP_SOURCE.exists():
        logger.error(f"Source directory {ZIPAPP_SOURCE} does not exist.")
        return

    cmd = [
        sys.executable,
        "-m",
        "zipapp",
        str(ZIPAPP_SOURCE),
        "-o",
        str(ZIPAPP_OUTPUT)
    ]
    run_command(cmd)
    logger.info("Zipapp created successfully.")

def run_zipapp():
    """Runs the created zipapp."""
    if not ZIPAPP_OUTPUT.exists():
        logger.error(f"Zipapp {ZIPAPP_OUTPUT} not found. Run make-zipapp first.")
        return

    logger.info(f"Running zipapp {ZIPAPP_OUTPUT}")
    cmd = [
        sys.executable,
        str(ZIPAPP_OUTPUT)
    ]
    run_command(cmd)
