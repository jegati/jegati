#!/usr/bin/env python3
"""Run isolated experiments with a frozen settings snapshot and reviewable output."""
import argparse, concurrent.futures, html, json, pathlib, subprocess, time
parser=argparse.ArgumentParser()
parser.add_argument('--output',default='reports/local/tirana-suite')
parser.add_argument('--config',default='config/gati.yaml')
parser.add_argument('--seeds',type=int,nargs='+',default=[42,43,44])
parser.add_argument('--jobs',type=int,choices=[1,2,3],default=3)
parser.add_argument('--resume',action='store_true',help='recheck completed runs and archive/retry failed runs using the saved config')
args=parser.parse_args()
if len(args.seeds)>10 or len(set(args.seeds))!=len(args.seeds) or any(s<0 or s>2**64-1 for s in args.seeds): parser.error('use at most ten distinct unsigned 64-bit seeds')
root=pathlib.Path(args.output);root.mkdir(parents=True,exist_ok=True)
previous={}
attempt=str(time.time_ns())
if args.resume:
    if not (root/'requested-config.yaml').exists(): parser.error('resume requires an existing suite config snapshot')
    if (root/'summary.json').exists():
        previous={(r['scenario'],r['seed']):r for r in json.loads((root/'summary.json').read_text())}
        (root/f'summary-before-resume-{attempt}.json').write_bytes((root/'summary.json').read_bytes())
else:
    if (root/'requested-config.yaml').exists(): parser.error('choose a new output directory, or use --resume to preserve and retry failures')
    (root/'requested-config.yaml').write_bytes(pathlib.Path(args.config).read_bytes())
cases=['tirana-population','tirana-sparse','tirana-dense','tirana-low-followthrough','tirana-churn','tirana-boundaries']
def run(case,seed):
    out=root/f'{case}-{seed}'
    reuse=args.resume and (out/'report.json').exists() and json.loads((out/'report.json').read_text()).get('status')=='completed'
    if out.exists() and not reuse:
        archive=root/'failed-attempts';archive.mkdir(exist_ok=True)
        out.rename(archive/f'{case}-{seed}-{attempt}')
    out.mkdir(exist_ok=True)
    start=time.monotonic()
    with (out/'execution.log').open('a' if reuse else 'w') as log:
        if reuse:
            log.write('\nRechecking completed evidence; simulation was not rerun.\n');log.flush()
            result=subprocess.CompletedProcess([],0)
        else:
            result=subprocess.run(['make','simulate-population',f'SCENARIO={case}',f'SEED={seed}',f'OUTPUT={out}',f'SIM_CONFIG={root / "requested-config.yaml"}'],stdout=log,stderr=subprocess.STDOUT)
        if result.returncode==0:
            # Make supplies the pinned Node PATH for checks, too.
            result=subprocess.run(['make','check-population',f'REPORT={out / "report.json"}'],stdout=log,stderr=subprocess.STDOUT)
    row={'scenario':case,'seed':seed,'status':'passed' if result.returncode==0 else 'failed','elapsed_seconds':round(time.monotonic()-start,2),'report':str(out.relative_to(root) / 'index.html')}
    if reuse:
        row['rechecked_completed_run']=True
        row['elapsed_seconds']=previous.get((case,seed),{}).get('elapsed_seconds',row['elapsed_seconds'])
    if (out/'report.json').exists():
        report=json.loads((out/'report.json').read_text());row.update(counts=report['counts'],config_sha256=report['config_sha256'],input_sha256=report['input_sha256'],wall_seconds=report['wall_seconds'],source_revision=report['source_revision'])
    print(f'{case} seed {seed}: {row["status"]} ({row["elapsed_seconds"]} s)',flush=True)
    return row
rows=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=args.jobs) as pool:
    futures=[pool.submit(run,c,s) for c in cases for s in args.seeds]
    for future in concurrent.futures.as_completed(futures):
        rows.append(future.result());(root/'summary.json').write_text(json.dumps(sorted(rows,key=lambda r:(r['scenario'],r['seed'])),indent=2)+'\n')
rows.sort(key=lambda r:(r['scenario'],r['seed']))
page=['<!doctype html><html lang="sq"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>GATI — Provat e Tiranës</title><style>body{font:16px system-ui;background:#fff8f3;color:#352c31;margin:24px}table{border-collapse:collapse}td,th{padding:10px;border-bottom:1px solid #ccc;text-align:left}a{color:#963754}</style><h1>🦩 SIMULIM — Provat e Tiranës</h1><p>Vetëm aktorë sintetikë. Afate virtuale; kufij abuzimi në kohë reale. Mbërritjet më poshtë numërojnë aktorë me të paktën një konfirmim.</p><table><tr><th>Skenari</th><th>Fara</th><th>Pranuar</th><th>Me ftesë</th><th>Mbërritur</th><th>Takime</th><th>JEMI KËTU</th><th>Sekonda reale</th><th>Kontrolli</th></tr>']
for row in rows:
    c=row.get('counts',{});page.append('<tr><td><a href="'+html.escape(row['report'])+'">'+html.escape(row['scenario'])+'</a></td>'+''.join('<td>'+html.escape(str(v))+'</td>' for v in [row['seed'],c.get('credentials_accepted',0),c.get('credentials_invited',0),c.get('credentials_arrived',0),c.get('gatherings_observed',0),c.get('gatherings_confirmed_observed',0),round(row.get('wall_seconds',0),1),'Kaloi' if row['status']=='passed' else 'Dështoi'])+'</tr>')
page.append('</table><p>Rezultatet nuk provojnë përdorshmërinë nga njerëz realë, prani fizike apo kapacitet për 100 mijë përdorues.</p></html>');(root/'index.html').write_text(''.join(page))
if any(r['status']!='passed' for r in rows):raise SystemExit('Some simulation checks failed; see per-run logs and summary.json')
