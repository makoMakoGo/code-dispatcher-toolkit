from __future__ import annotations

import contextlib
import io
import tempfile
import unittest
from pathlib import Path
from unittest import mock

import install
import runtime_assets
import uninstall


class RuntimeAssetsTests(unittest.TestCase):
    def test_prompt_assets_are_shared_by_installer_and_uninstaller(self) -> None:
        self.assertEqual(
            runtime_assets.prompt_install_files(),
            ["codex-prompt.md", "claude-prompt.md"],
        )
        self.assertEqual(
            runtime_assets.uninstall_prompt_files(),
            [
                "codex-prompt.md",
                "claude-prompt.md",
                "copilot-prompt.md",
                "gemini-prompt.md",
            ],
        )

    def test_skip_dispatcher_install_writes_manifest_assets(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            install_dir = Path(raw_dir)

            with contextlib.redirect_stdout(io.StringIO()):
                rc = install.main(["--install-dir", str(install_dir), "--skip-dispatcher"])

            self.assertEqual(rc, 0)
            self.assertTrue((install_dir / ".env").is_file())
            self.assertTrue((install_dir / "prompts" / "codex-prompt.md").is_file())
            self.assertTrue((install_dir / "prompts" / "claude-prompt.md").is_file())

    def test_uninstall_removes_manifest_assets_and_legacy_prompts(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            install_dir = Path(raw_dir)
            (install_dir / "bin").mkdir(parents=True)
            (install_dir / "prompts").mkdir(parents=True)
            runtime_assets.env_install_path(install_dir).write_text("env", encoding="utf-8")
            runtime_assets.binary_path(install_dir).write_text("bin", encoding="utf-8")
            for prompt_file in runtime_assets.uninstall_prompt_files():
                (install_dir / "prompts" / prompt_file).write_text("prompt", encoding="utf-8")
            unrelated = install_dir / "prompts" / "custom.md"
            unrelated.write_text("keep", encoding="utf-8")

            with contextlib.redirect_stdout(io.StringIO()):
                rc = uninstall.main(["--install-dir", str(install_dir), "--yes"])

            self.assertEqual(rc, 0)
            self.assertFalse(runtime_assets.env_install_path(install_dir).exists())
            self.assertFalse(runtime_assets.binary_path(install_dir).exists())
            for prompt_file in runtime_assets.uninstall_prompt_files():
                self.assertFalse((install_dir / "prompts" / prompt_file).exists())
            self.assertTrue(unrelated.is_file())

    def test_artifact_name_comes_from_manifest(self) -> None:
        with mock.patch("platform.system", return_value="Linux"), mock.patch(
            "platform.machine", return_value="x86_64"
        ):
            self.assertEqual(install._get_artifact_name(), "code-dispatcher-linux-amd64")


if __name__ == "__main__":
    unittest.main()
