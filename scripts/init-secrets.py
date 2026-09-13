#!/usr/bin/env python3
"""Create private local service secrets; never print values or overwrite them."""
from pathlib import Path
import hashlib, os, secrets, argparse
parser=argparse.ArgumentParser()
parser.add_argument("--simulation", action="store_true")
args=parser.parse_args()
root=Path(__file__).resolve().parent.parent / '.runtime'
root.mkdir(mode=0o700,exist_ok=True)
if args.simulation:
    root=root/'simulation';root.mkdir(mode=0o700,exist_ok=True)
prefix='gati-sim:' if args.simulation else 'gati:'
os.chmod(root,0o700)
passwords={}
for role in ['app','health','control'] if args.simulation else ['app','health']:
    path=root/(role+'-password')
    if not path.exists():
        with path.open('x') as f: f.write(secrets.token_hex(32)+'\n')
        # Parent is 0700 on host; individual file mounts need container UID reads.
        path.chmod(0o444)
    passwords[role]=path.read_text().strip()
acl=['user default off',
     'user health on #'+hashlib.sha256(passwords['health'].encode()).hexdigest()+' ~* +ping +config|get',
     'user app on #'+hashlib.sha256(passwords['app'].encode()).hexdigest()+' ~'+prefix+'* +ping +hello +get +mget +set +exists +del +pttl +pexpire +pexpireat +zadd +zrem +zremrangebyscore +zrangebyscore +zrange +zrank +zscan +zcard +zcount +zscore +eval +evalsha +script|load +time +incr']
p=root/'users.acl'
content='\n'.join(acl)+'\n'
if not p.exists() or p.read_text()!=content:
    temporary=root/'users.acl.tmp'
    temporary.write_text(content);temporary.chmod(0o444)
    temporary.replace(p)
print('Local service secrets ready; values are never printed.')
