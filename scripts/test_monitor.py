import pathlib,tempfile,unittest,json
from unittest.mock import patch
from monitor import Collector,alerts,snapshot
class MonitoringTests(unittest.TestCase):
 def test_stale_and_error_alerts(self):
  v={'workers':{'matcher':{'state':'running','last_success_seconds':40,'last_start_seconds':35},'cleanup':{'state':'error','last_success_seconds':5,'last_start_seconds':0}},'failures':'suppressed','p95_upper_ms':300}
  self.assertEqual(alerts(v),['matcher_stale','cleanup_error'])
 def test_monitor_failure_bounded_and_no_exception_logging(self):
  with tempfile.TemporaryDirectory() as d:
   out=pathlib.Path(d)/'monitor.json';c=Collector('/absent',out,interval=60,retention=120)
   with patch('monitor.snapshot',side_effect=OSError('Bearer synthetic-private-canary')):
    for _ in range(10):c.sample()
   c.save();value=json.loads(out.read_text());self.assertEqual(len(value['samples']),2);self.assertNotIn('canary',out.read_text());self.assertEqual(out.stat().st_mode&0o777,0o600)
 def test_storage_failure_does_not_escape(self):
  c=Collector('/absent','/nonexistent-gati-monitor/out.json');c.sample();c.save()
 def test_bounds(self):
  for args in [{'interval':0},{'retention':0},{'interval':1,'retention':86401}]:
   with self.assertRaises(ValueError):Collector('/absent','/absent',**args)
