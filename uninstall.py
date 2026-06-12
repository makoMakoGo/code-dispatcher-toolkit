#!/usr/bin/env python3
"""Uninstaller for code-dispatcher runtime assets.

Removes only the files installed by ./install.py and leaves unrelated user files intact.
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

import runtime_assets


DEFAULT_INSTALL_DIR = "~/.code-dispatcher"


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Uninstall code-dispatcher")
    p.add_argument(
        "--install-dir",
        default=DEFAULT_INSTALL_DIR,
        help="Install directory (default: ~/.code-dispatcher)",
    )
    p.add_argument(
        "-y",
        "--yes",
        action="store_true",
        help="Do not prompt (non-interactive).",
    )
    return p.parse_args(argv)


def _unlink(path: Path) -> bool:
    try:
        path.unlink()
        return True
    except FileNotFoundError:
        return False
    except IsADirectoryError:
        return False


def _rmdir_if_empty(path: Path) -> None:
    try:
        if path.is_dir() and not any(path.iterdir()):
            path.rmdir()
    except OSError:
        return


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    install_dir = Path(args.install_dir).expanduser().resolve()

    if not install_dir.exists():
        print(f"Install dir not found: {install_dir}")
        return 0

    if not args.yes:
        print(f"About to remove code-dispatcher installed files from: {install_dir}")
        print("Proceed? [y/N] ", end="", flush=True)
        if input().strip().lower() not in ("y", "yes"):
            print("Aborted.")
            return 0

    try:
        targets = [
            runtime_assets.env_install_path(install_dir),
            runtime_assets.binary_path(install_dir),
        ]
        prompt_files = runtime_assets.uninstall_prompt_files()
    except RuntimeError as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return 1

    removed = 0

    for path in targets:
        if _unlink(path):
            removed += 1
            print(f"Removed: {path}")

    prompts_dir = install_dir / "prompts"
    for prompt_file in prompt_files:
        path = prompts_dir / prompt_file
        if _unlink(path):
            removed += 1
            print(f"Removed: {path}")
    _rmdir_if_empty(prompts_dir)

    _rmdir_if_empty(install_dir / runtime_assets.binary_install_dir_name())
    _rmdir_if_empty(install_dir)

    if removed == 0:
        print("Nothing removed (targets not found).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
