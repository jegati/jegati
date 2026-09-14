#!/usr/bin/env python3
"""Scan exact local release images; no uploads, running services or remote targets."""
import argparse,datetime,hashlib,json,os,pathlib,subprocess,tempfile
from release import verify_files,verify_images
ROOT=pathlib.Path(__file__).resolve().parent.parent
BINARIES={'api':'/gati','web':'/usr/bin/caddy','tunnel':'/usr/local/bin/cloudflared'}

def classify(service,findings,exceptions,today):
 accepted=[];unresolved=[]
 for item in findings:
  matches=[e for e in exceptions if service in e['services'] and e['id']==item['id'] and e['package']==item['package'] and e['version']==item['version']]
  valid=[e for e in matches if datetime.date.fromisoformat(e['reviewed'])<=today<=datetime.date.fromisoformat(e['expires']) and e.get('reason') and e.get('owner') and e.get('evidence')]
  if len(valid)==1:accepted.append({**item,'exception':valid[0]})
  else:unresolved.append(item)
 return accepted,unresolved

def audit(release,out,trivy):
 if out.exists():raise ValueError('audit output must be a new directory')
 manifest=verify_files(release);verify_images(manifest)
 expected_version=dict(line.split('=',1) for line in (ROOT/'scripts/toolchain/versions.env').read_text().splitlines() if line and not line.startswith('#'))['GATI_TRIVY_VERSION']
 version=json.loads(subprocess.check_output([trivy,'version','--format','json']))
 if version['Version']!=expected_version:raise ValueError('unexpected scanner version')
 exceptions=json.loads((release/'deploy/security-exceptions.json').read_text())['exceptions']
 out.mkdir(parents=True)
 now=datetime.datetime.now(datetime.timezone.utc);passed=True
 result={'source_revision':manifest['source_revision'],'release_manifest_sha256':hashlib.sha256((release/'release.json').read_bytes()).hexdigest(),'scanned_at':now.isoformat(),'scanner':version,'scope':'local image inventory plus Go binary symbol checks; no exploitability or remote attestation guarantee','images':{}}
 with tempfile.TemporaryDirectory(prefix='gati-image-audit-') as temporary,(out/'commands.log').open('wb') as log:
  def run(*args):return subprocess.check_output(args,stderr=log).decode().strip()
  for name,image in manifest['images'].items():
   identity=image['image_id'];report=out/(name+'.json')
   # Scan the ID, never a mutable tag, and retain all findings (including unknown severity).
   subprocess.run([trivy,'image','--image-src','docker','--scanners','vuln','--format','json','--output',str(report),identity],stdout=log,stderr=log,check=True)
   scan=json.loads(report.read_text());findings=[]
   if not scan.get('Results'):raise ValueError('scanner produced no package results')
   for target in scan['Results']:
    for v in target.get('Vulnerabilities',[]):findings.append({'id':v['VulnerabilityID'],'package':v['PkgName'],'version':v['InstalledVersion'],'fixed':v.get('FixedVersion'),'severity':v['Severity']})
   accepted,unresolved=classify(name,findings,exceptions,now.date())
   entry={'image_id':identity,'accepted_findings':accepted,'unresolved_findings':unresolved}
   if name in BINARIES:
    binary=pathlib.Path(temporary)/name;container=run('docker','create','--network','none',identity)
    try:run('docker','cp',container+':'+BINARIES[name],str(binary))
    finally:run('docker','rm',container)
    entry['binary_sha256']=hashlib.sha256(binary.read_bytes()).hexdigest()
    if entry['binary_sha256']!=manifest['runtime_binary_sha256'][name]:raise ValueError('scanned binary differs from recorded release')
    with (out/(name+'-govuln.txt')).open('wb') as output:
     check=subprocess.run(['go','run','golang.org/x/vuln/cmd/govulncheck@v1.8.0','-mode','binary',str(binary)],stdout=output,stderr=output)
    entry['binary_scan_passed']=check.returncode==0
    passed=passed and entry['binary_scan_passed']
   subprocess.run([trivy,'image','--image-src','docker','--scanners','vuln','--format','cyclonedx','--output',str(out/(name+'-sbom.json')),identity],stdout=log,stderr=log,check=True)
   result['images'][name]=entry;passed=passed and not unresolved
   print(name+': '+str(len(unresolved))+' unresolved; '+str(len(accepted))+' scoped exceptions',flush=True)
 result['passed']=bool(passed)
 # Capture the database metadata after the first scan may have refreshed it.
 result['scanner']=json.loads(subprocess.check_output([trivy,'version','--format','json']))
 (out/'audit.json').write_text(json.dumps(result,indent=2)+'\n')
 print('Image audit '+('passed' if passed else 'FAILED'))
 return passed

if __name__=='__main__':
 p=argparse.ArgumentParser();p.add_argument('--release',type=pathlib.Path,required=True);p.add_argument('--output',type=pathlib.Path,required=True);p.add_argument('--trivy',default=str(pathlib.Path(os.environ.get('GATI_TOOLS_DIR',str(pathlib.Path.home()/'.local/share/gati/tools')))/'trivy-0.74.0/trivy'));a=p.parse_args()
 raise SystemExit(0 if audit(a.release.resolve(),a.output.resolve(),a.trivy) else 1)
