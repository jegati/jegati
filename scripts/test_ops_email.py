import email
import pathlib
import smtplib
import socketserver
import ssl
import subprocess
import sys
import tempfile
import threading
import unittest
from unittest.mock import Mock, patch

from ops_alerts import Dispatcher
from ops_email import email_config, send_email


class EmailTests(unittest.TestCase):
    def test_cli_rejects_bad_email_config_without_secret_diagnostics(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "alerts.env"
            path.write_text("GATI_SMTP_PASSWORD=private-canary\n")
            path.chmod(0o600)
            result = subprocess.run(
                [
                    sys.executable,
                    str(pathlib.Path(__file__).with_name("ops_alerts.py")),
                    "--monitor-file",
                    str(path.with_name("unused.json")),
                    "--email-env-file",
                    str(path),
                ],
                capture_output=True,
                text=True,
                timeout=5,
            )
            self.assertEqual(result.returncode, 1)
            self.assertEqual(result.stdout, "")
            self.assertEqual(
                result.stderr, "Cannot read private alert configuration.\n"
            )

    def settings(self, mode="ssl", port=465):
        return {
            "GATI_ALERT_TO": "project@example.invalid",
            "GATI_ALERT_FROM": "sender@example.invalid",
            "GATI_SMTP_HOST": "localhost",
            "GATI_SMTP_PORT": str(port),
            "GATI_SMTP_TLS": mode,
            "GATI_SMTP_USERNAME": "synthetic-user",
            "GATI_SMTP_PASSWORD": "private-canary-$literal#value",
        }

    def test_private_env_validation_and_literal_values(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory) / "alerts.env"
            value = self.settings()
            raw = "\n".join(f"{k}='{v}'" for k, v in value.items())
            path.write_text("# comment\n" + raw)
            path.chmod(0o600)
            self.assertEqual(email_config(path), value)
            for suffix in ["\nGATI_SMTP_TLS=ssl", "\nUNKNOWN=secret"]:
                path.write_text(raw + suffix)
                with self.assertRaises(ValueError):
                    email_config(path)
            for key, bad in [
                ("GATI_SMTP_TLS", "none"),
                ("GATI_SMTP_PORT", "0"),
                ("GATI_SMTP_PORT", "65536"),
                ("GATI_SMTP_PASSWORD", ""),
                ("GATI_ALERT_TO", "a@example.invalid,b@example.invalid"),
                ("GATI_ALERT_FROM", "a@example.invalid\r\nBcc: b@example.invalid"),
                ("GATI_SMTP_HOST", "smtp.example.invalid/path"),
            ]:
                data = {**value, key: bad}
                path.write_text("\n".join(f"{k}={v}" for k, v in data.items()))
                with self.assertRaises(ValueError, msg=key):
                    email_config(path)
            path.write_text(raw)
            path.chmod(0o644)
            with self.assertRaises(ValueError):
                email_config(path)
            path.chmod(0o600)
            link = path.with_name("link")
            link.symlink_to(path)
            with self.assertRaises(OSError):
                email_config(link)
            path.write_text("#" * 4097)
            with self.assertRaises(ValueError):
                email_config(path)

    def test_starttls_failure_prevents_auth_and_email(self):
        client = Mock()
        client.starttls.side_effect = smtplib.SMTPNotSupportedError("private-canary")
        with patch("ops_email.smtplib.SMTP", return_value=client):
            dispatcher = Dispatcher(
                lambda payload: send_email(self.settings("starttls"), payload)
            )
            dispatcher.step(("cleanup_stale",), 0)
            self.assertEqual(dispatcher.step(("cleanup_stale",), 5), "delivery_failed")
            self.assertEqual(dispatcher.step(("cleanup_stale",), 64), "idle")
        client.login.assert_not_called()
        client.send_message.assert_not_called()
        client.close.assert_called_once()

    def test_rejects_sensitive_payload_before_connecting(self):
        with patch("ops_email.smtplib.SMTP_SSL") as connect:
            with self.assertRaises(ValueError):
                send_email(
                    self.settings(),
                    {"service": "gati", "state": "alert", "alerts": ["private-canary"]},
                )
            connect.assert_not_called()

    def test_local_smtp_tls_modes_retry_recovery_and_untrusted_certificate(self):
        with tempfile.TemporaryDirectory() as directory:
            cert, key = (
                pathlib.Path(directory) / "cert.pem",
                pathlib.Path(directory) / "key.pem",
            )
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
            context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
            context.load_cert_chain(cert, key)
            trust = ssl.create_default_context(cafile=str(cert))
            for mode in ("ssl", "starttls"):
                with self.subTest(mode=mode):
                    messages, commands, unsafe = [], [], []
                    responses = iter([451, 250, 250])

                    class Receiver(socketserver.BaseRequestHandler):
                        def handle(self):
                            sock = self.request
                            sock.settimeout(3)
                            secured = False
                            try:
                                if mode == "ssl":
                                    sock = context.wrap_socket(sock, server_side=True)
                                    secured = True
                                stream = sock.makefile("rb")
                                sock.sendall(b"220 localhost synthetic SMTP\r\n")
                                while True:
                                    line = stream.readline(4096)
                                    if not line:
                                        break
                                    verb = line.split(b" ", 1)[0].strip().upper()
                                    commands.append(verb)
                                    if verb in (b"EHLO", b"HELO"):
                                        commands.append(line.strip())
                                        sock.sendall(
                                            b"250-localhost\r\n250-AUTH PLAIN\r\n250 STARTTLS\r\n"
                                        )
                                    elif verb == b"STARTTLS":
                                        sock.sendall(b"220 begin TLS\r\n")
                                        stream.close()
                                        sock = context.wrap_socket(
                                            sock, server_side=True
                                        )
                                        stream = sock.makefile("rb")
                                        secured = True
                                    elif verb in (b"AUTH", b"MAIL", b"RCPT"):
                                        if not secured:
                                            unsafe.append(verb)
                                        sock.sendall(
                                            b"235 authenticated\r\n"
                                            if verb == b"AUTH"
                                            else b"250 accepted\r\n"
                                        )
                                    elif verb == b"DATA":
                                        if not secured:
                                            unsafe.append(verb)
                                        sock.sendall(b"354 send data\r\n")
                                        body = bytearray()
                                        while True:
                                            row = stream.readline(4096)
                                            if row == b".\r\n" or not row:
                                                break
                                            body.extend(row)
                                        messages.append(
                                            email.message_from_bytes(bytes(body))
                                        )
                                        sock.sendall(
                                            str(next(responses)).encode()
                                            + b" synthetic result\r\n"
                                        )
                                    else:
                                        sock.sendall(b"250 ok\r\n")
                                stream.close()
                            except (ssl.SSLError, ConnectionError):
                                pass  # Expected certificate rejection; no secret diagnostics.
                            finally:
                                sock.close()

                    server = socketserver.TCPServer(("127.0.0.1", 0), Receiver)
                    thread = threading.Thread(target=server.serve_forever, daemon=True)
                    thread.start()
                    config = self.settings(mode, server.server_address[1])
                    payload = {
                        "service": "gati",
                        "state": "alert",
                        "alerts": ["cleanup_stale"],
                    }
                    try:
                        # Actual default trust must reject the synthetic self-signed cert.
                        with self.assertRaises(ssl.SSLCertVerificationError):
                            send_email(config, payload)
                        self.assertNotIn(b"AUTH", commands)
                        with patch(
                            "ops_email.ssl.create_default_context", return_value=trust
                        ):
                            dispatcher = Dispatcher(
                                lambda data: send_email(config, data)
                            )
                            dispatcher.step(("cleanup_stale",), 0)
                            self.assertEqual(
                                dispatcher.step(("cleanup_stale",), 5),
                                "delivery_failed",
                            )
                            self.assertEqual(
                                dispatcher.step(("cleanup_stale",), 65), "delivered"
                            )
                            dispatcher.step((), 100)
                            self.assertEqual(dispatcher.step((), 125), "delivered")
                        self.assertFalse(unsafe)
                        self.assertEqual(len(messages), 3)
                        bodies = [
                            m.get_payload(decode=True).decode().replace("\r\n", "\n")
                            for m in messages
                        ]
                        self.assertEqual(
                            bodies,
                            ["GATI\nalert\ncleanup_stale\n"] * 2
                            + ["GATI\nrecovered\n"],
                        )
                        self.assertNotIn(
                            "canary", "".join(m.as_string() for m in messages)
                        )
                        self.assertTrue(
                            all(m["To"] == config["GATI_ALERT_TO"] for m in messages)
                        )
                        self.assertIn(b"ehlo gati.invalid", commands)
                    finally:
                        server.shutdown()
                        server.server_close()
                        thread.join(timeout=2)


if __name__ == "__main__":
    unittest.main()
