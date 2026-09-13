#!/usr/bin/env python3
"""Compare repeated Go benchmarks; medians/ranges are not confidence intervals."""
import argparse, json, pathlib, re, statistics
from xml.sax.saxutils import escape

def read(directory):
    result={}
    for file in ('store.txt','geography.txt'):
        text=(directory/file).read_text()
        if '\nPASS\n' not in text:raise ValueError('benchmark did not pass: '+file)
        for line in text.splitlines():
            m=re.fullmatch(r'(Benchmark\S+)-\d+\s+(\d+)\s+([\d.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op',line.strip())
            if m:result.setdefault(m[1],[]).append({'ns':float(m[3]),'bytes':int(m[4]),'allocations':int(m[5]),'iterations':int(m[2])})
    if len(result) not in (6,8):raise ValueError('incomplete scaling benchmark suite')
    return result

def compare(before,after):
    a,b=read(before),read(after)
    if a.keys()!=b.keys():raise ValueError('benchmark suites differ')
    if (before/'toolchain.txt').read_text()!=(after/'toolchain.txt').read_text():raise ValueError('toolchains differ')
    rows=[]
    for name in a:
        left,right=a[name],b[name]
        if len(left)<3 or len(right)<3:raise ValueError('need at least three samples')
        median=lambda samples,key:statistics.median(s[key] for s in samples)
        old,new=median(left,'ns'),median(right,'ns')
        rows.append({'benchmark':name,'before_median_ms':old/1e6,'after_median_ms':new/1e6,'speedup':old/new,'before_range_ms':[min(s['ns'] for s in left)/1e6,max(s['ns'] for s in left)/1e6],'after_range_ms':[min(s['ns'] for s in right)/1e6,max(s['ns'] for s in right)/1e6],'before_bytes_per_op':median(left,'bytes'),'after_bytes_per_op':median(right,'bytes'),'before_samples':len(left),'after_samples':len(right)})
    return {'scope':'Isolated synthetic component benchmarks on one host; not HTTP, whole-app throughput or 1M-user capacity. Repeated-sample ranges are not confidence intervals. B/op measures allocation traffic, not peak/live RAM.','rows':rows}

def main():
    p=argparse.ArgumentParser();p.add_argument('before',type=pathlib.Path);p.add_argument('after',type=pathlib.Path);p.add_argument('--output',required=True,type=pathlib.Path);a=p.parse_args()
    report=compare(a.before,a.after);a.output.mkdir(parents=True,exist_ok=True)
    (a.output/'comparison.json').write_text(json.dumps(report,indent=2)+'\n')
    # Each row has its own relative scale; the exact milliseconds accompany it.
    height=172+78*len(report['rows'])
    parts=[f'<svg xmlns="http://www.w3.org/2000/svg" width="1000" height="{height}" viewBox="0 0 1000 {height}" role="img" aria-label="GATI component benchmark comparison">',f'<rect width="1000" height="{height}" fill="#fff"/>','<g font-family="sans-serif" fill="#183b36">','<text x="30" y="38" font-size="22">GATI: measured component speedups</text>','<text x="30" y="65" font-size="13">Gray = before; green = after. Each pair is scaled independently. Median time per operation.</text>']
    for n,r in enumerate(report['rows']):
        y=110+n*78;old,new=r['before_median_ms'],r['after_median_ms'];scale=380/max(old,new)
        label=escape(r['benchmark'].removeprefix('Benchmark'))
        parts += [f'<text x="30" y="{y}" font-size="15">{label}</text>',f'<rect x="360" y="{y-16}" width="{old*scale:.2f}" height="16" fill="#a8b5b2"/>',f'<rect x="360" y="{y+6}" width="{new*scale:.2f}" height="16" fill="#27806d"/>',f'<text x="755" y="{y}" font-size="13">{old:.4g} → {new:.4g} ms</text>',f'<text x="755" y="{y+22}" font-size="13">{r["speedup"]:.2f}×</text>']
    parts+=[f'<text x="30" y="{height-30}" font-size="13">Synthetic component results; these ratios are not whole-application or million-user capacity claims.</text>','</g></svg>']
    (a.output/'comparison.svg').write_text('\n'.join(parts)+'\n')
    print(json.dumps(report,indent=2))

if __name__=='__main__':main()
