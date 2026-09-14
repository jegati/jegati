#!/usr/bin/env python3
"""Owned, isolated production-binary lab. No user-supplied remote targets."""
import contextlib, json, os, pathlib, re, shutil, socket, subprocess, tarfile, tempfile, time, urllib.request
from monitor import Collector
ROOT=pathlib.Path(__file__).resolve().parent.parent

def command(*args,**kwargs): return subprocess.run(list(args),cwd=ROOT,check=True,**kwargs)
def free_port():
    with socket.socket() as s:s.bind(('127.0.0.1',0));return s.getsockname()[1]
def get(base,path):
    with urllib.request.urlopen(base+path,timeout=3) as r:return r.status,json.load(r)
class Lab:
    def __init__(self,output,replicas=1,store_mb=512,source_revision=None):
        if source_revision is not None and not re.fullmatch(r"[a-f0-9]{40}",source_revision):raise ValueError("source revision must be a full local commit hash")
        self.source_revision=source_revision;self.source_root=ROOT
        self.output=pathlib.Path(output).resolve()
        if (self.output/'manifest.json').exists():raise ValueError('choose a new output directory; existing lab evidence is preserved')
        self.output.mkdir(parents=True,exist_ok=True)
        self.temporary=tempfile.TemporaryDirectory(prefix='gati-lab-');self.directory=pathlib.Path(self.temporary.name)
        self.container=None;self.apis=[];self.collectors=[];self.replicas=replicas;self.store_mb=store_mb
    def __enter__(self):
        try:self.start();return self
        except BaseException:self.__exit__(None,None,None);raise
    def start(self):
        command('python3','scripts/init-secrets.py',stdout=subprocess.DEVNULL)
        self.binary=self.directory/'gati-lab'
        if self.source_revision:
            self.source_root=self.directory/'source';self.source_root.mkdir()
            archive=self.directory/'source.tar'
            with archive.open('wb') as stream:command('git','archive',self.source_revision,stdout=stream)
            with tarfile.open(archive) as tar:tar.extractall(self.source_root,filter='data')
        subprocess.run(['go','build','-trimpath','-buildvcs=false','-o',str(self.binary),'./cmd/gati'],cwd=self.source_root,check=True)
        c=json.loads(subprocess.check_output([str(self.binary),'-mode','config-show'],cwd=self.source_root));self.config=c
        if c['notifications']['push_enabled']:
            subprocess.run(['go','run','./cmd/gati-push-keys','-directory',str(self.directory/'push')],cwd=self.source_root,check=True,stdout=subprocess.DEVNULL)
        (self.output/'config.json').write_text(json.dumps(c,indent=2)+'\n')
        conf=(ROOT/'deploy/valkey.dev.conf').read_text().replace('maxmemory 128mb',f'maxmemory {self.store_mb}mb')
        (self.directory/'valkey.conf').write_text(conf);(self.directory/'valkey.conf').chmod(0o644)
        image=next(line.split('=',1)[1].strip('"\'') for line in (ROOT/'deploy/images.env').read_text().splitlines() if line.startswith('GATI_VALKEY_IMAGE='))
        self.container=subprocess.check_output(['docker','run','-d','--read-only','--user','999:999','--cap-drop','ALL','--security-opt','no-new-privileges','--ulimit','core=0','--memory',f'{self.store_mb+128}m','--memory-swap',f'{self.store_mb+128}m','--log-driver','none','--tmpfs','/data:size=128m','-p',f'127.0.0.1:{free_port()}:6379','-v',f'{self.directory}/valkey.conf:/etc/valkey/valkey.conf:ro','-v',f'{ROOT}/.runtime/users.acl:/run/secrets/users.acl:ro','-v',f'{ROOT}/.runtime/health-password:/run/secrets/health-password:ro',image,'valkey-server','/etc/valkey/valkey.conf'],text=True).strip()
        self.store_address=subprocess.check_output(['docker','port',self.container,'6379/tcp'],text=True).strip()
        for _ in range(100):
            probe=subprocess.run(['docker','exec',self.container,'sh','-c','VALKEYCLI_AUTH="$(cat /run/secrets/health-password)" valkey-cli --user health ping'],capture_output=True,text=True)
            if probe.returncode==0 and probe.stdout.strip()=='PONG':break
            time.sleep(.1)
        else:raise RuntimeError('owned test store unavailable')
        self.store_pid=self.find_store_pid()
        for i in range(self.replicas):self.start_api(i)
        import hashlib,platform
        manifest={'binary_source_revision':self.source_revision,'source_revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'source_dirty':bool(subprocess.check_output(['git','status','--porcelain','--untracked-files=no'],cwd=ROOT)),'platform':platform.platform(),'cpu_model':next((line.split(':',1)[1].strip() for line in pathlib.Path('/proc/cpuinfo').read_text().splitlines() if line.startswith('model name')),'unknown'),'logical_cpus':os.cpu_count(),'store_maxmemory_mb':self.store_mb,'same_host_generator':True,'config_sha256':get(self.base,'/api/config')[1]['sha256'],'binary_sha256':hashlib.sha256(self.binary.read_bytes()).hexdigest()}
        (self.output/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
    def start_api(self,index):
        port=free_port();sock=self.directory/f'ops-{index}.sock'
        if sock.exists():sock.unlink() # This private lab owns the dead process/socket.
        args=[str(self.binary),'-listen',f'127.0.0.1:{port}','-config',str(self.output/'config.json'),'-store-address',self.store_address,'-store-password-file',str(ROOT/'.runtime/app-password'),'-monitor-socket',str(sock),'-roads',str(self.source_root/'data/tirana/roads.geojson'),'-intersections',str(self.source_root/'data/tirana/intersections.json')]
        if self.config['notifications']['push_enabled']:args+=['-push-key-file',str(self.directory/'push/vapid.json')]
        with (self.output/f'api-{index}-startup.log').open('w') as log:
            process=subprocess.Popen(args,cwd=ROOT,stdout=log,stderr=log)
        value={'process':process,'base':f'http://127.0.0.1:{port}','socket':sock}
        if index<len(self.apis):
            self.apis[index]['collector'].close();self.apis[index]=value
        else:self.apis.append(value)
        deadline=time.monotonic()+180
        while time.monotonic()<deadline:
            if process.poll() is not None:raise RuntimeError('owned API exited during startup')
            try:
                if get(value['base'],'/healthz')[0]==200:break
            except (OSError,ValueError):pass
            time.sleep(.1)
        else:raise RuntimeError('owned API readiness timeout')
        collector=Collector(sock,self.output/f'monitor-{index}.json',process.pid,store_pid=self.store_pid).start();self.collectors.append(collector);value['collector']=collector
    def store(self,*args):
        # Only the owned synthetic store. Use the existing restricted app role;
        # never print credentials, command values or returned participant records.
        host,port=self.store_address.split(':')
        with socket.create_connection((host,int(port)),timeout=3) as conn:
            stream=conn.makefile('rb')
            def read():
                line=stream.readline();kind=line[:1];value=line[1:-2]
                if kind==b'+':return value.decode()
                if kind==b':':return int(value)
                if kind==b'$':
                    n=int(value)
                    if n<0:return None
                    data=stream.read(n);stream.read(2);return data.decode()
                if kind==b'*':return [read() for _ in range(int(value))]
                raise RuntimeError('owned store operation unavailable')
            def send(values):
                parts=[str(v).encode() for v in values]
                conn.sendall(b'*'+str(len(parts)).encode()+b'\r\n'+b''.join(b'$'+str(len(v)).encode()+b'\r\n'+v+b'\r\n' for v in parts));return read()
            password=(ROOT/'.runtime/app-password').read_text().strip()
            send(['AUTH','app',password]);return send(args)
    def find_store_pid(self):
        rows=subprocess.check_output(['docker','top',self.container,'-eo','pid,comm'],text=True).splitlines()[1:]
        return next(int(row.split()[0]) for row in rows if row.split()[-1]=='valkey-server')
    def ownership_file(self):
        p=self.directory/'owner.json';p.write_text(json.dumps({'base':self.base,'pid':self.apis[0]['process'].pid}));p.chmod(0o600);return p
    @property
    def base(self):return self.apis[0]['base']
    def stop_api(self,index,kill=False):
        process=self.apis[index]['process']
        if process.poll() is None:
            (process.kill if kill else process.terminate)()
            try:process.wait(timeout=10)
            except subprocess.TimeoutExpired:process.kill();process.wait()
    def __exit__(self,*_):
        for c in self.collectors:c.close()
        for i in range(len(self.apis)):self.stop_api(i)
        if self.container:subprocess.run(['docker','rm','-f',self.container],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
        self.temporary.cleanup()

if __name__=='__main__':
    import argparse
    p=argparse.ArgumentParser();p.add_argument('--output',default='reports/local/browser-lab');p.add_argument('args',nargs=argparse.REMAINDER);a=p.parse_args()
    with Lab(a.output) as lab:
        env={**os.environ,'GATI_TEST_API':lab.base,'GATI_API_PROXY':lab.base};args=a.args[1:] if a.args[:1]==['--'] else a.args
        if not args:raise SystemExit('a local test command is required')
        command(*args,env=env)
