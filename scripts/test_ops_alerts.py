import http.server
import json
import pathlib
import ssl
import subprocess
import tempfile
import threading
import unittest
from unittest.mock import Mock, patch

from ops_alerts import Dispatcher, current_alerts, send_webhook, webhook_config


class AlertTests(unittest.TestCase):
    def test_debounce_throttle_recovery_and_bounded_retry(self):
        delivery = Mock(side_effect=[OSError("private canary"), None, None])
        dispatcher = Dispatcher(delivery)
        self.assertEqual(dispatcher.step(("server_errors",), 0), "idle")
        self.assertEqual(dispatcher.step(("server_errors",), 5), "delivery_failed")
        self.assertEqual(dispatcher.step(("server_errors",), 64), "idle")
        self.assertEqual(dispatcher.step(("server_errors",), 65), "delivered")
        self.assertEqual(dispatcher.step(("server_errors",), 100), "idle")
        self.assertEqual(dispatcher.step((), 105), "idle")
        self.assertEqual(dispatcher.step((), 125), "delivered")
        self.assertEqual(
            delivery.call_args.args[0],
            {"service": "gati", "state": "recovered", "alerts": []},
        )
        self.assertEqual(dispatcher.step((), 10000), "idle")

    def test_stale_and_sensitive_input_never_forwarded(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "monitor.json"
            self.assertEqual(current_alerts(path, 100), ("monitor_unavailable",))
            for value in [
                {
                    "version": 1,
                    "expires_at": 101,
                    "samples": [{"at": 100, "alerts": ["private-canary"]}],
                },
                {"version": 1, "expires_at": 99, "samples": [{"at": 98, "alerts": []}]},
                {"version": 1, "expires_at": 10000, "samples": []},
            ]:
                path.write_text(json.dumps(value))
                self.assertEqual(current_alerts(path, 100), ("monitor_unavailable",))
            path.write_text(
                json.dumps(
                    {
                        "version": 1,
                        "expires_at": 110,
                        "samples": [
                            {
                                "at": 100,
                                "alerts": ["server_errors"],
                                "unexpected": "private-canary",
                            }
                        ],
                    }
                )
            )
            delivery = Mock()
            dispatcher = Dispatcher(delivery)
            dispatcher.step(current_alerts(path, 100), 0)
            dispatcher.step(current_alerts(path, 100), 5)
            self.assertEqual(
                delivery.call_args.args[0],
                {"service": "gati", "state": "alert", "alerts": ["server_errors"]},
            )

    def test_credentials_require_private_regular_https_file(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "webhook.json"
            path.write_text(
                json.dumps(
                    {
                        "url": "https://example.invalid/notify",
                        "authorization": "synthetic",
                    }
                )
            )
            path.chmod(0o644)
            with self.assertRaises(ValueError):
                webhook_config(path)
            path.chmod(0o600)
            self.assertEqual(webhook_config(path)["authorization"], "synthetic")
            link = path.with_name("link")
            link.symlink_to(path)
            with self.assertRaises(OSError):
                webhook_config(link)
            for url in [
                "http://example.invalid",
                "https://user@example.invalid",
                "https://example.invalid/#secret",
            ]:
                path.write_text(json.dumps({"url": url, "authorization": ""}))
                with self.assertRaises(ValueError):
                    webhook_config(path)

    def test_local_tls_receiver_failure_retry_and_no_redirect(self):
        # Ephemeral synthetic certificate; never a production secret or endpoint.
        with tempfile.TemporaryDirectory() as directory:
            cert = pathlib.Path(directory) / "cert.pem"
            key = pathlib.Path(directory) / "key.pem"
            subprocess.run(
                [
                    "openssl",
                    "req",
                    "-x509",
                    "-newkey",
                    "rsa:2048",
                    "-nodes",
                    "-days",
                    "1",
                    "-subj",
                    "/CN=localhost",
                    "-addext",
                    "subjectAltName=DNS:localhost",
                    "-keyout",
                    str(key),
                    "-out",
                    str(cert),
                ],
                check=True,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
            )
            messages = []
            codes = iter([503, 204, 302])

            class Receiver(http.server.BaseHTTPRequestHandler):
                def do_POST(self):
                    messages.append(
                        (
                            self.headers.get("Authorization"),
                            json.loads(
                                self.rfile.read(int(self.headers["Content-Length"]))
                            ),
                        )
                    )
                    self.send_response(next(codes))
                    self.send_header("Location", "https://example.invalid/never-follow")
                    self.end_headers()

                def log_message(self, *_):
                    pass

            server = http.server.HTTPServer(("127.0.0.1", 0), Receiver)
            context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
            context.load_cert_chain(cert, key)
            server.socket = context.wrap_socket(server.socket, server_side=True)
            thread = threading.Thread(target=server.serve_forever, daemon=True)
            thread.start()
            trust = ssl.create_default_context(cafile=str(cert))
            config = {
                "url": f"https://localhost:{server.server_port}/notify",
                "authorization": "synthetic-private-canary",
            }
            try:
                with patch("ops_alerts.ssl.create_default_context", return_value=trust):
                    dispatcher = Dispatcher(
                        lambda payload: send_webhook(config, payload)
                    )
                    dispatcher.step(("cleanup_stale",), 0)
                    self.assertEqual(
                        dispatcher.step(("cleanup_stale",), 5), "delivery_failed"
                    )
                    self.assertEqual(
                        dispatcher.step(("cleanup_stale",), 65), "delivered"
                    )
                    with self.assertRaises(OSError):
                        send_webhook(
                            config,
                            {"service": "gati", "state": "recovered", "alerts": []},
                        )
                self.assertEqual(len(messages), 3)
                self.assertEqual(messages[0][1], messages[1][1])
                self.assertNotIn("canary", json.dumps([m[1] for m in messages]))
            finally:
                server.shutdown()
                server.server_close()
                thread.join(timeout=2)


if __name__ == "__main__":
    unittest.main()
