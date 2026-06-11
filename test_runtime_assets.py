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
            self.assertTrue(runtime_assets.env_install_path(install_dir).is_file())
            for prompt_file in runtime_assets.prompt_install_files():
                self.assertTrue((install_dir / "prompts" / prompt_file).is_file())

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


class LazyManifestTests(unittest.TestCase):
    def setUp(self) -> None:
        runtime_assets._manifest.cache_clear()
        self.addCleanup(runtime_assets._manifest.cache_clear)

    def test_accessors_raise_runtime_error_when_manifest_missing(self) -> None:
        missing = Path(tempfile.gettempdir()) / "code-dispatcher-no-such-manifest.json"
        with mock.patch.object(runtime_assets, "MANIFEST_PATH", missing):
            with self.assertRaisesRegex(RuntimeError, "runtime asset manifest not found"):
                runtime_assets.prompt_install_files()

    def test_help_does_not_load_manifest(self) -> None:
        with mock.patch.object(
            runtime_assets, "_load_manifest", side_effect=AssertionError("manifest loaded during --help")
        ):
            for module in (install, uninstall):
                with self.subTest(module=module.__name__):
                    with contextlib.redirect_stdout(io.StringIO()):
                        with self.assertRaises(SystemExit) as ctx:
                            module.main(["--help"])
                    self.assertEqual(ctx.exception.code, 0)

    def test_install_reports_clean_error_when_manifest_missing(self) -> None:
        missing = Path(tempfile.gettempdir()) / "code-dispatcher-no-such-manifest.json"
        with tempfile.TemporaryDirectory() as raw_dir:
            with mock.patch.object(runtime_assets, "MANIFEST_PATH", missing):
                stderr = io.StringIO()
                with contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(stderr):
                    rc = install.main(["--install-dir", raw_dir, "--skip-dispatcher"])
        self.assertEqual(rc, 1)
        self.assertIn("ERROR: runtime asset manifest not found", stderr.getvalue())

    def test_uninstall_reports_clean_error_when_manifest_missing(self) -> None:
        missing = Path(tempfile.gettempdir()) / "code-dispatcher-no-such-manifest.json"
        with tempfile.TemporaryDirectory() as raw_dir:
            with mock.patch.object(runtime_assets, "MANIFEST_PATH", missing):
                stderr = io.StringIO()
                with contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(stderr):
                    rc = uninstall.main(["--install-dir", raw_dir, "--yes"])
        self.assertEqual(rc, 1)
        self.assertIn("ERROR: runtime asset manifest not found", stderr.getvalue())

    def test_install_reports_clean_error_when_manifest_malformed(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            bad_manifest = Path(raw_dir) / "runtime_assets.json"
            bad_manifest.write_text("{not valid json", encoding="utf-8")
            with mock.patch.object(runtime_assets, "MANIFEST_PATH", bad_manifest):
                stderr = io.StringIO()
                with contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(stderr):
                    rc = install.main(["--install-dir", str(Path(raw_dir) / "x"), "--skip-dispatcher"])
        self.assertEqual(rc, 1)
        self.assertIn("ERROR: invalid runtime asset manifest", stderr.getvalue())

    def test_uninstall_reports_clean_error_when_manifest_malformed(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            bad_manifest = Path(raw_dir) / "runtime_assets.json"
            bad_manifest.write_text("{not valid json", encoding="utf-8")
            with mock.patch.object(runtime_assets, "MANIFEST_PATH", bad_manifest):
                stderr = io.StringIO()
                with contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(stderr):
                    rc = uninstall.main(["--install-dir", raw_dir, "--yes"])
        self.assertEqual(rc, 1)
        self.assertIn("ERROR: invalid runtime asset manifest", stderr.getvalue())

    def test_manifest_loads_once_across_accessors(self) -> None:
        fake_manifest = {
            "binary": {"install_dir": "bin"},
            "backends": [],
        }
        with mock.patch.object(
            runtime_assets, "_load_manifest", return_value=fake_manifest
        ) as load_manifest:
            runtime_assets.binary()
            runtime_assets.backend_assets()
        load_manifest.assert_called_once_with()


if __name__ == "__main__":
    unittest.main()
