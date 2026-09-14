import collections
import unittest
from unittest.mock import Mock, patch

from lifecycle_run import Exercise, run


class LifecycleRunTests(unittest.TestCase):
    def exercise(self):
        exercise = Exercise.__new__(Exercise)
        exercise.lab = Mock()
        exercise.counts = collections.Counter()
        exercise.grid = {
            "size_meters": 100,
            "west": 19,
            "south": 41,
            "lon_step": 0.001,
            "lat_step": 0.001,
        }
        exercise.request = Mock()
        return exercise

    def test_logical_expiry_does_not_mask_retained_keys(self):
        exercise = self.exercise()
        person = {
            "cap": "synthetic",
            "hash": "synthetic-hash",
            "expires": 1000,
            "verified": False,
            "body": {"cell": "synthetic-cell"},
        }
        exercise.people = [person]
        exercise.lab.store.return_value = 1
        with patch("lifecycle_run.time.time", return_value=100):
            with self.assertRaisesRegex(ValueError, "keys survived"):
                exercise.tick()
        self.assertFalse(person["verified"])
        exercise.lab.store.side_effect = [0, None, "1000"]
        with patch("lifecycle_run.time.time", return_value=100):
            with self.assertRaisesRegex(ValueError, "cell index"):
                exercise.tick()
        self.assertFalse(person["verified"])
        exercise.lab.store.side_effect = [0, None, None]
        with patch("lifecycle_run.time.time", return_value=100):
            exercise.tick()
        self.assertTrue(person["verified"])

    def test_replay_cannot_hide_freshness_extension(self):
        exercise = self.exercise()
        exercise.request.side_effect = [
            {},
            {"arrival_until": 2000},
            {"arrival_until": 2001},
        ]
        with self.assertRaisesRegex(ValueError, "replay changed"):
            exercise.arrive(
                {"expires": 10000},
                {"intersection": {"point": [19.5, 41.5]}, "ends_at": 10000},
            )

    def test_short_run_cannot_claim_overnight_coverage(self):
        with patch("lifecycle_run.Lab") as lab:
            for hours in [0, 1, 7, 25]:
                with self.assertRaises(ValueError):
                    run("unused", hours)
            lab.assert_not_called()

    def test_invited_state_commits_before_arriving(self):
        exercise = self.exercise()
        exercise.gatherings = {}
        exercise.late = set()
        exercise.arrive = Mock()
        person = {"expires": 500000, "verified": False, "gone": False, "role": "attend"}
        invitation = {
            "id": "synthetic",
            "ends_at": 400000,
            "intersection": {"id": "synthetic"},
            "state": "jemi_gati",
        }
        exercise.people = [person]
        exercise.request.side_effect = [
            {"expires_at": 500000, "state": "invited", "invitation": invitation},
            {"expires_at": 500000, "state": "going", "invitation": invitation},
        ]
        with patch("lifecycle_run.time.time", return_value=100):
            exercise.tick()
        self.assertEqual(exercise.request.call_args.args[1], "/api/going")
        exercise.arrive.assert_called_once_with(person, invitation)


if __name__ == "__main__":
    unittest.main()
