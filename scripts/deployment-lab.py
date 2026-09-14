#!/usr/bin/env python3
"""Test built production images in an owned stack; never starts a real tunnel."""
import argparse, concurrent.futures, json, os, pathlib, shutil, subprocess, tempfile, time
from recovery import call, eventually, token
from monitor import snapshot
ROOT=pathlib.Path(__file__).resolve().parent.parent

def run(out):
 out=pathlib.Path(out).resolve();out.mkdir(parents=True,exist_ok=True)
 project='gati-deployment-lab-'+os.urandom(4).hex();checks=[];started=time.time()
 with tempfile.TemporaryDirectory(prefix=project+'-') as directory, (out/'commands.log').open('wb') as log:
  root=pathlib.Path(directory)
  def command(*args, capture=False, **kw):
   return subprocess.check_output(args,stderr=log,**kw).decode().strip() if capture else subprocess.run(args,stdout=log,stderr=log,check=True,**kw)
  for name in ['compose.production.yaml','config/gati.yaml','deploy/valkey.production.conf','scripts/init-secrets.py']:
   dest=root/name;dest.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(ROOT/name,dest)
  command('python3',str(root/'scripts/init-secrets.py'))
  command('go','build','-trimpath','-buildvcs=false','-o',str(root/'probe'),'./scripts/deployment-probe',cwd=ROOT,env={**os.environ,'CGO_ENABLED':'0'})
  env={**os.environ,'GATI_API_ROLE':'api','GATI_API_IMAGE':'gati-api:local','GATI_WEB_IMAGE':'gati-web:local','GATI_PUBLIC_HOST':'localhost','GATI_RUNTIME_DIR':str(root/'.runtime'),'GATI_VALKEY_CONFIG':str(root/'deploy/valkey.production.conf')}
  def compose(*args,capture=False):return command('docker-compose','-p',project,'-f',str(root/'compose.production.yaml'),'--profile','workers',*args,capture=capture,env=env,cwd=root)
  connector=project+'-connector'
  try:
   compose('up','-d','--no-build','--pull','never','--scale','api=2','--wait','--wait-timeout','240','valkey','api','worker','web')
   ids={service:compose('ps','-q',service,capture=True).splitlines() for service in ['api','worker','web','valkey']}
   assert len(ids['api'])==2
   command('docker','run','-d','--name',connector,'--network',project+'_edge','--ip','172.30.10.2','--read-only','--cap-drop','ALL','--security-opt','no-new-privileges:true','--log-driver','none','-p','127.0.0.1::8090','-v',str(root/'probe')+':/probe:ro','--entrypoint','/probe','gati-api:local')
   port=int(command('docker','port',connector,'8090',capture=True).rsplit(':',1)[1]);base=f'http://127.0.0.1:{port}'
   # Host is the configured public hostname; no DNS or external request is needed.
   import http.client
   def request(method,path,headers=None,data=None):
    c=http.client.HTTPConnection('127.0.0.1',port,timeout=15)
    try:
     c.request(method,path,data,{'Host':'localhost',**(headers or {})});r=c.getresponse();body=r.read(5000000)
     return r.status,dict(r.getheaders()),body
    finally:c.close()
   try:eventually(lambda:request('GET','/api/config')[0]==200,20)
   except AssertionError:
    s,h,b=request('GET','/api/config');print('Anonymous config probe failed:',s,h,b[:256],flush=True)
    for container in ids['web']:
     print('Web status:',command('docker','inspect','--format','{{.State.Status}} {{.State.ExitCode}} {{.RestartCount}}',container,capture=True),flush=True)
    raise
   for service, containers in ids.items():
    for container in containers:
     v=json.loads(command('docker','inspect',container,capture=True))[0];h=v['HostConfig']
     assert h['ReadonlyRootfs'] and h['LogConfig']['Type']=='none' and h['Memory']>0 and h['MemorySwap']==h['Memory']
     assert 'ALL' in h['CapDrop'] and 'no-new-privileges:true' in h['SecurityOpt'] and not h['PortBindings']
     assert any(u['Name']=='core' and u['Hard']==0 for u in h['Ulimits'])
     assert not any(m['Type']=='volume' or (m['Type']=='bind' and m['RW']) for m in v['Mounts'])
   checks.append('actual containers: read-only, no request logs/ports/persistent writes, no container swap/core dumps, bounded memory')
   settings=json.loads(command('docker','exec',ids['valkey'][0],'sh','-c','VALKEYCLI_AUTH=$(cat /run/secrets/health-password) valkey-cli --user health --json CONFIG GET save appendonly maxmemory-policy acllog-max-len slowlog-log-slower-than',capture=True))
   assert settings=={'save':'','appendonly':'no','maxmemory-policy':'noeviction','acllog-max-len':'0','slowlog-log-slower-than':'-1'}
   checks.append('actual Valkey persistence, eviction and metadata-log settings verified')
   command('docker','run','--rm','--network',project+'_edge','--read-only','--cap-drop','ALL','--log-driver','none','-v',str(root/'probe')+':/probe:ro','--entrypoint','/probe','gati-api:local','-check','http://web:8080/api/config')
   command('docker','run','--rm','--network',project+'_application','--read-only','--cap-drop','ALL','--log-driver','none','-v',str(root/'probe')+':/probe:ro','--entrypoint','/probe','gati-api:local','-check','http://api:8080/api/config')
   for headers in [{'Host':'wrong.test'},{'X-Gati-Lab-IP':'invalid'},{'X-Gati-Lab-IP':'198.51.100.1, 198.51.100.2'}]:assert request('GET','/api/config',headers)[0]==403
   for ip in ['198.51.100.1','2001:db8::1']:assert request('GET','/api/config',{'X-Gati-Lab-IP':ip})[0]==200
   assert request('GET','/api/config?probe=1')[0]==400
   for path in ['/api/simulation/clock','/metrics','/healthz']:assert request('GET',path)[0]==404
   assert request('POST','/api/signals',{'Authorization':'Bearer '+token(),'Content-Type':'application/json'},'x'*5000)[0] in (400,413)
   checks.append('untrusted peers, wrong host, malformed IP, query and oversized body rejected; IPv4/IPv6 accepted; no public test/ops routes')
   status,headers,html=request('GET','/');assert status==200 and b'lang="sq"' in html and 'Content-Security-Policy' in headers
   import re
   asset=re.search(rb'/assets/[^" ]+\.js',html).group().decode()
   for path in ['/',asset,'/api/map/roads']:
    for credential in ['Authorization','Cookie']:
     s,h,b=request('GET',path,{credential:'synthetic-private-canary'});assert s==200 and h['Cache-Control']=='no-store'
   assert 'immutable' in request('GET',asset)[1]['Cache-Control']
   assert 'public' in request('GET','/api/map/roads')[1]['Cache-Control']
   assert request('GET','/api/signal')[1]['Cache-Control']=='no-store'
   checks.append('built Albanian client and first-party assets; Authorization/Cookie responses bypass shared caching')
   command('node',str(ROOT/'scripts/deployment-browser.mjs'),f'http://localhost:{port}',cwd=ROOT)
   checks.append('Chromium runs the built client through production CSP/proxy: map loads, device-derived willingness, status and cancellation')
   # Use a local Host-preserving adapter for the reusable journey helper.
   transient_proxy_failures=[]
   def api(method,path,cap=None,data=None,nonce=None,network='198.51.100.40'):
    headers={'X-Gati-Lab-IP':network}
    if cap:headers['Authorization']='Bearer '+cap
    if data is not None:headers['Content-Type']='application/json'
    if nonce:headers['X-Gati-Arrival-Nonce']=nonce
    s,h,b=request(method,path,headers,json.dumps(data) if data is not None else None)
    if s>=500:transient_proxy_failures.append(s)
    return s,json.loads(b) if b and h.get('Content-Type','').startswith('application/json') else None
   config=api('GET','/api/config')[1]['config'];caps=[token() for _ in range(config['matching']['activation_count'])]
   body={'cell':'tirana-v1:100:55:55','radius_km':3,'availability_minutes':30}
   for cap in caps:assert api('POST','/api/signals',cap,body)[0]==200
   state=eventually(lambda:(lambda r:r[1] if r[0]==200 and r[1].get('invitation') else None)(api('GET','/api/signal',caps[0])),50);g=state['invitation']
   for cap in caps:assert api('GET','/api/signal',cap)[1]['invitation']['id']==g['id']
   assert api('POST','/api/going',caps[0],{'gathering_id':g['id']})[0]==200
   grid=api('GET','/api/geography')[1];x,y=g['intersection']['point'];arrival={'cell':f"tirana-v1:100:{int((x-grid['west'])/grid['lon_step'])}:{int((y-grid['south'])/grid['lat_step'])}"}
   nonce=token();assert api('POST','/api/arrival-nonce',caps[0],nonce=nonce)[0]==200
   with concurrent.futures.ThreadPoolExecutor(max_workers=8) as pool:results=list(pool.map(lambda _:api('POST','/api/arrival',caps[0],arrival,nonce),range(8)))
   assert all(r[0]==200 for r in results) and len({r[1]['arrival_until'] for r in results})==1
   assert api('DELETE','/api/arrival',caps[0])[0]==200
   assert api('POST','/api/arrival',caps[0],arrival,nonce)[0]==410
   command('docker','stop',ids['api'][0]);eventually(lambda:api('GET','/api/signal',caps[0])[0]==200,20)
   # Docker DNS can still include the stopped replica for five seconds. A
   # successful GET need not mean the next request selects the surviving peer.
   eventually(lambda:api('DELETE','/api/signal',caps[0])[0]==204,20)
   eventually(lambda:api('GET','/api/signal',caps[0])[0]==410,20)
   checks.append('two API replicas plus one worker: real activation/going, concurrent arrival retry, replay rejection, surviving API after one replica stops')
   apiMetrics=snapshot('/run/ops/gati.sock',ids['api'][1]);workerMetrics=snapshot('/run/ops/gati.sock',ids['worker'][0])
   assert set(apiMetrics['workers'])=={'view'} and {'matcher','publisher','cleanup'}<=set(workerMetrics['workers'])
   checks.append('API-only role has destination view only; matching/publication/cleanup worker and private docker-exec monitoring verified')
   # Independent network budgets survive proxying; spoofing forwarding headers cannot reset them.
   for n in range(config['limits']['new_signals_per_network_window']):assert api('POST','/api/signals',token(),body,network='198.51.100.200')[0]==200
   s,_,_=request('POST','/api/signals',{'Authorization':'Bearer '+token(),'Content-Type':'application/json','X-Gati-Lab-IP':'198.51.100.200','CF-Connecting-IP':'198.51.100.201','X-Gati-Client-IP':'198.51.100.201','X-Forwarded-For':'198.51.100.201'},json.dumps(body));assert s==429
   assert api('POST','/api/signals',token(),body,network='198.51.100.201')[0]==200
   checks.append('shared per-network admission limit survives replicas and forged forwarding headers; another network retains admission')
   report={'synthetic':True,'checks':checks,'transient_proxy_5xx_during_failover':len(transient_proxy_failures),'wall_seconds':round(time.time()-started,2),'scope':'local production images and Caddy with a fake loopback connector; no Cloudflare, real TLS/device/push or VPS capacity test'}
   (out/'deployment.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report,indent=2))
  finally:
   subprocess.run(['docker','rm','-f',connector],stdout=log,stderr=log)
   compose('down','--remove-orphans','--volumes')

if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--output',required=True);run(p.parse_args().output)
