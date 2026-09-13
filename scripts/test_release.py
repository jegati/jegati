import hashlib,json,pathlib,tempfile,unittest
from unittest.mock import patch
from release import verify_files,activate,runtime_dir

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

if __name__=='__main__':unittest.main()
