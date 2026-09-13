#!/usr/bin/env python3
import argparse,json,pathlib
from lab import Lab,command
p=argparse.ArgumentParser();p.add_argument('--output',required=True);p.add_argument('--sizes',default='1000,10000,100000');p.add_argument('--seconds',type=int,default=20);p.add_argument('--distribution',choices=['uniform','hotspot'],default='uniform');p.add_argument('--burst',action='store_true');a=p.parse_args()
with Lab(a.output) as lab:
 command('go','build','-o','bin/gati-loadtest','./cmd/loadtest')
 args=['bin/gati-loadtest','-lab-file',str(lab.ownership_file()),'-output',str(lab.output/'load.json'),'-sizes',a.sizes,'-seconds',str(a.seconds),'-distribution',a.distribution]
 if a.burst:args+=['-burst']
 command(*args)

command("python3","scripts/load-report.py",str(lab.output))
command("python3","scripts/check-load.py",str(lab.output/"load.json"))
