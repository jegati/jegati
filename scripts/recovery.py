#!/usr/bin/env python3
"""Real process/lease/store faults, confined to a disposable local lab."""
import argparse,base64,concurrent.futures,hashlib,http.client,json,os,pathlib,signal,subprocess,time,urllib.parse
from lab import Lab,command
from monitor import snapshot,alerts

def token():return base64.urlsafe_b64encode(os.urandom(32)).rstrip(b'=').decode()
def call(base,method,path,cap=None,data=None,nonce=None):
 u=urllib.parse.urlsplit(base);client=http.client.HTTPConnection(u.hostname,u.port,timeout=5)
 headers={}
 if cap:headers['Authorization']='Bearer '+cap
 if nonce:headers['X-Gati-Arrival-Nonce']=nonce
 if data is not None:headers['Content-Type']='application/json'
 try:
  client.request(method,path,json.dumps(data) if data is not None else None,headers);r=client.getresponse();raw=r.read(1000001)
  return r.status,json.loads(raw) if raw else None
 finally:client.close()
def eventually(check,seconds=35):
 deadline=time.monotonic()+seconds
 while time.monotonic()<deadline:
  try:
   result=check()
   if result:return result
  except (OSError,ValueError):pass
  time.sleep(.5)
 raise AssertionError('bounded recovery condition did not hold')

def run(out):
 checks=[];start=time.time()
 with Lab(out,replicas=2) as lab:
  a,b=lab.apis;cell='tirana-v1:100:55:55';body={'cell':cell,'radius_km':3,'availability_minutes':30}
  caps=[token() for _ in range(lab.config['matching']['activation_count'])]
  initial=[]
  for cap in caps:
   code,v=call(b['base'],'POST','/api/signals',cap,body);assert code==200;initial.append(v)
  os.kill(a['process'].pid,signal.SIGSTOP)
  try:
   # The second real worker must take over the expiring lease and finish activation.
   state=eventually(lambda: (lambda r:r[1] if r[0]==200 and r[1].get('invitation') else None)(call(b['base'],'GET','/api/signal',caps[0])),45)
   g=state['invitation'];assert g['ends_at']==min(v['expires_at'] for v in initial)
   eventually(lambda:any('monitor_unavailable' in r['alerts'] for r in a['collector'].rows),10)
   checks.append('paused owner detected; second replica activated a frozen destination/deadline')
  finally:os.kill(a['process'].pid,signal.SIGCONT)
  ids=[]
  for cap in caps:
   code,v=call(b['base'],'GET','/api/signal',cap);assert code==200;ids.append(v['invitation']['id'])
  assert len(set(ids))==1
  assert call(b['base'],'POST','/api/going',caps[0],{'gathering_id':g['id']})[0]==200
  nonce=token();assert call(b['base'],'POST','/api/arrival-nonce',caps[0],nonce=nonce)[0]==200
  _,grid=call(b['base'],'GET','/api/geography');x,y=g['intersection']['point']
  arrival={'cell':f"tirana-v1:100:{int((x-grid['west'])/grid['lon_step'])}:{int((y-grid['south'])/grid['lat_step'])}"}
  with concurrent.futures.ThreadPoolExecutor(max_workers=12) as pool:
   results=list(pool.map(lambda n:call(lab.apis[n%2]['base'],'POST','/api/arrival',caps[0],arrival,nonce),range(12)))
  assert all(r[0]==200 for r in results);assert len({r[1]['arrival_until'] for r in results})==1
  assert lab.store('ZCARD','gati:arrivals:'+g['id'])==1
  assert call(b['base'],'DELETE','/api/arrival',caps[0])[0]==200
  assert call(b['base'],'POST','/api/arrival',caps[0],arrival,nonce)[0]==410
  checks.append('two replicas preserve one arrival and reject replay after retraction')
  # A response lost with a crashed API is recovered by an idempotent same-capability retry.
  cap=token();code,created=call(a['base'],'POST','/api/signals',cap,body);assert code==200
  lab.stop_api(0,kill=True)
  code,retry=call(b['base'],'POST','/api/signals',cap,body);assert code==200 and retry['expires_at']==created['expires_at']
  assert call(b['base'],'DELETE','/api/signal',cap)[0]==204
  lab.start_api(0)
  assert call(lab.base,'POST','/api/signals',cap,body)[0]==410
  checks.append('API crash/restart preserves idempotency and cancellation tombstone')
  command('docker','pause',lab.container,stdout=subprocess.DEVNULL)
  try:
   assert call(b['base'],'GET','/api/signal',caps[0])[0]==503
   eventually(lambda:any(w['state']=='error' for w in snapshot(b['socket'])['workers'].values()),15)
  finally:command('docker','unpause',lab.container,stdout=subprocess.DEVNULL)
  eventually(lambda:call(b['base'],'GET','/api/signal',caps[0])[0]==200)
  checks.append('store interruption detected by monitoring; requests recover after reconnection')
  command('docker','restart',lab.container,stdout=subprocess.DEVNULL)
  eventually(lambda:lab.store('PING')=='PONG',15)
  for c in lab.collectors:c.store_pid=lab.find_store_pid()
  assert call(b['base'],'GET','/api/signal',caps[0])[0]==410
  assert lab.store('ZCARD','gati:expiry')==0
  checks.append('nonpersistent store restart loses sessions without reconstruction')
  # Monitor consumers are stopped; the application must remain usable.
  for c in lab.collectors:c.close()
  code,_=call(b['base'],'POST','/api/signals',token(),body);assert code==200
  checks.append('stopping monitoring does not block participation')
  report={'synthetic':True,'checks':checks,'wall_seconds':round(time.time()-start,2)}
  (pathlib.Path(out)/'recovery.json').write_text(json.dumps(report,indent=2)+'\n')
  print('Recovery checks passed:',len(checks))
if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--output',required=True);run(p.parse_args().output)
