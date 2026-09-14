"""Evidence summaries must not hide failed or incompatible experiments."""

import importlib.util
import pathlib
import unittest

spec = importlib.util.spec_from_file_location(
    "daystudy", pathlib.Path(__file__).with_name("render-daystudy.py")
)
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class SummaryTests(unittest.TestCase):
    def report(self):
        return {
            "status": "completed",
            "results": [
                {
                    "scenario": {"name": "case", "population": 3000},
                    "seed": seed,
                    "metrics": {"gatherings_activated": n},
                }
                for seed, n in [(42, 0), (43, 2), (44, 8)]
            ],
        }

    def test_keeps_zero_runs_and_range(self):
        result = module.summarize(self.report())[0]
        self.assertEqual(result["seeds"], [42, 43, 44])
        self.assertEqual(
            result["metrics"]["gatherings_activated"],
            dict(median=2, minimum=0, maximum=8),
        )

    def test_refuses_partial_duplicate_and_mixed_assumptions(self):
        for mutate in [
            lambda r: r.update(status="failed"),
            lambda r: r["results"].append(r["results"][0]),
            lambda r: r["results"][1]["scenario"].update(population=10000),
        ]:
            report = self.report()
            mutate(report)
            with self.assertRaises(ValueError):
                module.summarize(report)


if __name__ == "__main__":
    unittest.main()
