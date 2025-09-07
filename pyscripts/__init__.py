# pyscripts/__init__.py
from pathlib import Path
import os

# On import, loads .env
# Soft-fail if env may already be set (idempotent)
try:
    from dotenv import load_dotenv, find_dotenv 
    load_dotenv(find_dotenv(), override=True)
except Exception:
    pass

# Path constants
PROJECT_ROOT = Path(os.environ["BOT_PATH"]).resolve()  # will KeyError if missing -> fail fast (good)
BIN_DIR = PROJECT_ROOT / "bin"
BOT_BIN = BIN_DIR / "bot"

__all__ = ["PROJECT_ROOT", "BIN_DIR", "BOT_BIN"]