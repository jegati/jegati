#!/usr/bin/env python3
"""Standalone, offline charts from owned synthetic load reports; no external assets."""
import argparse,html,json,pathlib

def chart(rows,key,title,cpu=False):
    series=[(r['at'],r[key]['rss_bytes']/1048576) for r in rows if key in r]
    if cpu:
        samples=[(r['at'],r[key]['cpu_seconds']) for r in rows if key in r]
        series=[(b[0],max(0,(b[1]-a[1])*100/(b[0]-a[0]))) for a,b in zip(samples,samples[1:]) if b[0]>a[0]]
    if not series:return ''
    low=min(t for t,_ in series);high=max(t for t,_ in series);peak=max(v for _,v in series)
    points=' '.join(f'{40+720*(t-low)/max(1,high-low):.1f},{190-160*v/max(1,peak):.1f}' for t,v in series)
    unit='% of one core' if cpu else 'MiB'
    return f'<figure><figcaption>{title}: peak {peak:.1f} {unit}</figcaption><svg viewBox="0 0 800 220" role="img" aria-label="{title} over elapsed time"><path d="M40 20V190H760" fill="none" stroke="#aaa"/><polyline points="{points}" fill="none" stroke="#963754" stroke-width="2"/><text x="40" y="213">0 s</text><text x="700" y="213">{high-low} s</text></svg></figure>'

p=argparse.ArgumentParser();p.add_argument('directory');a=p.parse_args();d=pathlib.Path(a.directory)
r=json.loads((d/'load.json').read_text());m=json.loads((d/'monitor-0.json').read_text());manifest=json.loads((d/'manifest.json').read_text())
if r.get('synthetic') is not True:raise SystemExit('only synthetic lab reports')
rows=m['samples'];tr=[]
for phase in r['phases']:
    codes=phase['statuses'];accepted=sum(n for c,n in codes.items() if c.startswith('2'))
    latency=phase.get('accepted_p95_upper_ms',phase['p95_upper_ms'] if accepted==phase['attempted'] else 'unmeasured')
    tr.append('<tr>'+''.join(f'<td>{html.escape(str(v))}</td>' for v in [phase['name'],phase['attempted'],accepted,codes.get('429',0),sum(n for c,n in codes.items() if c not in ('200','204','429')),phase['generator_dropped'],latency])+ '</tr>')
alerts=sorted({v for row in rows for v in row['alerts']})
text='<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>GATI synthetic load evidence</title><style>body{font:16px system-ui;max-width:1050px;margin:2rem auto;padding:1rem;color:#352c31;background:#fff8f3}table{border-collapse:collapse;width:100%}td,th{text-align:left;padding:.6rem;border-bottom:1px solid #baa5ae}svg{width:100%;max-height:260px}pre{white-space:pre-wrap;overflow-wrap:anywhere}</style><h1>GATI — synthetic load evidence</h1>'
text+='<p>Archived local benchmark with monitoring enabled. Same computer runs API, Valkey and clients. These charts are historical; they are not a live availability dashboard. No participant telemetry or external assets.</p>'
text+=f'<p>Distribution: {html.escape(r["distribution"])}. Monitoring alerts observed: {html.escape(", ".join(alerts) or "none")}.</p>'
text+='<table><tr><th>Phase</th><th>Attempts</th><th>Accepted</th><th>429</th><th>Other/errors</th><th>Dropped</th><th>Accepted p95 ≤ ms</th></tr>'+''.join(tr)+'</table>'
text+=chart(rows,'resources','API resident memory')+chart(rows,'store_resources','Valkey resident memory')
text+=chart(rows,'resources','API CPU',True)+chart(rows,'store_resources','Valkey CPU',True)
text+='<p>Resources are sampled every five seconds; peaks between samples may be missed. Worker health indicates completed loops, not per-gathering matching lag. No cached-edge, real-device or remote reference-host claim.</p><details><summary>Build and machine evidence</summary><pre>'+html.escape(json.dumps(manifest,indent=2))+'</pre></details></html>'
(d/'index.html').write_text(text);print(d/'index.html')
