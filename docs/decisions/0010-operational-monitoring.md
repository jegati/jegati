# 0010: Bounded local operational monitoring

2026-09-13. The user authorized automated predeployment testing and monitoring,
including monitored scale runs. This does not authorize public deployment.

The API optionally exposes operational summaries on an owner-only Unix socket
(`-monitor-socket`) whose parent directory must also be private. No public route,
TCP metrics listener, tracking SDK or individual event log is added. The flag is
listener wiring, not a hidden matching/participant behavior override. Schema 8
functional configuration remains unchanged. Monitoring protocol version 1 publishes
its fixed privacy policy alongside summaries; changing these bounds requires review.

Only the completed fixed 60-second request window is readable. Two windows exist
in bounded memory, with 20-sample suppression and coarse request/failure buckets.
Latency is a histogram-derived p95 upper bound, available only above the sample
floor. There are no route, IP, token, cell, gathering, endpoint or participant labels.
Worker health records only four fixed tasks, completion/error state and ages rounded
down to five seconds. Errors themselves are never retained. These summaries are
operational signals, not an anonymity proof or protected public activity releases.

The first-party collector reads this socket and optionally Linux process CPU/RSS.
It validates the response schema and fixed labels before retaining bounded samples;
default retention is one hour with five-second sampling. A snapshot file carries
an expiry deadline for freshness, and bounded history is overwritten atomically.
A stopped/crashed collector can leave old bytes on disk: consumers must treat an
expired file as unavailable, not a live health report. No forensic erasure is claimed.
Public activity counts continue to use their own delayed/suppressed publisher.

Monitoring failures cannot require participant state changes. Slow/unavailable
collectors do not sit in the application request path; instrumentation keeps only
bounded counters and task times. Socket access grants operational visibility and
belongs to the operator, not normal users. Correlation through timing/resource use
remains a residual risk. No per-user funnel or real-location usability recording.

An owned lab creates its own loopback API/store, uses normal production binaries
and thresholds, and starts the same collector. Exact client-side load statistics
are synthetic report data, never a new production metrics surface. Fixed host
ports within the owned lab make store-restart tests meaningful; the lab refuses to
attach load tools to an arbitrary target. Real network/edge protection, host rollout
and remote artifact verification remain later deployment work.
