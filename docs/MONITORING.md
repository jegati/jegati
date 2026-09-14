# Local service monitoring

Optional API flag: `-monitor-socket /private/directory/gati.sock`. The directory
must have no group/other permissions; the socket is mode 0600. No TCP metrics port
or participant monitoring route exists. Inspect locally with:

```sh
curl --unix-socket /private/directory/gati.sock http://localhost/metrics
python3 scripts/monitor.py --socket /private/directory/gati.sock --output .runtime/monitor.json --pid API_PID
```

`API_PID` denotes the process to measure; omit it to collect only socket summaries.
The collector defaults to five-second samples and one-hour retention; `--interval`
and `--retention` accept bounded seconds (see CLI validation). It overwrites a bounded
JSON file, never appends an event log. `expires_at` marks file freshness; an expired
file means monitoring is unavailable. Physical old files can remain after a crash.

Protocol v2 reports only a completed 60-second request window, count buckets suppressed
below 20, and a p95 latency upper bound. `p95_upper_ms: -1` means suppressed;
`0` means above the largest 10,000 ms bucket. No individual request metadata is
retained. Worker states describe loop completion, not delivery guarantees or a
measurement of every gathering's matching latency. `last_success_seconds: -1`
means no successful completion has been observed. Worker ages are rounded to five
seconds. The default 30-second stale alert is operational policy in the collector,
not an override of GATI's matching deadlines.

Alerts: unavailable monitor, worker error/staleness, a released server-failure
bucket or high latency. Below-threshold failures are suppressed, so absence of a
failure bucket is not proof of zero errors. The lab's exact synthetic client report
separately records response codes/timeouts. Losing monitoring must not block the app.

Run `make test-monitor` for schema, suppression/expiry, concurrency, socket, alert
and collector-failure tests. `make load` creates an isolated monitored lab and writes
synthetic reports under `reports/local/`; it never points load at the normal local
app. See [decision 0010](decisions/0010-operational-monitoring.md). Deployment wiring,
real-device feedback and operational retention review remain separate work.

The monitored lab also samples the actual Valkey process, not the container's init
supervisor. Load reports include an offline CPU/RSS chart and client-side accepted
versus rejected latency/counts. Generate another view with
`python3 scripts/load-report.py reports/local/RUN_DIRECTORY`. These synthetic
archives are explicitly historical. CPU percentages use one core as 100%; sampled
peaks are not instantaneous maxima. No generic database statistics, per-cell labels
or participant queues are exposed by the production monitor.

Local response guide: `monitor_unavailable` means inspect socket/process/collector
health; worker error/staleness means check private store reachability and worker
liveness; sustained released failure/latency buckets mean inspect capacity and
admission behavior. Do not enable individual access logging to diagnose these.
Recovery tests inject a paused API owner and paused/restarted store, then verify
alerts and recovery. A stopped collector's file expires; dashboards must honor that
freshness deadline. Push loop completion alone cannot alert on provider delivery
failure because individual send outcomes are deliberately not exposed here; this
remains a monitoring gap, not a passing claim.

Protocol v2 adds each worker's last duration upper bound (`-1` unavailable, `0`
above 10 seconds) and `deadline_lag_seconds` rounded down to five seconds. Both
observations expire after 120 seconds; errors clear unavailable lag observations.
Matcher lag measures the oldest due reservation before processing, not the time
from each person's willingness to an invitation. Cleanup lag measures the oldest
overdue expiry-index entry before cleanup, independently of logical key expiry.
Publisher lag measures missing expected public releases, with a bounded initial
publication warmup; it does not mistake the deliberate privacy delay for an outage.
No identifiers, cells or queue members enter monitoring. The collector accepts
archived v1 and current v2 and alerts on deadline lag of 30 seconds or more.

`make load MIXED=1` additionally overlaps ordinary polling, public origin reads
and up to 1,000 real going confirmations sampled from private invitations. It
requires at least one invitation and records each HTTP phase separately. This
does not test arrivals, all-population admission, synchronized expiry or a CDN.

Production Compose places the socket in an owner-only container tmpfs. Use
`python3 scripts/monitor.py --container gati-production-api-1 --socket /run/ops/gati.sock --output /PRIVATE/monitor-api.json`
to read the binary's bounded monitor-snapshot command through docker exec. No TCP
port or writable host mount enters the application container. Use private RAM-backed
output and honor expires_at. API-only replicas expose the fixed view worker; monitor
each expected process. The collector records local alerts but does not send messages.
Real tunnel/certificate/provider checks and an operator-selected alert destination
remain deployment tasks. See [production wiring](DEPLOYMENT.md).


## Optional outbound operational alerts

scripts/ops_alerts.py reads the collector's bounded local file and sends only
service=gati, state=alert/recovered and fixed alert labels to a configured HTTPS
webhook. It never forwards monitoring JSON, counts, locations, capabilities, IPs,
request bodies, error text or resource measurements. The endpoint receives normal
network/provider metadata. No endpoint is selected or contacted automatically.

Two consistent samples debounce changes; successful sends are at least 60 seconds
apart. Persistent problems remind at most every 30 minutes. Failed sends back off
from 60 to 300 seconds; only the latest state is kept in memory, without a durable
queue. Recovery is sent only after a delivered alert. Restart loses delivery state
and may repeat a continuing alert. Delivery has a five-second timeout, verified TLS
and no redirect following or environment proxy. Sender failure cannot block GATI.

Privately create /etc/gati/alerts.json, owned by the sender user, mode 0600, with
exactly url (HTTPS endpoint) and authorization (header value, or an empty string).
Do not paste either value in chat, Git, shell arguments or environment variables.
This is a generic JSON receiver contract, not a built-in Slack/Telegram/SMTP client.
Review the chosen receiver/provider and retention before enabling it. Keep endpoint
URL/query secrets private even if a receiver uses no Authorization header.

Prepared service files are deploy/gati-monitor.service and deploy/gati-alerts.service.
They assume the current immutable release is linked at /opt/gati/current and the
combined process container is gati-production-api-1. The collector writes to a
0700 /run/gati-monitor directory with 0600 files; sender state remains in RAM. The
host root/Docker operator is trusted, as with existing release administration. A
separate worker/replica topology needs additional collectors, not a silent reuse
of this single-process template. Review/install services on the actual host; they
are not installed on the workstation by this change.

Run make test-ops-alerts for a local TLS receiver exercising failure/retry/recovery,
redirection rejection, fixed payloads and credentials-file permissions. Tests make
no external requests and use an ephemeral certificate. OpenSSL is the test-only
certificate prerequisite. The notification destination and actual delivery remain
pending operator selection/private credentials. Missing/stale collector files alert
as monitor_unavailable; the sender's own process health also needs host supervision.
A complete VPS/network outage cannot be reported by a process on that VPS: configure
an independently observed availability check when hosting and alert channels exist.
