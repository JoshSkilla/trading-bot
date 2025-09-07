# pyscripts/cli.py
from pathlib import Path
import os
import subprocess

from . import PROJECT_ROOT, BIN_DIR, BOT_BIN  # from __init__.py

def _ensure_built():
    """Build the Go CLI if the binary doesn't exist."""
    if not BOT_BIN.exists():
        BIN_DIR.mkdir(parents=True, exist_ok=True)
        print("Building Go CLI -> bin/bot …")
        subprocess.run(
            ["go", "build", "-o", str(BOT_BIN), "./cmd/bot"],
            cwd=str(PROJECT_ROOT),
            check=True,
        )

# Runs bot command till completion, capturing output
def launch_bot(args, env=None, check=False):
    """
    Run the bot CLI and capture output.
    Example: launch_bot(["create", "-name", "PyPortfolio", "-cash", "1000"])
    """
    _ensure_built()
    full_env = os.environ.copy()
    if env:
        full_env.update(env)
    # Make sure BOT_PATH is available to the child process
    full_env.setdefault("BOT_PATH", str(PROJECT_ROOT))

    return subprocess.run(
        [str(BOT_BIN), *args],
        cwd=str(PROJECT_ROOT),
        env=full_env,
        text=True,
        capture_output=True,
        check=check,  # True -> raise on non-zero exit
    )

# Live streams bot outputs as running
def stream_bot(args, env=None):
    """
    Run the bot CLI and stream stdout/stderr live (good for long-running runs).
    """
    _ensure_built()
    full_env = os.environ.copy()
    if env:
        full_env.update(env)
    full_env.setdefault("BOT_PATH", str(PROJECT_ROOT))

    with subprocess.Popen(
        [str(BOT_BIN), *args],
        cwd=str(PROJECT_ROOT),
        env=full_env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        bufsize=1,
    ) as p:
        for line in p.stdout:
            print(line, end="")
        return p.wait()