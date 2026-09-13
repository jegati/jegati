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
  with self.assertRaises(ValueError):Collector('/absent','/absent',container='--privileged')
 def test_container_reader_uses_fixed_private_command(self):
  import subprocess
  from monitor import ContainerHTTP
  with patch('monitor.subprocess.run',return_value=subprocess.CompletedProcess([],0,b'{}')) as run:
   client=ContainerHTTP('gati-api-1','/run/ops/gati.sock');response=client.getresponse()
   self.assertEqual(response.read(8193),b'{}')
   self.assertEqual(run.call_args.args[0],['docker','exec','gati-api-1','/gati','-mode','monitor-snapshot','-monitor-socket','/run/ops/gati.sock'])
 def test_unexpected_sensitive_fields_fail_closed(self):
  base={'version':1,'window_seconds':60,'retention_seconds':120,'minimum_samples':20,'observed_epoch':1,'requests':'suppressed','failures':'suppressed','p95_upper_ms':-1,'workers':{}}
  class Response:
   status=200
   def __init__(self,value):self.value=value
   def read(self,_):return json.dumps(self.value).encode()
  for value in [{**base,'token':'private-canary'},{**base,'workers':{'participant':{}}},[]]:
   with patch('monitor.UnixHTTP') as client:
    client.return_value.getresponse.return_value=Response(value)
    with self.assertRaises(ValueError):snapshot('/synthetic')
