"""Shared runtime asset manifest helpers for installer tooling."""

from __future__ import annotations

import functools
import json
import os
import platform
from pathlib import Path
from typing import Any


PROJECT_ROOT = Path(__file__).resolve().parent
MANIFEST_PATH = PROJECT_ROOT / "code-dispatcher" / "runtime_assets.json"


def _load_manifest() -> dict[str, Any]:
    try:
        data = json.loads(MANIFEST_PATH.read_text(encoding="utf-8"))
    except FileNotFoundError as e:
        raise RuntimeError(f"runtime asset manifest not found: {MANIFEST_PATH}") from e
    except json.JSONDecodeError as e:
        raise RuntimeError(
            f"invalid runtime asset manifest: {MANIFEST_PATH}: "
            f"line {e.lineno}, column {e.colno}: {e.msg}"
        ) from e
    except OSError as e:
        raise RuntimeError(f"failed to read runtime asset manifest: {MANIFEST_PATH}: {e}") from e

    if not isinstance(data, dict):
        raise RuntimeError("runtime asset manifest must be a JSON object")
    _validate_binary_names(data)
    return data


def _validate_binary_names(data: dict[str, Any]) -> None:
    binary = data.get("binary")
    if not isinstance(binary, dict):
        raise RuntimeError("runtime asset manifest binary must be an object")
    names = binary.get("name_by_os")
    if not isinstance(names, dict):
        raise RuntimeError("runtime asset manifest binary.name_by_os must be an object")
    for key in ("default", "nt"):
        value = names.get(key)
        if not isinstance(value, str) or not value.strip():
            raise RuntimeError(
                f"runtime asset manifest binary.name_by_os.{key} must be a non-empty string"
            )


@functools.lru_cache(maxsize=1)
def _manifest() -> dict[str, Any]:
    """Load the manifest on first use so importing this module never raises.

    Lazy loading keeps `install.py --help` / `uninstall.py --help` working and
    lets callers surface RuntimeError through their own error handling instead
    of dying with an import-time traceback.
    """
    return _load_manifest()


def _mapping(value: Any, label: str) -> dict[str, Any]:
    if not isinstance(value, dict):
        raise RuntimeError(f"runtime asset manifest {label} must be an object")
    return value


def _list(value: Any, label: str) -> list[Any]:
    if not isinstance(value, list):
        raise RuntimeError(f"runtime asset manifest {label} must be a list")
    return value


def _string(value: Any, label: str) -> str:
    if not isinstance(value, str) or not value.strip():
        raise RuntimeError(f"runtime asset manifest {label} must be a non-empty string")
    return value


def runtime_config() -> dict[str, Any]:
    return _mapping(_manifest().get("runtime_config"), "runtime_config")


def binary() -> dict[str, Any]:
    return _mapping(_manifest().get("binary"), "binary")


def backend_assets() -> list[dict[str, Any]]:
    backends = _list(_manifest().get("backends"), "backends")
    return [_mapping(backend, f"backends[{idx}]") for idx, backend in enumerate(backends)]


def env_template_path() -> Path:
    return PROJECT_ROOT / _string(runtime_config().get("template"), "runtime_config.template")


def env_install_path(install_dir: Path) -> Path:
    return install_dir / _string(runtime_config().get("install_path"), "runtime_config.install_path")


def binary_install_dir_name() -> str:
    return _string(binary().get("install_dir"), "binary.install_dir")


def binary_name() -> str:
    names = _mapping(binary().get("name_by_os"), "binary.name_by_os")
    key = "nt" if os.name == "nt" else "default"
    return _string(names.get(key), f"binary.name_by_os.{key}")


def binary_path(install_dir: Path) -> Path:
    return install_dir / binary_install_dir_name() / binary_name()


def prompt_template_path(backend: dict[str, Any]) -> Path:
    return PROJECT_ROOT / _string(backend.get("prompt_template"), "backend.prompt_template")


def prompt_file(backend: dict[str, Any]) -> str:
    return _string(backend.get("prompt_file"), "backend.prompt_file")


def prompt_install_files() -> list[str]:
    return [prompt_file(backend) for backend in backend_assets()]


def legacy_prompt_files() -> list[str]:
    legacy = _mapping(_manifest().get("legacy_uninstall"), "legacy_uninstall")
    files = _list(legacy.get("prompt_files"), "legacy_uninstall.prompt_files")
    return [_string(file, "legacy_uninstall.prompt_files[]") for file in files]


def uninstall_prompt_files() -> list[str]:
    return prompt_install_files() + legacy_prompt_files()


def artifact_name_for_current_platform() -> str:
    system = platform.system()
    machine = platform.machine().lower()
    release_assets = _list(binary().get("release_assets"), "binary.release_assets")

    for idx, raw_asset in enumerate(release_assets):
        asset = _mapping(raw_asset, f"binary.release_assets[{idx}]")
        if _string(asset.get("system"), "release_asset.system") != system:
            continue
        machines = _list(asset.get("machines"), "release_asset.machines")
        normalized = [_string(item, "release_asset.machines[]").lower() for item in machines]
        if machine in normalized:
            return _string(asset.get("asset"), "release_asset.asset")

    raise RuntimeError(f"unsupported platform for release asset: {system}/{machine}")
