#!/usr/bin/env python3
"""Build locked Caddy/cloudflared binaries from verified source, without containers."""
import argparse,json,os,pathlib,shutil,subprocess,tempfile
ROOT=pathlib.Path(__file__).resolve().parent.parent

def build(out,test=False):
 if out.exists():raise ValueError('runtime output must be a new directory')
 out.mkdir(parents=True)
 env={**os.environ,'CGO_ENABLED':'0'}
 def run(*args,cwd=ROOT):subprocess.run(args,cwd=cwd,env=env,check=True)
 lock=json.loads((ROOT/'deploy/tunnel/source.json').read_text())
 info=json.loads(subprocess.check_output(['go','mod','download','-json',lock['Path']+'@'+lock['Version']],cwd=ROOT,env=env))
 if any(info.get(k)!=lock[k] for k in ['Path','Version','Sum','GoModSum']):raise ValueError('upstream source checksum differs')
 with tempfile.TemporaryDirectory(prefix='gati-runtime-source-') as temp:
  source=pathlib.Path(temp)/'tunnel';shutil.copytree(info['Dir'],source);source.chmod(0o755)
  for file in source.rglob('*'):
   file.chmod(0o755 if file.is_dir() else 0o644)
  for name in ['go.mod','go.sum']:shutil.copyfile(ROOT/'deploy/tunnel'/name,source/name)
  run('go','run','./scripts/runtimeprep',str(source))
  for package in ['access','tunnel']:
   (source/'cmd/cloudflared'/package/'gati_privacy_test.go').write_text((ROOT/'deploy/tunnel/privacy_test.go.txt').read_text().replace('PACKAGE',package))
  run('go','mod','download','all',cwd=source);run('go','mod','verify',cwd=source)
  if test:
   run('go','test','-count=1','-run','^TestGATINoCrashReporting$','./cmd/cloudflared/access','./cmd/cloudflared/tunnel',cwd=source)
   run('go','run','golang.org/x/vuln/cmd/govulncheck@v1.8.0','./cmd/cloudflared',cwd=source)
  run('go','build','-mod=readonly','-trimpath','-buildvcs=false','-ldflags=-X main.Version=2026.9.1-gati.1 -X main.BuildTime=reproducible -X github.com/cloudflare/cloudflared/metrics.Runtime=virtual','-o',str(out/'cloudflared'),'./cmd/cloudflared',cwd=source)
 run('go','mod','download','all',cwd=ROOT/'deploy/caddy');run('go','mod','verify',cwd=ROOT/'deploy/caddy')
 if test:
  run('go','test','-count=1','./...',cwd=ROOT/'deploy/caddy')
  run('go','run','golang.org/x/vuln/cmd/govulncheck@v1.8.0','./...',cwd=ROOT/'deploy/caddy')
 run('go','build','-mod=readonly','-trimpath','-buildvcs=false','-o',str(out/'caddy'),'.',cwd=ROOT/'deploy/caddy')

if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--output',type=pathlib.Path,required=True);p.add_argument('--test',action='store_true');a=p.parse_args();build(a.output.resolve(),a.test)
