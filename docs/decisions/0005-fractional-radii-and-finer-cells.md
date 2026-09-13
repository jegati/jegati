# 0005 — Fractional travel radii and optional 100 m cells

Date: 2026-09-13. Grid-default choice superseded by decision 0006; the other bounds remain.

The user requested simulations using their current settings:
30 willing credentials, 20 arrivals, travel radii 0.1/0.5/1/3 km, and support for
cells down to 0.1 km if needed. This supersedes whole-kilometre radius and 500 m
minimum-grid restrictions. Schema 6 permits finite increasing radii in 0.1 km
increments (0.1–20 km) and nominal cells of 100–5,000 m.

Radius and grid resolution remain different inputs. The user's current grid stays
1,000 m; the simulator must report that its 100/500 m travel choices have no reachable
landmark under the whole-cell conservative bound. Do not replace the current
configuration to manufacture matches. An explicitly labeled comparison may use
100 m cells with a tighter device accuracy limit (at most 50 m for that grid).

The nearest shared mapped intersection rule and full-cell conservative bound stay
unchanged. Even a 100 m cell does not make every 100 m journey eligible, particularly
near boundaries. Smaller cells expose a more precise participant area and increase
index computation/memory. They do not authenticate position. Original inference
counterexamples remain historical evidence with their map/config hashes.

Regression tests use a checked-in config fixture instead of the developer's mutable
operational file. Simulations record their full effective config/hash, distinguishing
profile-only conversion from any comparison changes. No new manual device-location
input, identity field, persistent record or public participant interface is added.
