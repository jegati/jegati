import copy,hashlib,json,pathlib,tempfile,unittest
from unittest.mock import patch
from release import verify_files,activate,runtime_dir,verify_runtime_shape,verify_service_selection

class ReleaseTests(unittest.TestCase):
 def test_existing_acl_changes_are_rejected_without_rewrite(self):
  import subprocess
  script=pathlib.Path(__file__).with_name('init-secrets.py')
  with tempfile.TemporaryDirectory() as d:
   args=['python3',str(script),'--directory',d]
   subprocess.run(args,check=True,stdout=subprocess.DEVNULL)
   subprocess.run(args+['--check'],check=True,stdout=subprocess.DEVNULL)
   acl=pathlib.Path(d)/'users.acl';acl.chmod(0o600);acl.write_text('synthetic changed ACL')
   self.assertNotEqual(subprocess.run(args+['--check'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL).returncode,0)
   self.assertEqual(acl.read_text(),'synthetic changed ACL')
 def manifest(self,root):
  (root/'config').mkdir();(root/'config/gati.yaml').write_text('synthetic config')
  digest=hashlib.sha256((root/'config/gati.yaml').read_bytes()).hexdigest()
  value={'version':1,'config_sha256':digest,'files':{'config/gati.yaml':digest}}
  (root/'release.json').write_text(json.dumps(value));(root/'release.sha256').write_text(hashlib.sha256((root/'release.json').read_bytes()).hexdigest()+'  release.json\n')
  return value
 def test_tamper_fails_before_deployment(self):
  with tempfile.TemporaryDirectory() as d:
   root=pathlib.Path(d);self.manifest(root);verify_files(root)
   (root/'config/gati.yaml').write_text('modified')
   with self.assertRaises(ValueError):verify_files(root)
 def test_rollback_rejects_incompatible_config_before_mutation(self):
  with tempfile.TemporaryDirectory() as d:
   root=pathlib.Path(d);self.manifest(root)
   with patch('release.run') as command:
    with self.assertRaises(ValueError):activate(root,'gati-test',{'config_sha256':'different'},1,False,root)
    command.assert_not_called()
 def test_releases_share_operational_mounts(self):
  self.assertEqual(runtime_dir(pathlib.Path('/opt/gati/releases/a'),'gati-production'),runtime_dir(pathlib.Path('/opt/gati/releases/b'),'gati-production'))

class RuntimeTests(unittest.TestCase):
 def setUp(self):
  self.image={'Entrypoint':['/gati'],'Cmd':['-role','combined'],'User':'10001:10001','Env':['PATH=/bin']}
  self.service={'command':['-trusted-proxies','172.30.11.2/32'],'environment':{'GOMEMLIMIT':'1GiB'},'mem_limit':'1024','memswap_limit':'1024','pids_limit':256,'cpus':2,'volumes':[{'target':'/config/gati.yaml','type':'bind','source':'/release/config/gati.yaml','read_only':True}],'tmpfs':['/run/ops:mode=0700,size=1m'],'networks':{'application':{'ipv4_address':'172.30.11.2'}}}
  self.networks={'application':{'name':'test_application'}}
  self.value={'Config':{**self.image,'Cmd':self.service['command'],'Env':['PATH=/bin','GOMEMLIMIT=1GiB']},'HostConfig':{'Memory':1024,'MemorySwap':1024,'PidsLimit':256,'NanoCpus':2000000000,'Tmpfs':{'/run/ops':'mode=0700,size=1m'}},'Mounts':[{'Destination':'/config/gati.yaml','Type':'bind','Source':'/release/config/gati.yaml','RW':False}],'NetworkSettings':{'Networks':{'test_application':{'IPAddress':'172.30.11.2'}}}}
 def check(self,value):verify_runtime_shape(value,self.service,self.networks,self.image)
 def test_expected_launch_passes_with_image_defaults(self):self.check(self.value)
 def test_launch_drift_rejected_without_disclosing_values(self):
  for field,replacement in [('Cmd',['-trusted-proxies','0.0.0.0/0']),('Entrypoint',['/alternate']),('User','0'),('Env',['PRIVATE_TEST_SECRET=must-not-print']),('WorkingDir','/alternate')]:
   with self.subTest(field=field):
    value=copy.deepcopy(self.value);value['Config'][field]=replacement
    with self.assertRaises(ValueError) as failure:self.check(value)
    self.assertNotIn('must-not-print',str(failure.exception))
 def test_extra_or_replaced_read_only_mount_rejected(self):
  for mounts in [[{'Destination':'/etc/caddy/Caddyfile','Type':'bind','Source':'/alternate','RW':False}],[]]:
   value=copy.deepcopy(self.value);value['Mounts']=mounts
   with self.assertRaises(ValueError):self.check(value)
 def test_network_or_proxy_address_drift_rejected(self):
  for networks in [{},{'test_application':{'IPAddress':'172.30.11.3'}},{**self.value['NetworkSettings']['Networks'],'egress':{}}]:
   value=copy.deepcopy(self.value);value['NetworkSettings']['Networks']=networks
   with self.assertRaises(ValueError):self.check(value)
 def test_privilege_resource_and_tmpfs_drift_rejected(self):
  for field,replacement in [('Privileged',True),('CapAdd',['NET_ADMIN']),('PidMode','host'),('NetworkMode','host'),('Memory',2048),('MemorySwap',2048),('PidsLimit',0),('NanoCpus',0),('Tmpfs',{})]:
   value=copy.deepcopy(self.value);value['HostConfig'][field]=replacement
   with self.assertRaises(ValueError):self.check(value)
 def test_omitted_live_tunnel_or_worker_rejected(self):
  for target in ['worker','tunnel']:
   with patch('release.compose',side_effect=lambda *args,**kw:'container' if args[-1]==target else ''):
    with self.assertRaises(ValueError):verify_service_selection(pathlib.Path('/release'),'test',1,False)
 def test_selected_services_are_not_treated_as_orphans(self):
  with patch('release.compose') as command:
   verify_service_selection(pathlib.Path('/release'),'test',2,True);command.assert_not_called()

if __name__=='__main__':unittest.main()
