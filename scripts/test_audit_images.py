import datetime,importlib.util,pathlib,unittest
spec=importlib.util.spec_from_file_location('audit_images',pathlib.Path(__file__).with_name('audit-images.py'));audit=importlib.util.module_from_spec(spec);spec.loader.exec_module(audit)
class AuditTests(unittest.TestCase):
 def setUp(self):
  self.today=datetime.date(2026,9,14);self.item={'id':'TEST-1','package':'example','version':'v1','severity':'HIGH'}
  self.exception={'id':'TEST-1','package':'example','version':'v1','services':['web'],'reviewed':'2026-09-14','expires':'2026-10-14','reason':'unlinked function','owner':'GATI maintainer','evidence':'binary scan'}
 def test_exact_current_exception(self):
  accepted,unresolved=audit.classify('web',[self.item],[self.exception],self.today);self.assertEqual(len(accepted),1);self.assertFalse(unresolved)
 def test_changed_findings_are_not_suppressed(self):
  for field,value in [('id','TEST-2'),('package','other'),('version','v2')]:
   accepted,unresolved=audit.classify('web',[{**self.item,field:value}],[self.exception],self.today);self.assertFalse(accepted);self.assertEqual(len(unresolved),1)
 def test_wrong_service_expired_future_or_unexplained_exception_fails(self):
  for change in [{'services':['api']},{'expires':'2026-09-13'},{'reviewed':'2026-09-15'},{'reason':''},{'owner':''},{'evidence':''}]:
   accepted,unresolved=audit.classify('web',[self.item],[{**self.exception,**change}],self.today);self.assertFalse(accepted);self.assertEqual(len(unresolved),1)
 def test_unknown_severity_without_exception_fails(self):
  self.assertEqual(len(audit.classify('web',[{**self.item,'severity':'UNKNOWN'}],[],self.today)[1]),1)
if __name__=='__main__':unittest.main()
