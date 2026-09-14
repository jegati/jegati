"""Optional operator email: private file configuration, TLS, fixed health labels."""

import re
import smtplib
import ssl
from email.message import EmailMessage

from ops_alerts import LABELS, read_private_config

KEYS = {
    "GATI_ALERT_TO",
    "GATI_ALERT_FROM",
    "GATI_SMTP_HOST",
    "GATI_SMTP_PORT",
    "GATI_SMTP_TLS",
    "GATI_SMTP_USERNAME",
    "GATI_SMTP_PASSWORD",
}


def email_config(path):
    values = {}
    for line in read_private_config(path).splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        key, separator, value = line.partition("=")
        key, value = key.strip(), value.strip()
        if not separator or key not in KEYS or key in values:
            raise ValueError("invalid email setting")
        if value.startswith(("'", '"')):
            if len(value) < 2 or value[-1] != value[0]:
                raise ValueError("invalid quoted email setting")
            value = value[1:-1]
        if (
            not value
            or len(value) > 1024
            or any(ord(c) < 32 or ord(c) == 127 for c in value)
        ):
            raise ValueError("invalid email setting value")
        values[key] = value
    if set(values) != KEYS:
        raise ValueError("missing email settings")
    # Deliberately support one bare ASCII mailbox per field, not arbitrary headers.
    for key in ["GATI_ALERT_TO", "GATI_ALERT_FROM"]:
        if not re.fullmatch(
            r"[A-Za-z0-9.!#$%&'*+/=?^_`{|}~-]+@[A-Za-z0-9](?:[A-Za-z0-9.-]*[A-Za-z0-9])?",
            values[key],
        ):
            raise ValueError("one bare email address required")
    if not re.fullmatch(
        r"[A-Za-z0-9](?:[A-Za-z0-9.-]{0,251}[A-Za-z0-9])?", values["GATI_SMTP_HOST"]
    ):
        raise ValueError("invalid SMTP hostname")
    if (
        not values["GATI_SMTP_PORT"].isascii()
        or not values["GATI_SMTP_PORT"].isdigit()
        or not 1 <= int(values["GATI_SMTP_PORT"]) <= 65535
    ):
        raise ValueError("invalid SMTP port")
    if values["GATI_SMTP_TLS"] not in ("ssl", "starttls"):
        raise ValueError("verified SMTP TLS required")
    return values


def send_email(config, payload):
    if (
        set(payload) != {"service", "state", "alerts"}
        or payload["service"] != "gati"
        or payload["state"] not in ("alert", "recovered")
        or not isinstance(payload["alerts"], list)
        or len(payload["alerts"]) > len(LABELS)
        or not all(isinstance(v, str) and v in LABELS for v in payload["alerts"])
    ):
        raise ValueError("invalid operational email payload")
    message = EmailMessage()
    message["From"] = config["GATI_ALERT_FROM"]
    message["To"] = config["GATI_ALERT_TO"]
    message["Subject"] = "GATI: " + (
        "njoftim teknik" if payload["state"] == "alert" else "shërbimi u rikthye"
    )
    message.set_content(
        "GATI\n" + payload["state"] + "\n" + "\n".join(payload["alerts"])
    )
    context = ssl.create_default_context()
    args = (config["GATI_SMTP_HOST"], int(config["GATI_SMTP_PORT"]))
    options = {"timeout": 5, "local_hostname": "gati.invalid"}
    if config["GATI_SMTP_TLS"] == "ssl":
        client = smtplib.SMTP_SSL(*args, context=context, **options)
    else:
        client = smtplib.SMTP(*args, **options)
    try:
        if config["GATI_SMTP_TLS"] == "starttls":
            client.starttls(context=context)  # Never authenticate or send before TLS.
        client.login(config["GATI_SMTP_USERNAME"], config["GATI_SMTP_PASSWORD"])
        client.send_message(message)
    finally:
        client.close()  # No response/debug logging or potentially blocking QUIT.
