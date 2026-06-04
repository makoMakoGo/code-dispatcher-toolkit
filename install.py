#!/usr/bin/env python3
"""Installer for code-dispatcher runtime assets.

Installs:
- code-dispatcher binary (downloaded from GitHub Release assets)
- per-backend default prompt templates under ~/.code-dispatcher/prompts
- ~/.code-dispatcher/.env template for runtime configuration

Targets:
- Linux x86_64
- macOS Apple Silicon
- Windows x86_64
"""

from __future__ import annotations

import argparse
import json
import os
import platform
import shutil
import sys
import urllib.error
import urllib.request
from pathlib import Path


DEFAULT_INSTALL_DIR = "~/.code-dispatcher"
RELEASE_REPO = "zhu-jl18/code-dispatcher-toolkit"
RELEASE_TAG = "latest"
HTTP_TIMEOUT_SEC = 30
ENV_TEMPLATE = Path(__file__).resolve().parent / "templates" / "runtime-config.env"

BACKENDS = ("codex", "claude")


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    p = argparse.ArgumentParser(description="Installer for code-dispatcher")
    p.add_argument(
        "--install-dir",
        default=DEFAULT_INSTALL_DIR,
        help="Install directory (default: ~/.code-dispatcher)",
    )
    p.add_argument(
        "--force",
        action="store_true",
        help="Overwrite existing files",
    )
    p.add_argument(
        "--skip-dispatcher",
        action="store_true",
        help="Skip installing code-dispatcher binary (only install runtime config/assets)",
    )
    return p.parse_args(argv)


def _ensure_dir(path: Path) -> None:
    path.mkdir(parents=True, exist_ok=True)


def _write_if_missing(path: Path, content: str, *, force: bool) -> None:
    _ensure_dir(path.parent)
    if path.exists() and not force:
        return
    path.write_text(content, encoding="utf-8")


def _install_prompts(install_dir: Path, *, force: bool) -> None:
    prompts_dir = install_dir / "prompts"
    _ensure_dir(prompts_dir)
    bundled_dir = Path(__file__).resolve().parent / "prompts"
    for backend in BACKENDS:
        dest = prompts_dir / f"{backend}-prompt.md"
        src = bundled_dir / f"{backend}-prompt.md"
        if dest.exists() and not force:
            continue
        if src.is_file():
            shutil.copy2(src, dest)
        else:
            _write_if_missing(dest, "", force=force)


def _install_env_template(install_dir: Path, *, force: bool) -> None:
    try:
        env_template = ENV_TEMPLATE.read_text(encoding="utf-8")
    except FileNotFoundError as e:
        raise RuntimeError(f"required env template not found: {ENV_TEMPLATE}") from e
    except OSError as e:
        raise RuntimeError(f"failed to read env template {ENV_TEMPLATE}: {e}") from e

    _write_if_missing(install_dir / ".env", env_template, force=force)


def _get_artifact_name() -> str:
    """Return release asset name for the current OS/arch."""
    system = platform.system()
    machine = platform.machine().lower()

    if system == "Windows" and machine in ("amd64", "x86_64"):
        return "code-dispatcher-windows-amd64.exe"
    if system == "Darwin" and machine in ("arm64", "aarch64"):
        return "code-dispatcher-darwin-arm64"
    if system == "Linux" and machine in ("amd64", "x86_64"):
        return "code-dispatcher-linux-amd64"

    raise RuntimeError(f"unsupported platform for release asset: {system}/{machine}")


def _github_headers() -> dict[str, str]:
    headers = {
        "Accept": "application/vnd.github+json",
        "User-Agent": "code-dispatcher-installer",
    }
    token = os.environ.get("GITHUB_TOKEN", "").strip()
    if token:
        headers["Authorization"] = f"Bearer {token}"
    return headers


def _load_release_metadata(repo: str, tag: str) -> dict:
    api_url = f"https://api.github.com/repos/{repo}/releases/tags/{tag}"
    req = urllib.request.Request(api_url, headers=_github_headers())
    try:
        with urllib.request.urlopen(req, timeout=HTTP_TIMEOUT_SEC) as resp:
            payload = resp.read().decode("utf-8")
    except urllib.error.HTTPError as e:
        if e.code == 404:
            raise FileNotFoundError(f"release tag not found: {repo}@{tag}") from e
        if e.code == 403:
            raise RuntimeError("GitHub API access denied or rate-limited; set GITHUB_TOKEN and retry") from e
        raise RuntimeError(f"failed to fetch release metadata ({e.code})") from e
    except urllib.error.URLError as e:
        raise RuntimeError(f"network error while fetching release metadata: {e.reason}") from e

    try:
        data = json.loads(payload)
    except json.JSONDecodeError as e:
        raise RuntimeError("invalid release metadata response from GitHub API") from e

    if not isinstance(data, dict):
        raise RuntimeError("unexpected release metadata structure from GitHub API")
    return data


def _resolve_asset_download_url(release: dict, asset_name: str) -> str:
    assets = release.get("assets")
    if not isinstance(assets, list):
        raise RuntimeError("release metadata missing assets list")

    for asset in assets:
        if not isinstance(asset, dict):
            continue
        if asset.get("name") == asset_name:
            url = asset.get("browser_download_url", "")
            if isinstance(url, str) and url:
                return url
            break

    available = sorted(
        str(asset.get("name"))
        for asset in assets
        if isinstance(asset, dict) and isinstance(asset.get("name"), str)
    )
    listed = ", ".join(available) if available else "<none>"
    raise FileNotFoundError(f"release asset not found: {asset_name} (available: {listed})")


def _download_to_path(url: str, out: Path) -> None:
    _ensure_dir(out.parent)
    tmp = out.with_suffix(out.suffix + ".tmp")
    req = urllib.request.Request(url, headers={"User-Agent": "code-dispatcher-installer"})
    try:
        with urllib.request.urlopen(req, timeout=HTTP_TIMEOUT_SEC) as resp:
            with tmp.open("wb") as f:
                shutil.copyfileobj(resp, f)
        os.replace(tmp, out)
    except urllib.error.URLError as e:
        raise RuntimeError(f"network error while downloading binary: {e.reason}") from e
    finally:
        if tmp.exists():
            tmp.unlink(missing_ok=True)


def _install_router_from_release(install_dir: Path, *, force: bool) -> Path:
    bin_dir = install_dir / "bin"
    _ensure_dir(bin_dir)
    exe_name = "code-dispatcher.exe" if os.name == "nt" else "code-dispatcher"
    out = bin_dir / exe_name

    if out.exists() and not force:
        return out

    asset_name = _get_artifact_name()
    release = _load_release_metadata(RELEASE_REPO, RELEASE_TAG)
    download_url = _resolve_asset_download_url(release, asset_name)
    _download_to_path(download_url, out)

    if os.name != "nt":
        try:
            out.chmod(0o755)
        except OSError:
            pass
    return out


def _print_path_hint(bin_path: Path) -> None:
    """Print PATH setup instructions based on platform and shell."""
    print("")
    print("PATH setup:")

    if os.name == "nt":
        print("  Windows PowerShell (current user env):")
        print(f'  [Environment]::SetEnvironmentVariable("Path", $env:Path + ";{bin_path.as_posix()}", "User")')
        print("")
        print("  Windows CMD (current user env):")
        print(f'  setx PATH "%PATH%;{bin_path}"')
        print("")
        print("  Git Bash (e.g. Claude Code on Windows):")
        print(f'  echo \'export PATH="{bin_path.as_posix()}:$PATH"\' >> ~/.bashrc')
    else:
        print("  Bash (for ~/.bashrc):")
        print(f'  echo \'export PATH="$PATH:{bin_path}"\' >> ~/.bashrc')
        print("")
        print("  Zsh (for ~/.zshrc):")
        print(f'  echo \'export PATH="$PATH:{bin_path}"\' >> ~/.zshrc')
        print("")
        print("  Fish:")
        print(f'  echo \'set -gx PATH "{bin_path}" $PATH\' >> ~/.config/fish/config.fish')


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    install_dir = Path(args.install_dir).expanduser().resolve()

    _ensure_dir(install_dir)

    _install_prompts(install_dir, force=args.force)
    try:
        _install_env_template(install_dir, force=args.force)
    except RuntimeError as e:
        print(f"ERROR: {e}", file=sys.stderr)
        return 1

    router_path: Path | None = None
    if not args.skip_dispatcher:
        try:
            router_path = _install_router_from_release(
                install_dir,
                force=args.force,
            )
        except (FileNotFoundError, RuntimeError) as e:
            print(f"ERROR: {e}", file=sys.stderr)
            print("Hint: verify network access and release assets, or use --skip-dispatcher to install config only.", file=sys.stderr)
            return 1

    print(f"Installed to: {install_dir}")
    print(f"- env:      {install_dir / '.env'}")
    print(f"- prompts:  {install_dir / 'prompts'} (*-prompt.md default templates)")
    if router_path is not None:
        print(f"- binary:   {router_path} (downloaded from GitHub release {RELEASE_REPO}@{RELEASE_TAG})")

    print("")
    print("Manual setup required:")
    print("- Copy selected skills/<skill-name>/ directories into each target CLI skills directory as needed.")
    print("- This installer does not auto-copy skills into Claude Code/Codex CLI/OpenCode roots.")

    if router_path is not None:
        _print_path_hint(install_dir / "bin")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
