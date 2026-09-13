#!/usr/bin/env python3
"""Two clean source exports, isolated dependency installs and byte comparisons."""
import argparse,hashlib,io,json,os,pathlib,subprocess,tarfile,tempfile
ROOT=pathlib.Path(__file__).resolve().parent.parent
p=argparse.ArgumentParser();p.add_argument('--output',required=True);a=p.parse_args();out=pathlib.Path(a.output).resolve();out.mkdir(parents=True,exist_ok=True)
def run(*args,cwd=ROOT,**kw):return subprocess.check_output(args,cwd=cwd,**kw)
if run('git','status','--porcelain','--untracked-files=no'):raise SystemExit('commit tracked work before the clean-build rehearsal')
revision=run('git','rev-parse','HEAD',text=True).strip();archive=run('git','archive',revision);manifests=[]
with tempfile.TemporaryDirectory(prefix='gati-reproduce-') as d:
 for n in range(2):
  checkout=pathlib.Path(d)/str(n);checkout.mkdir()
  with tarfile.open(fileobj=io.BytesIO(archive)) as tar:tar.extractall(checkout,filter='data')
  with (out/f'build-{n}.log').open('wb') as log:
   env={**os.environ,'CGO_ENABLED':'0','GOCACHE':str(checkout/'.go-build-cache')}
   for args in [('go','mod','verify'),('go','build','-trimpath','-buildvcs=false','-o','gati','./cmd/gati'),('npm','--prefix','web','ci','--ignore-scripts','--no-audit','--no-fund'),('npm','--prefix','web','run','build')]:
    subprocess.run(args,cwd=checkout,env=env,stdout=log,stderr=log,check=True)
  files=[checkout/'gati',*(checkout/'web/dist').rglob('*')]
  manifests.append({str(f.relative_to(checkout)):hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted(files) if f.is_file()})
  if n==0:
   (out/'go-modules.jsonl').write_bytes(run('go','list','-m','-json','all',cwd=checkout))
   lock=json.loads((checkout/'web/package-lock.json').read_text())
   (out/'npm-inventory.json').write_text(json.dumps({k:{f:v[f] for f in ['version','integrity','license','dev'] if f in v} for k,v in lock['packages'].items()},indent=2)+'\n')
result={'source_revision':revision,'equal':manifests[0]==manifests[1],'artifacts':manifests,'go':run('go','version',text=True).strip(),'node':run('node','--version',text=True).strip(),'scope':'two clean exports and independent compilation caches on this same host/toolchain; no independent reviewer, separate OS or remote attestation'}
(out/'reproduction.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items() if k!='artifacts'},indent=2));raise SystemExit(0 if result['equal'] else 1)
