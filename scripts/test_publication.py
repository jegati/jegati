import pathlib
import subprocess
import sys
import tempfile
import unittest

from check_publication import URL_FIXTURES, inspect_blob


class PublicationTests(unittest.TestCase):
    def test_home_paths_and_credentials_are_detected(self):
        examples = [
            "/home/" + "synthetic-person/work",
            "/Users/" + "synthetic-person/work",
            "C:" + "\\Users\\" + "synthetic-person\\work",
            "-----BEGIN " + "OPENSSH PRIVATE KEY-----",
            "ghp_" + "A" * 36,
            "github_pat_" + "B" * 40,
            "AKIA" + "C" * 16,
            "https://" + "synthetic:password@example.invalid",
        ]
        for value in examples:
            with self.subTest(value=value[:8]):
                findings, skipped = inspect_blob("notes.md", value.encode())
                self.assertTrue(findings)
                self.assertFalse(skipped)

    def test_project_identity_and_portable_examples_allowed(self):
        data = (
            "jegati@proton.me jamgati.com OVHcloud Cloudflare "
            "$HOME/work ~/work /opt/gati /app /run/secrets "
            "https://example.invalid https://github.com/jegati/jegati"
        ).encode()
        self.assertEqual(inspect_blob("README.md", data), ([], False))

    def test_sensitive_paths_blocked_even_for_empty_or_binary_files(self):
        for name in [
            ".runtime/token",
            "nested/.env",
            ".env.production",
            "host.key",
            "web/dist/index.html",
            "reports/local/report.json",
            "commands.txt",
            "nested/node_modules/file",
            "id_ed25519",
        ]:
            with self.subTest(name=name):
                self.assertTrue(inspect_blob(name, b"")[0])
                self.assertTrue(inspect_blob(name, b"\x00\xff")[0])
        self.assertEqual(inspect_blob(".env.example", b""), ([], False))

    def test_binary_skip_is_explicit(self):
        self.assertEqual(inspect_blob("map.gz", b"\x00\xff"), ([], True))

    def test_fixture_exceptions_are_exact_and_file_scoped(self):
        for name, value in URL_FIXTURES:
            self.assertEqual(inspect_blob(name, value.encode()), ([], False))
            self.assertTrue(inspect_blob("other_test.go", value.encode())[0])
            changed = value.replace("user:", "another-user:")
            self.assertTrue(inspect_blob(name, changed.encode())[0])
            self.assertTrue(inspect_blob(name, (value + " " + changed).encode())[0])

    def test_index_is_scanned_not_worktree_or_ignored_files(self):
        script = pathlib.Path(__file__).with_name("check_publication.py").resolve()
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)

            def git(*args):
                subprocess.run(
                    ["git", *args], cwd=root, check=True, capture_output=True
                )

            def scan():
                return subprocess.run(
                    [sys.executable, str(script)],
                    cwd=root,
                    capture_output=True,
                    text=True,
                )

            git("init", "-q")
            secret = "ghp_" + "D" * 36
            (root / "note.md").write_text(secret)
            git("add", "note.md")
            (root / "note.md").write_text("clean worktree")
            result = scan()
            self.assertEqual(result.returncode, 1)
            self.assertIn("provider credential pattern", result.stdout)
            self.assertNotIn(secret, result.stdout + result.stderr)
            git("add", "note.md")
            (root / ".git/info/exclude").write_text("/commands.txt\n")
            (root / "commands.txt").write_text(secret)
            (root / "link").symlink_to("commands.txt")
            git("add", "link")
            self.assertEqual(scan().returncode, 0)
            git("add", "-f", "commands.txt")
            self.assertEqual(scan().returncode, 1)
            git("rm", "--cached", "commands.txt")
            git("rm", "--cached", "note.md")
            self.assertEqual(scan().returncode, 0)

    def test_missing_repository_fails_closed(self):
        script = pathlib.Path(__file__).with_name("check_publication.py").resolve()
        with tempfile.TemporaryDirectory() as directory:
            result = subprocess.run(
                [sys.executable, str(script)], cwd=directory, capture_output=True
            )
            self.assertEqual(result.returncode, 2)


if __name__ == "__main__":
    unittest.main()
