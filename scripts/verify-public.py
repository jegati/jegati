#!/usr/bin/env python3
"""Compare a chosen public endpoint to a separately obtained release manifest."""

import argparse, hashlib, http.client, json, pathlib, urllib.parse
from release import verify_files


def verify(root, base):
    manifest = verify_files(root)
    u = urllib.parse.urlsplit(base)
    if (
        u.username
        or u.password
        or u.query
        or u.fragment
        or u.path not in ("", "/")
        or u.scheme not in ("http", "https")
    ):
        raise ValueError("use an origin URL without credentials/path/query")
    if u.scheme == "http" and u.hostname not in ("localhost", "127.0.0.1", "::1"):
        raise ValueError("public verification requires HTTPS")

    def fetch(path):
        client = (
            http.client.HTTPSConnection
            if u.scheme == "https"
            else http.client.HTTPConnection
        )(u.hostname, u.port, timeout=15)
        try:
            client.request("GET", path, headers={"Accept-Encoding": "identity"})
            r = client.getresponse()
            body = r.read(5000001)
            if r.status != 200 or len(body) > 5000000:
                raise ValueError("public asset missing, redirected or oversized")
            if r.getheader("Set-Cookie"):
                raise ValueError(
                    "unexpected public response cookie; review edge configuration"
                )
            return body
        finally:
            client.close()

    for name, want in manifest["browser_assets"].items():
        if name.startswith("/") or ".." in pathlib.PurePosixPath(name).parts:
            raise ValueError("invalid asset path")
        path = "/" if name == "index.html" else "/" + urllib.parse.quote(name, safe="/")
        if hashlib.sha256(fetch(path)).hexdigest() != want:
            raise ValueError("served browser artifact differs: " + name)
    envelope = json.loads(fetch("/api/config"))
    if envelope.get("sha256") != manifest["effective_config_sha256"]:
        raise ValueError("published functional configuration differs")
    print(
        "Verified browser asset bytes and advertised config hash. This does not attest backend code, host behavior or provider logging."
    )


if __name__ == "__main__":
    p = argparse.ArgumentParser()
    p.add_argument("--release", type=pathlib.Path, required=True)
    p.add_argument("--url", required=True)
    a = p.parse_args()
    verify(a.release.resolve(), a.url)
