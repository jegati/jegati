#!/usr/bin/env python3
"""Evaluate synthetic client evidence; completion alone is not a capacity pass."""
import argparse,json,pathlib

def evaluate(report):
    failures=[]
    if report.get('synthetic') is not True or report.get('real_time') is not True:failures.append('not a real-time synthetic report')
    phases=report.get('phases',[])
    if not phases:failures.append('no measured phases')
    for phase in phases:
        name=phase['name'];codes=phase['statuses'];accepted=sum(n for c,n in codes.items() if c.startswith('2'))
        if phase['generator_dropped']:failures.append(name+': generator dropped scheduled requests')
        if sum(codes.values())!=phase['attempted']:failures.append(name+': inconsistent response accounting')
        if any(c not in ('200','204','429') for c in codes):failures.append(name+': unexpected responses/timeouts')
        if not accepted:failures.append(name+': no accepted traffic')
        # Legacy reports with no split histogram qualify only if every response succeeded.
        latency=phase.get('accepted_p95_upper_ms',phase['p95_upper_ms'] if accepted==phase['attempted'] else None)
        if latency is None:failures.append(name+': accepted latency not separately measured')
        elif not 0<latency<300:failures.append(name+': accepted p95 <300ms not established')
        if name.startswith('status-') or name in ('post-burst-recovery','cancellation','mixed-status','mixed-going','invitation-sample'):
            if accepted!=phase['attempted']:failures.append(name+': ordinary traffic rejected')
        if name.startswith('create-') and accepted/phase['seconds']<290:failures.append(name+': 300/s target not sustained within scheduling tolerance')
    stages=report.get('stages',[])
    if not stages:failures.append('no population stages')
    for stage in stages:
        if stage['accepted_responses']!=stage['attempted_population']:failures.append('population admission below requested stage')
    return {'client_gate_passed':not failures,'failures':failures,'maximum_accepted':max([s['accepted_responses'] for s in stages],default=0),'scope':'local synthetic client throughput/latency only; matching lag, expiry lag, reference host and cached edge traffic are separate gates'}

if __name__=='__main__':
    p=argparse.ArgumentParser();p.add_argument('report');a=p.parse_args();path=pathlib.Path(a.report)
    result=evaluate(json.loads(path.read_text()));path.with_name('client-gate.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2));raise SystemExit(0 if result['client_gate_passed'] else 1)
