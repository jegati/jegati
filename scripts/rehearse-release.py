#!/usr/bin/env python3
"""Local release activation, artifact comparison and compatible rollback rehearsal."""

import argparse, json, os, pathlib, shutil, subprocess, tempfile, time
from release import activate, compose, runtime_dir, verify_files, verify_running
from recovery import token

ROOT = pathlib.Path(__file__).resolve().parent.parent


def rehearse(source, out):
    source = source.resolve()
    out = out.resolve()
    out.mkdir(parents=True, exist_ok=True)
    verify_files(source)
    project = "gati-release-lab-" + os.urandom(4).hex()
    connector = project + "-connector"
    started = time.time()
    with (
        tempfile.TemporaryDirectory(prefix="gati-release-rehearsal-") as directory,
        (out / "commands.log").open("wb") as log,
    ):
        parent = pathlib.Path(directory)
        a = parent / "a"
        b = parent / "b"
        shutil.copytree(source, a)
        shutil.copytree(source, b, copy_function=os.link)
        manifest = verify_files(a)

        def command(*args, **kw):
            return subprocess.check_output(args, stderr=log, **kw).decode().strip()

        try:
            activate(a, project, manifest, 1, False)
            web = compose(a, project, "ps", "-q", "web")
            command("docker", "network", "connect", project + "_backend", web)
            try:
                try:
                    verify_running(a, project, manifest, 1, False)
                except ValueError as e:
                    if str(e) != "runtime network attachments differ":
                        raise
                else:
                    raise AssertionError(
                        "unexpected network attachment passed release verification"
                    )
            finally:
                command("docker", "network", "disconnect", project + "_backend", web)
            verify_running(a, project, manifest, 1, False)
            store = compose(a, project, "ps", "-q", "valkey")
            before = json.loads(command("docker", "inspect", store))[0]["State"][
                "StartedAt"
            ]
            command(
                "go",
                "build",
                "-trimpath",
                "-buildvcs=false",
                "-o",
                str(parent / "probe"),
                "./scripts/deployment-probe",
                cwd=ROOT,
                env={**os.environ, "CGO_ENABLED": "0"},
            )
            command(
                "docker",
                "run",
                "-d",
                "--name",
                connector,
                "--network",
                project + "_edge",
                "--ip",
                "172.30.10.2",
                "--read-only",
                "--cap-drop",
                "ALL",
                "--security-opt",
                "no-new-privileges:true",
                "--log-driver",
                "none",
                "-p",
                "127.0.0.1::8090",
                "-v",
                str(parent / "probe") + ":/probe:ro",
                "--entrypoint",
                "/probe",
                manifest["images"]["api"]["reference"],
            )
            port = int(command("docker", "port", connector, "8090").rsplit(":", 1)[1])
            base = f"http://localhost:{port}"
            command(
                "python3",
                str(ROOT / "scripts/verify-public.py"),
                "--release",
                str(a),
                "--url",
                base,
            )
            import http.client

            cap = token()

            def request(method, path, data=None):
                c = http.client.HTTPConnection("localhost", port, timeout=10)
                try:
                    c.request(
                        method,
                        path,
                        json.dumps(data) if data is not None else None,
                        {
                            "Authorization": "Bearer " + cap,
                            "Content-Type": "application/json",
                        },
                    )
                    r = c.getresponse()
                    raw = r.read(50000)
                    return r.status, json.loads(raw) if raw else None
                finally:
                    c.close()

            status, created = request(
                "POST",
                "/api/signals",
                {
                    "cell": "tirana-v1:100:55:55",
                    "radius_km": 3,
                    "availability_minutes": 30,
                },
            )
            assert status == 200
            # The second immutable directory has identical code/config: this specifically
            # tests stable mounts and rollback mechanics, not arbitrary schema migrations.
            activate(b, project, verify_files(b), 1, False, a)
            assert compose(b, project, "ps", "-q", "valkey") == store
            assert (
                json.loads(command("docker", "inspect", store))[0]["State"]["StartedAt"]
                == before
            )
            assert request("GET", "/api/signal")[0] == 200
            activate(a, project, manifest, 1, False, b)
            assert compose(a, project, "ps", "-q", "valkey") == store
            assert (
                json.loads(command("docker", "inspect", store))[0]["State"]["StartedAt"]
                == before
            )
            status, state = request("GET", "/api/signal")
            assert status == 200 and state["expires_at"] == created["expires_at"]
            assert request("DELETE", "/api/signal")[0] == 204
            command(
                "python3",
                str(ROOT / "scripts/verify-public.py"),
                "--release",
                str(a),
                "--url",
                base,
            )
            report = {
                "synthetic": True,
                "source_revision": manifest["source_revision"],
                "checks": [
                    "immutable release verification and actual combined-role activation",
                    "unexpected live network attachment rejected, restored topology accepted",
                    "served browser assets and effective config hash match",
                    "two compatible release-directory switches preserve store identity/start time and live capability",
                    "return to first release preserves original expiry and cancellation works",
                ],
                "wall_seconds": round(time.time() - started, 2),
                "scope": "same-code/config local rollback mechanics; no schema migration, external host, real Cloudflare or public publication",
            }
            (out / "release-rehearsal.json").write_text(
                json.dumps(report, indent=2) + "\n"
            )
            print(json.dumps(report, indent=2))
        finally:
            subprocess.run(["docker", "rm", "-f", connector], stdout=log, stderr=log)
            compose(a, project, "down", "--remove-orphans", "--volumes")
            shutil.rmtree(runtime_dir(a, project), ignore_errors=True)


if __name__ == "__main__":
    p = argparse.ArgumentParser()
    p.add_argument("--release", type=pathlib.Path, required=True)
    p.add_argument("--output", type=pathlib.Path, required=True)
    a = p.parse_args()
    rehearse(a.release, a.output)
