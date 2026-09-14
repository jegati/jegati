"""Deliver only fixed operational labels; never forward monitoring documents."""

import argparse
import http.client
import json
import os
import pathlib
import ssl
import stat
import time
import urllib.parse

LABELS = {"monitor_unavailable", "server_errors", "slow_requests"} | {
    worker + "_" + kind
    for worker in ["matcher", "publisher", "cleanup", "push", "view"]
    for kind in ["error", "stale", "deadline_lag"]
}


def current_alerts(path, now):
    try:
        with path.open("rb") as stream:
            raw = stream.read(8 * 1024 * 1024 + 1)
        if len(raw) > 8 * 1024 * 1024:
            raise ValueError("oversized monitor file")
        value = json.loads(raw)
        if (
            type(value["version"]) is not int
            or value["version"] != 1
            or type(value["expires_at"]) is not int
            or not now < value["expires_at"] <= now + 180
        ):
            raise ValueError("stale or unknown monitor file")
        row = value["samples"][-1]
        if type(row["at"]) is not int or not now - 180 <= row["at"] <= now + 5:
            raise ValueError("stale monitor sample")
        labels = row["alerts"]
        if (
            not isinstance(labels, list)
            or len(labels) > len(LABELS)
            or not all(isinstance(v, str) and v in LABELS for v in labels)
        ):
            raise ValueError("unexpected alert labels")
        return tuple(sorted(set(labels)))
    except (OSError, ValueError, KeyError, IndexError, TypeError):
        return ("monitor_unavailable",)


def read_private_config(path):
    # Keep secrets outside argv, process environment, logs and Git.
    flags = os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK
    with os.fdopen(os.open(path, flags), "r") as stream:
        st = os.fstat(stream.fileno())
        if (
            not stat.S_ISREG(st.st_mode)
            or st.st_size > 4096
            or st.st_mode & 0o077
            or st.st_uid != os.geteuid()
        ):
            raise ValueError("config must be private and owned by this user")
        raw = stream.read(4097)
        if len(raw) > 4096:
            raise ValueError("oversized private config")
        return raw


def webhook_config(path):
    value = json.loads(read_private_config(path))
    if set(value) != {"url", "authorization"} or not all(
        isinstance(v, str) for v in value.values()
    ):
        raise ValueError("invalid webhook config")
    if (
        len(value["url"]) > 2048
        or len(value["authorization"]) > 1024
        or any(c in value["url"] + value["authorization"] for c in "\r\n")
    ):
        raise ValueError("invalid webhook config")
    u = urllib.parse.urlsplit(value["url"])
    if u.scheme != "https" or not u.hostname or u.username or u.password or u.fragment:
        raise ValueError("webhook requires a credential-free HTTPS authority")
    if u.port is not None and not 1 <= u.port <= 65535:
        raise ValueError("invalid webhook port")
    return value


def send_webhook(config, payload):
    u = urllib.parse.urlsplit(config["url"])
    connection = http.client.HTTPSConnection(
        u.hostname, u.port or 443, timeout=5, context=ssl.create_default_context()
    )
    headers = {"Content-Type": "application/json"}
    if config["authorization"]:
        headers["Authorization"] = config["authorization"]
    try:
        connection.request(
            "POST",
            urllib.parse.urlunsplit(("", "", u.path or "/", u.query, "")),
            json.dumps(payload),
            headers,
        )
        response = connection.getresponse()
        # No redirect, proxy environment, response logging or unlimited body read.
        if not 200 <= response.status < 300:
            raise OSError("webhook rejected")
    finally:
        connection.close()


class Dispatcher:
    def __init__(self, deliver):
        self.deliver = deliver
        self.candidate = None
        self.stable = 0
        self.sent = ()
        self.last_sent = None
        self.next_attempt = 0
        self.failures = 0

    def step(self, labels, now):
        if not all(v in LABELS for v in labels):
            labels = ("monitor_unavailable",)
        labels = tuple(sorted(set(labels)))
        self.stable = min(self.stable + 1, 2) if labels == self.candidate else 1
        self.candidate = labels
        due = labels != self.sent or (
            bool(labels) and self.last_sent is not None and now - self.last_sent >= 1800
        )
        if self.stable < 2 or not due or now < self.next_attempt:
            return "idle"
        payload = {
            "service": "gati",
            "state": "alert" if labels else "recovered",
            "alerts": list(labels),
        }
        try:
            self.deliver(payload)
        except (OSError, ValueError, http.client.HTTPException):
            self.failures = min(self.failures + 1, 4)
            self.next_attempt = now + min(60 * 2 ** (self.failures - 1), 300)
            return "delivery_failed"
        self.sent = labels
        self.last_sent = now
        self.next_attempt = now + 60
        self.failures = 0
        return "delivered"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--monitor-file", type=pathlib.Path, required=True)
    delivery = parser.add_mutually_exclusive_group(required=True)
    delivery.add_argument("--webhook-file", type=pathlib.Path)
    delivery.add_argument("--email-env-file", type=pathlib.Path)
    args = parser.parse_args()
    try:
        if args.email_env_file:
            from ops_email import email_config, send_email

            config = email_config(args.email_env_file)
            deliver = lambda payload: send_email(config, payload)
        else:
            config = webhook_config(args.webhook_file)
            deliver = lambda payload: send_webhook(config, payload)
    except (OSError, ValueError, TypeError):
        raise SystemExit("Cannot read private alert configuration.") from None
    dispatcher = Dispatcher(deliver)
    while True:
        result = dispatcher.step(
            current_alerts(args.monitor_file, time.time()), time.monotonic()
        )
        if result != "idle":
            print("Operational alert: " + result, flush=True)
        time.sleep(5)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        pass
