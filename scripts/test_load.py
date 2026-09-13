import importlib.util,pathlib,unittest
spec=importlib.util.spec_from_file_location('check_load',pathlib.Path(__file__).with_name('check-load.py'));module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
class LoadGateTests(unittest.TestCase):
 def report(self):return {'synthetic':True,'real_time':True,'stages':[{'accepted_responses':300,'attempted_population':300}], 'phases':[{'name':'create-300','statuses':{'200':300},'attempted':300,'generator_dropped':0,'seconds':1,'p95_upper_ms':10}]}
 def test_passing_client_evidence(self):self.assertTrue(module.evaluate(self.report())['client_gate_passed'])
 def test_fast_rejections_cannot_hide_slow_acceptance(self):
  r=self.report();r['phases'][0].update(statuses={'200':300,'429':10000},attempted=10300,accepted_p95_upper_ms=1000)
  self.assertFalse(module.evaluate(r)['client_gate_passed'])
 def test_missing_split_histogram_drops_and_unaccepted_population_fail(self):
  for mutate in [lambda r:r['phases'][0].update(statuses={'200':299,'429':1}),lambda r:r['phases'][0].update(generator_dropped=1),lambda r:r['stages'][0].update(accepted_responses=299)]:
   r=self.report();mutate(r);self.assertFalse(module.evaluate(r)['client_gate_passed'])
