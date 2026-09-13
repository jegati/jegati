#!/usr/bin/env python3
"""First-party local monitoring. Fixed summaries only; no participant event logs."""
import collections, http.client, json, os, pathlib, socket, threading, time, tempfile

class UnixHTTP(http.client.HTTPConnection):
    def __init__(self, path):
        super().__init__('localhost', timeout=2); self.path = str(path)
    def connect(self):
        self.sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        self.sock.settimeout(self.timeout); self.sock.connect(self.path)

def snapshot(path):
    client = UnixHTTP(path)
    try:
        client.request('GET', '/metrics'); response = client.getresponse(); raw = response.read(8193)
        if response.status != 200 or len(raw) > 8192: raise ValueError('invalid metrics')
        v = json.loads(raw)
        if not isinstance(v,dict):raise ValueError('invalid metrics object')
        if set(v) != {'version','window_seconds','retention_seconds','minimum_samples','observed_epoch','requests','failures','p95_upper_ms','workers'}: raise ValueError('unexpected metrics')
        if any(type(v[k]) is not int for k in ['version','window_seconds','retention_seconds','minimum_samples','observed_epoch','p95_upper_ms']): raise ValueError('invalid metrics')
        if v['version'] not in (1,2) or (v['window_seconds'],v['retention_seconds'],v['minimum_samples']) != (60,120,20): raise ValueError('unknown privacy policy')
        for key in ['requests','failures']:
            if v[key] not in ['suppressed','20+','50+','100+','250+','500+','1000+','5000+','10000+']: raise ValueError('invalid bucket')
        if v['p95_upper_ms'] not in [-1,0,10,50,100,300,1000,3000,10000]: raise ValueError('invalid latency')
        if not isinstance(v['workers'],dict) or not set(v['workers']) <= {'matcher','publisher','cleanup','push'}: raise ValueError('invalid workers')
        for w in v['workers'].values():
            if not isinstance(w,dict):raise ValueError('invalid worker object')
            expected={'state','last_start_seconds','last_finish_seconds','last_success_seconds'}
            if v['version']==2:expected|={'duration_upper_ms','deadline_lag_seconds'}
            if set(w) != expected or w['state'] not in ['ok','error','running']: raise ValueError('invalid worker')
            if v['version']==2 and w['duration_upper_ms'] not in [-1,0,10,50,100,300,1000,3000,10000]:raise ValueError('invalid duration')
            if any(type(w[k]) is not int or w[k]<-1 for k in w if k!='state'): raise ValueError('invalid age')
        return v
    finally: client.close()

def process_resources(pid):
    # Never inspect environment, command lines, file descriptors or process memory.
    stat = pathlib.Path(f'/proc/{pid}/stat').read_text().rsplit(')',1)[1].split()
    return {'cpu_seconds':round((int(stat[11])+int(stat[12]))/os.sysconf('SC_CLK_TCK'),2),'rss_bytes':int(stat[21])*os.sysconf('SC_PAGE_SIZE')}

def alerts(metrics):
    result=[]
    for name,w in metrics['workers'].items():
        if w.get('deadline_lag_seconds',-1)>=30:result.append(name+'_deadline_lag')
        if w['state']=='error': result.append(name+'_error')
        if w['last_success_seconds']>30 or (w['last_success_seconds']==-1 and w['last_start_seconds']>=30): result.append(name+'_stale')
    if metrics['failures']!='suppressed': result.append('server_errors')
    if metrics['p95_upper_ms'] in [1000,3000,10000,0]: result.append('slow_requests')
    return result

class Collector:
    def __init__(self, target, output, pid=None, interval=5, retention=3600, store_pid=None):
        if not 1<=interval<=60 or not 120<=retention<=86400 or retention//interval>8640: raise ValueError('invalid monitoring bounds')
        self.store_pid=store_pid; self.target=pathlib.Path(target);self.output=pathlib.Path(output);self.pid=pid
        self.interval=interval;self.retention=retention;self.rows=collections.deque(maxlen=retention//interval)
        self.stop=threading.Event();self.thread=None;self.closed=False
    def sample(self):
        row={'at':int(time.time())}
        try:
            start=time.monotonic();row['service']=snapshot(self.target);row['probe_ms']=round((time.monotonic()-start)*1000);row['alerts']=alerts(row['service'])
        except (OSError,ValueError,http.client.HTTPException): row['alerts']=['monitor_unavailable']
        if self.pid:
            try: row['resources']=process_resources(self.pid)
            except (OSError,ValueError,IndexError): pass
        if self.store_pid:
            try:row['store_resources']=process_resources(self.store_pid)
            except (OSError,ValueError,IndexError):pass
        self.rows.append(row)
        while self.rows and self.rows[0]['at']<time.time()-self.retention:self.rows.popleft()
        return row
    def save(self):
        value={'version':1,'retention_seconds':self.retention,'expires_at':int(time.time())+self.interval*3,'samples':list(self.rows)}
        temporary=None
        try:
            fd,name=tempfile.mkstemp(prefix='.gati-monitor-',dir=self.output.parent)
            temporary=pathlib.Path(name)
            with os.fdopen(fd,'w') as stream:stream.write(json.dumps(value,separators=(',',':'))+'\n')
            temporary.replace(self.output)
        except OSError: pass # Monitoring storage failure cannot interrupt the application.
        finally:
            if temporary:
                try:temporary.unlink(missing_ok=True)
                except OSError:pass
    def run(self):
        while not self.stop.is_set(): self.sample();self.save();self.stop.wait(self.interval)
    def start(self):self.thread=threading.Thread(target=self.run,daemon=True);self.thread.start();return self
    def close(self):
        if self.closed:return
        self.closed=True;self.stop.set()
        if self.thread:self.thread.join(timeout=3)
        self.sample();self.save()

if __name__=='__main__':
    import argparse
    p=argparse.ArgumentParser();p.add_argument('--socket',required=True);p.add_argument('--output',required=True);p.add_argument('--pid',type=int);p.add_argument('--interval',type=int,default=5);p.add_argument('--retention',type=int,default=3600)
    a=p.parse_args();collector=Collector(a.socket,a.output,a.pid,a.interval,a.retention)
    try:collector.run()
    except KeyboardInterrupt:collector.close()
