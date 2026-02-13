import subprocess
import logging
import sys

logging.basicConfig(level=logging.INFO, format="%(message)s")
logger = logging.getLogger("builder")


def run_command(command, cwd=None, env=None, check=True):
    logger.info(f"Exec: {' '.join(command)}")
    try:
        subprocess.run(command, cwd=cwd, env=env, check=check, text=True)
    except subprocess.CalledProcessError as e:
        logger.error(f"Command failed: {e}")
        if check:
            sys.exit(1)
