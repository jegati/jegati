"""Offline guard for the Git index, not a history audit or complete secret scanner."""

import json
import pathlib
import re
import subprocess
import sys


# Project email/domain are public by choice. This does not flag email addresses,
# public hostnames, commit timestamps or reproducible benchmark hardware details.
RULES = {
    "personal home path": re.compile(
        r"/(?:home|Users)/[A-Za-z0-9_.-]+(?:/|\b)|"
        r"[A-Za-z]:[\\/]+Users[\\/]+[A-Za-z0-9_.-]+",
        re.I,
    ),
    "private key marker": re.compile(r"-----BEGIN (?:[A-Z0-9]+ )*PRIVATE KEY-----"),
    "provider credential pattern": re.compile(
        r"\b(?:gh[pousr]_[A-Za-z0-9]{36,}|github_pat_[A-Za-z0-9_]{22,}|"
        r"(?:AKIA|ASIA)[A-Z0-9]{16}|xox[baprs]-[A-Za-z0-9-]{20,})\b"
    ),
    "credential-bearing URL": re.compile(
        r"\b[a-z][a-z0-9+.-]*://[^\s/:@]+:[^\s/@]+@[^\s\"'<>]+", re.I
    ),
}
BLOCKED_PARTS = {
    ".runtime",
    "node_modules",
    "__pycache__",
    ".tmp",
    "coverage",
    "test-results",
    "playwright-report",
}
BLOCKED_PREFIXES = ("web/dist/", "reports/local/", "bin/")
# Exact synthetic rejection fixtures, not exemptions for entire test files.
URL_FIXTURES = {
    (
        "internal/notification/transport_test.go",
        "https://" + "user:secret@push.example.org/path",
    ),
    (
        "internal/simulation/scenario_test.go",
        "http://" + "user:password@127.0.0.1:8082",
    ),
}


def blocked_path(name):
    path = pathlib.PurePosixPath(name)
    return (
        bool(set(path.parts) & BLOCKED_PARTS)
        or name.startswith(BLOCKED_PREFIXES)
        or name == "commands.txt"
        or (path.name.startswith(".env") and path.name != ".env.example")
        or path.suffix.lower() in {".pem", ".key", ".p12", ".pfx"}
        or path.name in {"id_rsa", "id_ed25519", "id_ecdsa", "credentials"}
    )


def inspect_blob(name, data):
    findings = []
    if blocked_path(name):
        findings.append((0, "runtime/credential or local-only filename"))
    if RULES["personal home path"].search(name):
        findings.append((0, "personal home path in filename"))
    # No PDF/archive parser: binary contents need separate human review.
    try:
        text = data.decode("utf-8")
    except UnicodeDecodeError:
        return findings, True
    if "\x00" in text:
        return findings, True
    for line_number, line in enumerate(text.splitlines(), 1):
        for category, pattern in RULES.items():
            matches = pattern.finditer(line)
            if any(
                category != "credential-bearing URL"
                or (name, match.group()) not in URL_FIXTURES
                for match in matches
            ):
                findings.append((line_number, category))
    return findings, False


def scan_index():
    entries = subprocess.check_output(["git", "ls-files", "--stage", "-z"])
    count = skipped = failures = 0
    # Read staged blobs, never working-tree paths or symlink targets. A staged
    # secret must still fail even when the working copy was subsequently cleaned.
    with subprocess.Popen(
        ["git", "cat-file", "--batch"], stdin=subprocess.PIPE, stdout=subprocess.PIPE
    ) as objects:
        for entry in entries.split(b"\x00"):
            if not entry:
                continue
            metadata, raw_name = entry.split(b"\t", 1)
            mode, oid, stage = metadata.split()
            name = raw_name.decode("utf-8", errors="backslashreplace")
            count += 1
            if stage != b"0" or mode not in {b"100644", b"100755", b"120000"}:
                print(f"{json.dumps(name)}: unresolved or unsupported index entry")
                failures += 1
                continue
            objects.stdin.write(oid + b"\n")
            objects.stdin.flush()
            header = objects.stdout.readline().split()
            if len(header) != 3 or header[1] != b"blob":
                raise ValueError("cannot read indexed blob")
            data = objects.stdout.read(int(header[2]))
            if len(data) != int(header[2]) or objects.stdout.read(1) != b"\n":
                raise ValueError("incomplete indexed blob")
            findings, binary = inspect_blob(name, data)
            skipped += binary
            for line, category in findings:
                print(f"{json.dumps(name)}:{line}: {category}")
            failures += len(findings)
        objects.stdin.close()
        if objects.wait() != 0:
            raise ValueError("Git object reader failed")
    print(
        f"Index check: {count} entries, {failures} findings; "
        f"{skipped} binary contents skipped (filenames checked)."
    )
    print("Scope: staged snapshot only; no history, metadata or untracked-file audit.")
    return 1 if failures else 0


if __name__ == "__main__":
    try:
        sys.exit(scan_index())
    except (OSError, ValueError, subprocess.SubprocessError):
        print("Publication check could not finish; no pass recorded.", file=sys.stderr)
        sys.exit(2)
