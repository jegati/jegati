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

Protocol v1 reports only a completed 60-second window, count buckets suppressed
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
