#!/usr/bin/env python3
"""Owned Valkey no-eviction saturation and recovery; never fills a real store."""
import argparse,json,pathlib,time
from lab import Lab
from recovery import call,token,eventually
p=argparse.ArgumentParser();p.add_argument('--output',required=True);a=p.parse_args()
with Lab(a.output,store_mb=8) as lab:
 body={'cell':'tirana-v1:100:55:55','radius_km':3,'availability_minutes':30}
 cap=token();status,original=call(lab.base,'POST','/api/signals',cap,body);assert status==200
 keys=[];saturated=False
 try:
  for n in range(128):
   key='gati:synthetic-pressure:'+str(n)
   try:lab.store('SET',key,'x'*131072,'EX',120);keys.append(key)
   except RuntimeError:saturated=True;break
  assert saturated,'owned fixture did not reach memory limit'
  uncertain=token();status,_=call(lab.base,'POST','/api/signals',uncertain,body);assert status==503,'memory saturation must fail closed'
 finally:
  for key in keys:lab.store('DEL',key)
 restored=eventually(lambda:(lambda r:r[1] if r[0]==200 else None)(call(lab.base,'GET','/api/signal',cap)),20)
 assert restored['expires_at']==original['expires_at'],'memory pressure evicted or renewed a live signal'
 assert call(lab.base,'POST','/api/signals',uncertain,body)[0]==200
 assert call(lab.base,'DELETE','/api/signal',cap)[0]==204
 assert call(lab.base,'POST','/api/signals',cap,body)[0]==410
 report={'synthetic':True,'noeviction_pressure_passed':True,'store_maxmemory_mb':8,'filler_blocks':len(keys),'checks':['saturation returns 503','live credential/deadline survives pressure','uncertain creation can be retried','cancellation remains terminal after recovery']}
 (pathlib.Path(a.output)/'pressure.json').write_text(json.dumps(report,indent=2)+'\n');print('Owned memory-pressure checks passed')
