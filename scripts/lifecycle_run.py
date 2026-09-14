"""Real-clock, production-config lifecycle exercise in an owned disposable lab."""

import argparse
import base64
import collections
import hashlib
import json
import os
import pathlib
import signal
import subprocess
import time

from lab import Lab
from monitor import process_resources
from recovery import call, token


def arrival_cell(grid, point):
    x = int((point[0] - grid["west"]) / grid["lon_step"])
    y = int((point[1] - grid["south"]) / grid["lat_step"])
    return f"tirana-v1:{grid['size_meters']}:{x}:{y}"


def write_report(path, report):
    temporary = path.with_suffix(".tmp")
    temporary.write_text(json.dumps(report, indent=2) + "\n")
    temporary.chmod(0o600)
    temporary.replace(path)


class Exercise:
    def __init__(self, lab):
        self.lab = lab
        self.config = lab.config
        self.grid = call(lab.base, "GET", "/api/geography")[1]
        self.cell = f"tirana-v1:{self.grid['size_meters']}:55:55"
        self.people = []
        self.gatherings = {}
        self.late = set()
        self.counts = collections.Counter()

    def request(
        self, method, path, person=None, body=None, nonce=None, expected=(200,)
    ):
        status, value = call(
            self.lab.base, method, path, person["cap"] if person else None, body, nonce
        )
        self.counts[f"http_{status}"] += 1
        if status not in expected:
            # Never include credentials, responses or participant locations in errors.
            raise ValueError(f"{method} {path}: unexpected HTTP {status}")
        return value

    def create(self, minutes, role="attend", cell=None, gathering=None):
        body = {
            "cell": cell or self.cell,
            "radius_km": 3,
            "availability_minutes": minutes,
        }
        person = {
            "cap": token(),
            "role": role,
            "body": body,
            "invited": None,
            "arrival_until": 0,
            "gone": False,
            "verified": False,
        }
        value = self.request(
            "POST",
            "/api/join" if gathering else "/api/signals",
            person,
            {**body, **({"gathering_id": gathering} if gathering else {})},
        )
        person["expires"] = value["expires_at"]
        person["hash"] = hashlib.sha256(
            base64.urlsafe_b64decode(person["cap"] + "=")
        ).hexdigest()
        self.people.append(person)
        self.counts["created"] += 1
        return person

    def cohort(self):
        count = max(
            self.config["matching"]["activation_count"] + 4,
            self.config["arrivals"]["confirmation_count"] + 4,
        )
        if count + len(self.config["availability"]["choices_minutes"]) > 55:
            raise ValueError("cohort exceeds this bounded single-network scenario")
        for n in range(count):
            role = (
                "cancel"
                if n == 0
                else "decline"
                if n == 1
                else "retract"
                if n == 2
                else "attend"
            )
            self.create(self.config["availability"]["maximum_minutes"], role)
        # Below-threshold keepers exercise each offered lifetime, without arrivals.
        far = f"tirana-v1:{self.grid['size_meters']}:{self.grid['columns'] - 1}:{self.grid['rows'] - 1}"
        for minutes in self.config["availability"]["choices_minutes"]:
            self.create(minutes, "wait", far)
        self.counts["cohorts"] += 1

    def arrive(self, person, invitation, renewal=False):
        nonce = token()
        path = "/api/arrival-renewal" if renewal else "/api/arrival"
        self.request("POST", path + "-nonce", person, nonce=nonce)
        body = {"cell": arrival_cell(self.grid, invitation["intersection"]["point"])}
        value = self.request("POST", path, person, body, nonce)
        replay = self.request("POST", path, person, body, nonce)
        if replay["arrival_until"] != value["arrival_until"]:
            raise ValueError("arrival replay changed freshness")
        if renewal and value["arrival_until"] <= person["arrival_until"]:
            raise ValueError("renewal did not extend a renewable contribution")
        person["arrival_until"] = value["arrival_until"]
        person["last_nonce"] = nonce
        if person["arrival_until"] > min(person["expires"], invitation["ends_at"]):
            raise ValueError("arrival exceeded original deadlines")
        self.counts["renewals" if renewal else "arrivals"] += 1
        self.counts["arrival_retries"] += 1

    def tick(self):
        now = int(time.time() * 1000)
        for person in list(self.people):
            if person["verified"]:
                continue
            if now > person["expires"] + 30000:
                self.request("GET", "/api/signal", person, expected=(410,))
                h = person["hash"]
                keys = [
                    f"gati:{prefix}:{h}"
                    for prefix in ["s", "cap", "arrival-nonce", "push", "push-gap"]
                ]
                if self.lab.store("EXISTS", *keys) != 0:
                    raise ValueError("participant keys survived deadline")
                if (
                    self.lab.store(
                        "ZSCORE", "gati:expiry", h + "|" + person["body"]["cell"]
                    )
                    is not None
                ):
                    raise ValueError("expiry index retained expired member")
                if (
                    self.lab.store("ZSCORE", "gati:cell:" + person["body"]["cell"], h)
                    is not None
                ):
                    raise ValueError("cell index retained expired member")
                person["verified"] = True
                self.counts["verified_expiries"] += 1
                continue
            if person["gone"] or now >= person["expires"]:
                continue
            if person["role"] == "cancel":
                self.request("DELETE", "/api/signal", person, expected=(204,))
                self.request(
                    "POST", "/api/signals", person, person["body"], expected=(410,)
                )
                person["gone"] = True
                self.counts["cancellations"] += 1
                continue
            value = self.request("GET", "/api/signal", person)
            if value["expires_at"] != person["expires"]:
                raise ValueError("session deadline changed")
            invitation = value.get("invitation")
            if not invitation or person["role"] == "wait":
                continue
            gid = invitation["id"]
            frozen = (invitation["ends_at"], invitation["intersection"]["id"])
            if self.gatherings.setdefault(gid, frozen) != frozen:
                raise ValueError("gathering destination/deadline changed")
            if person["role"] == "decline":
                if person["invited"] != gid:
                    self.request("POST", "/api/decline", person, {"gathering_id": gid})
                    person["invited"] = gid
                    self.counts["declines"] += 1
                continue
            if value["state"] in ("gati", "invited"):
                value = self.request(
                    "POST", "/api/going", person, {"gathering_id": gid}
                )
                self.counts["going"] += 1
            if value["state"] == "going" and invitation["ends_at"] - now > 30000:
                self.arrive(person, invitation)
            elif value["state"] == "here":
                if person["role"] == "retract":
                    self.request("DELETE", "/api/arrival", person)
                    self.request(
                        "POST",
                        "/api/arrival",
                        person,
                        {
                            "cell": arrival_cell(
                                self.grid, invitation["intersection"]["point"]
                            )
                        },
                        person["last_nonce"],
                        expected=(410,),
                    )
                    person["role"] = "wait"
                    self.counts["arrival_retractions"] += 1
                    continue
                person["arrival_until"] = value["arrival_until"]
                remaining = value["arrival_until"] - now
                window = self.config["arrivals"]["renewal_window_seconds"] * 1000
                if 10000 < remaining < window - 5000 and value["arrival_until"] < min(
                    person["expires"], invitation["ends_at"]
                ):
                    self.arrive(person, invitation, True)
            if invitation["state"] == "jemi_ketu" and gid not in self.late:
                cutoff = (
                    self.config["matching"]["late_join_min_remaining_minutes"] * 60000
                )
                if invitation["ends_at"] - now > cutoff + 30000:
                    self.create(30, gathering=gid)
                    self.late.add(gid)
                    self.counts["joins_after_presence"] += 1


def run(output, hours=8, smoke=False):
    if not 8 <= hours <= 24:
        raise ValueError("full lifecycle runs require 8..24 hours")
    output = pathlib.Path(output).resolve()
    report = {
        "version": 1,
        "synthetic": True,
        "real_time": True,
        "smoke_only": smoke,
        "status": "starting",
        "pid": os.getpid(),
        "counts": {},
        "samples": [],
        "scope": "owned local API, production config; no remote target or OS/device proof",
    }
    revision = subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
    report["source_revision"] = revision
    report["driver_sha256"] = hashlib.sha256(
        pathlib.Path(__file__).read_bytes()
    ).hexdigest()
    with Lab(output, source_revision=revision) as lab:
        scenario = Exercise(lab)
        maximum = lab.config["availability"]["maximum_minutes"] * 60
        duration = 90 if smoke else hours * 3600
        start = time.time()
        monotonic = time.monotonic()
        next_cohort = start
        last_sample = 0
        report["started_at"] = int(start)
        report["expected_finish_at"] = int(start + duration)
        path = output / "lifecycle.json"
        try:
            while True:
                now = time.time()
                elapsed = now - start
                if abs(elapsed - (time.monotonic() - monotonic)) > 30:
                    raise ValueError(
                        "host suspension or wall-clock jump invalidated run"
                    )
                if now >= next_cohort and (
                    elapsed == 0
                    or not scenario.people
                    or elapsed < duration - maximum - 60
                ):
                    scenario.cohort()
                    next_cohort += maximum
                scenario.tick()
                if now - last_sample >= 60:
                    public = scenario.request(
                        "GET", "/api/activity/latest", expected=(200, 204)
                    )
                    if public is not None:
                        if public["expires_at"] <= now * 1000:
                            raise ValueError("expired public snapshot served")
                        scenario.counts["public_snapshot_observations"] += 1
                    report["samples"].append(
                        {
                            "elapsed_seconds": int(elapsed),
                            "api_rss_bytes": process_resources(
                                lab.apis[0]["process"].pid
                            )["rss_bytes"],
                            "store_rss_bytes": process_resources(lab.store_pid)[
                                "rss_bytes"
                            ],
                            "expiry_members": lab.store("ZCARD", "gati:expiry"),
                            "gathering_members": lab.store("ZCARD", "gati:gatherings"),
                        }
                    )
                    last_sample = now
                report.update(
                    status="running",
                    updated_at=int(now),
                    elapsed_seconds=int(elapsed),
                    counts=dict(scenario.counts),
                )
                write_report(path, report)
                if elapsed >= duration:
                    break
                time.sleep(min(15, duration - elapsed))
            if not smoke:
                if not all(p["verified"] for p in scenario.people):
                    raise ValueError("not all participant lifetimes were verified")
                for key in ["gati:expiry", "gati:gatherings", "gati:push-due"]:
                    if lab.store("ZCARD", key) != 0:
                        raise ValueError("index did not return to baseline")
                if scenario.counts["renewals"] == 0:
                    raise ValueError("presence renewal was not exercised")
                if scenario.counts["public_snapshot_observations"] == 0:
                    raise ValueError("no real delayed public snapshot observed")
                # A conservative drift gate, separate from formal retention proof.
                first, last = report["samples"][0], report["samples"][-1]
                if (
                    last["api_rss_bytes"] - first["api_rss_bytes"] > 256 * 1024**2
                    or last["store_rss_bytes"] - first["store_rss_bytes"] > 64 * 1024**2
                ):
                    raise ValueError("memory drift exceeded scenario budget")
            for key in [
                "arrivals",
                "arrival_retries",
                "arrival_retractions",
                "joins_after_presence",
                "declines",
                "cancellations",
            ]:
                if scenario.counts[key] == 0:
                    raise ValueError("required journey was not exercised: " + key)
            report["status"] = "smoke_passed" if smoke else "passed"
        except BaseException as error:
            report["status"] = (
                "interrupted" if isinstance(error, KeyboardInterrupt) else "failed"
            )
            report["failure_type"] = type(error).__name__
            raise
        finally:
            report["counts"] = dict(scenario.counts)
            report["updated_at"] = int(time.time())
            write_report(path, report)
    print("Lifecycle " + report["status"] + "; see private local report.")


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", required=True)
    parser.add_argument("--hours", type=int, default=8)
    parser.add_argument("--smoke", action="store_true")
    args = parser.parse_args()
    signal.signal(signal.SIGTERM, lambda *_: (_ for _ in ()).throw(KeyboardInterrupt()))
    run(args.output, args.hours, args.smoke)
