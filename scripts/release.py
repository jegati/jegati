#!/usr/bin/env python3
"""Local immutable release preparation and explicit operator deployment commands.

Never uploads source, creates accounts, opens host ports or handles remote login.
The operator chooses the machine; --edge is the explicit public-tunnel action.
"""
import argparse, hashlib, io, json, os, pathlib, re, subprocess, tarfile, tempfile
ROOT=pathlib.Path(__file__).resolve().parent.parent

def run(*args,**kw):return subprocess.check_output(args,**kw).decode().strip()
def digest(path):
 h=hashlib.sha256()
 with path.open('rb') as f:
  for block in iter(lambda:f.read(1024*1024),b''):h.update(block)
 return h.hexdigest()
def runtime_dir(root,project):return root.parent/('.'+project+'-runtime')
def compose(root,project,*args,replicas=1,edge=False):
 runtime=runtime_dir(root,project)
 settings=dict(line.split('=',1) for line in (root/'deployment.env').read_text().splitlines() if line)
 env={**os.environ,**settings,'GATI_API_ROLE':'combined' if replicas==1 else 'api','GATI_RUNTIME_DIR':str(runtime),'GATI_VALKEY_CONFIG':str(runtime/'valkey.conf')}
 push=json.loads((root/'release.json').read_text()).get('push_enabled',False)
 return run('docker-compose','--env-file',str(root/'deployment.env'),'-p',project,'-f',str(root/'compose.production.yaml'),*(['-f',str(root/'compose.production.push.yaml')] if push else []),*(['--profile','workers'] if replicas>1 else []),*(['--profile','edge'] if edge else []),*args,cwd=root,env=env)

def prepare(root,host):
 if root.exists():raise ValueError('release output must be a new directory')
 if not re.fullmatch(r'[a-z0-9](?:[a-z0-9.-]{0,251}[a-z0-9])?',host) or '..' in host:raise ValueError('use a lowercase ASCII hostname without a port')
 if run('git','status','--porcelain','--untracked-files=no',cwd=ROOT):raise ValueError('commit tracked changes before preparing a release')
 revision=run('git','rev-parse','HEAD',cwd=ROOT);archive=subprocess.check_output(['git','archive',revision],cwd=ROOT)
 root.mkdir(parents=True)
 with tarfile.open(fileobj=io.BytesIO(archive)) as tar:tar.extractall(root,filter='data')
 (root/'source.tar').write_bytes(archive)
 refs={'api':'gati-api:release-'+revision[:12],'web':'gati-web:release-'+revision[:12]}
 (root/'deployment.env').write_text(f'GATI_API_IMAGE={refs["api"]}\nGATI_WEB_IMAGE={refs["web"]}\nGATI_PUBLIC_HOST={host}\n')
 with (root/'build.log').open('wb') as log:
  subprocess.run(['docker-compose','--env-file',str(root/'deployment.env'),'-f',str(root/'compose.production.yaml'),'build','api','web'],cwd=root,env={**os.environ,'GATI_API_IMAGE':refs['api'],'GATI_WEB_IMAGE':refs['web'],'GATI_PUBLIC_HOST':host},stdout=log,stderr=log,check=True)
 images={k:{'reference':ref,'image_id':run('docker','image','inspect','--format','{{.Id}}',ref)} for k,ref in refs.items()}
 upstream=dict(line.split('=',1) for line in (root/'deploy/images.env').read_text().splitlines() if line and not line.startswith('#'))
 for service,key in [('valkey','GATI_VALKEY_IMAGE'),('tunnel','GATI_CLOUDFLARED_IMAGE')]:
  ref=upstream[key]
  try:identity=run('docker','image','inspect','--format','{{.Id}}',ref)
  except subprocess.CalledProcessError:
   run('docker','pull',ref);identity=run('docker','image','inspect','--format','{{.Id}}',ref)
  images[service]={'reference':ref,'image_id':identity}
 effective_raw=run('docker','run','--rm','--network','none',refs['api'],'-mode','config-show','-config','/config/gati.yaml')
 effective=json.loads(effective_raw)
 with tempfile.TemporaryDirectory(prefix='gati-release-assets-') as d:
  container=run('docker','create','--network','none',refs['web'])
  try:run('docker','cp',container+':/srv/.',d)
  finally:run('docker','rm',container)
  assets={str(p.relative_to(d)):digest(p) for p in sorted(pathlib.Path(d).rglob('*')) if p.is_file()}
 run('docker','save','-o',str(root/'images.tar'),*[image['reference'] for image in images.values()])
 immutable={str(p.relative_to(root)):digest(p) for p in sorted(root.rglob('*')) if p.is_file() and p.name!='build.log'}
 manifest={'version':1,'source_revision':revision,'config_sha256':digest(root/'config/gati.yaml'),'effective_config_sha256':hashlib.sha256(effective_raw.encode()).hexdigest(),'push_enabled':effective['notifications']['push_enabled'],'images':images,'browser_assets':assets,'files':immutable,'scope':'source/files and local image identities; served asset comparison and privileged host inspection are separate checks; no remote honesty attestation'}
 (root/'release.json').write_text(json.dumps(manifest,indent=2)+'\n')
 (root/'release.sha256').write_text(digest(root/'release.json')+'  release.json\n')
 print('Prepared local release',revision,'manifest SHA256',digest(root/'release.json'))

def verify_files(root):
 manifest=json.loads((root/'release.json').read_text())
 if manifest.get('version')!=1:raise ValueError('unsupported release manifest')
 if (root/'release.sha256').read_text().split()[0]!=digest(root/'release.json'):raise ValueError('release manifest checksum changed')
 for name,want in manifest['files'].items():
  path=root/name
  if path.resolve()!=path or not path.is_relative_to(root) or digest(path)!=want:raise ValueError('release file verification failed')
 return manifest

def verify_images(manifest):
 for image in manifest['images'].values():
  if run('docker','image','inspect','--format','{{.Id}}',image['reference'])!=image['image_id']:raise ValueError('release image identity changed')

def preflight():
 problems=[]
 if pathlib.Path('/proc/swaps').read_text().strip().count('\n'):problems.append('host swap is enabled; configure the deployment host before public launch')
 if pathlib.Path('/sys/power/resume').exists() and pathlib.Path('/sys/power/resume').read_text().strip()!='0:0':problems.append('host resume/hibernation device is configured')
 info=json.loads(run('docker','info','--format','{{json .}}'))
 if not info.get('MemoryLimit') or not info.get('SwapLimit'):problems.append('Docker memory/swap limits are not supported')
 if info.get('MemTotal',0)<6*1024**3:problems.append('less than 6 GiB visible to Docker; validate a smaller configuration before deployment')
 print(json.dumps({'passed':not problems,'problems':problems,'manual_checks':['provider RAM snapshots disabled; no host request/body tracing or process-memory capture','host patching, access control, firewall and core dump policy reviewed','Cloudflare cache/trust/TLS and privacy settings verified on the actual hostname']},indent=2))
 return not problems

def verify_running(root,project,manifest,replicas,edge=False):
 verify_images(manifest);seen=[]
 for service in ['valkey','api','web']+(['worker'] if replicas>1 else [])+(['tunnel'] if edge else []):
  ids=compose(root,project,'ps','-q',service,replicas=replicas,edge=edge).splitlines()
  if len(ids)!=(replicas if service=='api' else 1):raise ValueError('unexpected running service count')
  for container in ids:
   value=json.loads(run('docker','inspect',container))[0];h=value['HostConfig']
   if value['State']['Status']!='running' or value['State'].get('Health',{}).get('Status','healthy')!='healthy':raise ValueError('unhealthy release service')
   if not h['ReadonlyRootfs'] or h['LogConfig']['Type']!='none' or h['PortBindings'] or h['Memory']<=0 or h['MemorySwap']!=h['Memory']:raise ValueError('runtime privacy/resource settings differ')
   if 'ALL' not in h['CapDrop'] or 'no-new-privileges:true' not in h['SecurityOpt'] or not any(u['Name']=='core' and u['Hard']==0 for u in h['Ulimits']):raise ValueError('runtime privilege/core limits differ')
   if any(m['Type']=='volume' or (m['Type']=='bind' and m['RW']) for m in value['Mounts']):raise ValueError('persistent writable mount')
   key='api' if service=='worker' else service
   if key in manifest['images'] and value['Image']!=manifest['images'][key]['image_id']:raise ValueError('running image differs from release')
   if service in ['api','worker']:
    mounted=run('docker','exec',container,'/gati','-mode','config-show','-config','/config/gati.yaml')
    expected=run('docker','run','--rm','--network','none',manifest['images']['api']['reference'],'-mode','config-show')
    if mounted!=expected:raise ValueError('mounted functional settings differ from the release image')
   if service=='valkey':
    settings=json.loads(run('docker','exec',container,'sh','-c','VALKEYCLI_AUTH=$(cat /run/secrets/health-password) valkey-cli --user health --json CONFIG GET save appendonly maxmemory-policy acllog-max-len slowlog-log-slower-than'))
    if settings!={'save':'','appendonly':'no','maxmemory-policy':'noeviction','acllog-max-len':'0','slowlog-log-slower-than':'-1'}:raise ValueError('store privacy settings differ')
   seen.append(service)
 print('Verified local running services:',', '.join(seen),'; remote honesty and edge behavior are not established.')

def activate(root,project,manifest,replicas,edge,current=None):
 runtime=runtime_dir(root,project)
 if current:
  previous=verify_files(current)
  if previous['config_sha256']!=manifest['config_sha256']:raise ValueError('automatic rollback requires identical functional config; review schema/state compatibility explicitly')
  if runtime_dir(current,project)!=runtime:raise ValueError('rollback releases must share the same parent and stable operational directory')
 if edge:
  if not preflight():raise ValueError('host preflight failed')
  if not (runtime/'tunnel-token').is_file():raise ValueError('named tunnel credential must be provisioned privately on the host')
  if 'GATI_PUBLIC_HOST=localhost\n' in (root/'deployment.env').read_text():raise ValueError('public deployment needs a real hostname release')
 runtime.mkdir(mode=0o700,exist_ok=True)
 if manifest.get('push_enabled') and not (runtime/'vapid.json').is_file():raise ValueError('optional push requires its privately provisioned VAPID key')
 run('python3',str(root/'scripts/init-secrets.py'),'--directory',str(runtime),*(['--check'] if (runtime/'app-password').is_file() else []))
 if not (runtime/'valkey.conf').exists():
  import shutil
  shutil.copyfile(root/'deploy/valkey.production.conf',runtime/'valkey.conf')
 elif digest(runtime/'valkey.conf')!=digest(root/'deploy/valkey.production.conf'):raise ValueError('store settings changed; review state-loss/restart implications explicitly')
 try:verify_images(manifest)
 except subprocess.CalledProcessError:
  run('docker','load','-i',str(root/'images.tar'));verify_images(manifest)
 services=['valkey','api','web']+(['worker'] if replicas>1 else [])+(['tunnel'] if edge else [])
 print(compose(root,project,'up','-d','--no-build','--pull','never','--scale',f'api={replicas}','--wait','--wait-timeout','240',*services,replicas=replicas,edge=edge))
 verify_running(root,project,manifest,replicas,edge)

def main():
 p=argparse.ArgumentParser();p.add_argument('action',choices=['prepare','verify','up','rollback','preflight']);p.add_argument('--release',type=pathlib.Path);p.add_argument('--current',type=pathlib.Path);p.add_argument('--public-host',default='localhost');p.add_argument('--project',default='gati-production');p.add_argument('--replicas',type=int,choices=[1,2],default=1);p.add_argument('--edge',action='store_true');p.add_argument('--running',action='store_true');a=p.parse_args()
 if not re.fullmatch(r'[a-z][a-z0-9-]{1,63}',a.project):raise ValueError('invalid project name')
 if a.action=='preflight':raise SystemExit(0 if preflight() else 1)
 if a.release is None:p.error('--release is required')
 root=a.release.resolve()
 if a.action=='prepare':prepare(root,a.public_host);return
 manifest=verify_files(root)
 if a.action=='verify':
  if a.running:verify_running(root,a.project,manifest,a.replicas,a.edge)
  else:print('Release file hashes verified. Compare release.sha256 through an independent channel.')
 elif a.action=='rollback':
  if a.current is None:p.error('rollback requires --current for compatibility and operational secrets')
  activate(root,a.project,manifest,a.replicas,a.edge,a.current.resolve())
 else:activate(root,a.project,manifest,a.replicas,a.edge)

if __name__=='__main__':main()
