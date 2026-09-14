"""Build a print-first Albanian pitch from completed synthetic study evidence."""

import argparse
import hashlib
import html
import importlib.util
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
spec = importlib.util.spec_from_file_location(
    "day_report", ROOT / "scripts/render-daystudy.py"
)
day_report = importlib.util.module_from_spec(spec)
spec.loader.exec_module(day_report)

CASES = [
    (
        "10000-3pct-clustered-reaction10",
        "Nis nga afërsia.",
        "Një audiencë më e vogël, e përqendruar në tri zona, mund të mbështesë takime në kohë të ndryshme gjatë ditës.",
        "Tri grumbullime sintetike",
        "Këtu vlen afërsia mes njerëzve. Modeli nis me rreth 300 persona GATI në një çast, brenda tri grumbullimeve të supozuara. Me bashkimet nga harta, aktiviteti vazhdon në disa zona gjatë pjesës më të madhe të ditës.",
        "Kjo është një mundësi për zona me dendësi lokale; nuk tregon mbulim të gjithë Tiranës.",
    ),
    (
        "30000-5pct-reaction10",
        "Një ditë në shumë zona.",
        "Me një audiencë më të gjerë, gatishmëria dhe bashkimet e reja mund të shpërndajnë mundësitë për takim në qytet.",
        "Shpërndarje uniforme",
        "Rreth 1 500 persona GATI në një çast përbëjnë bazën. Edhe me 10% bashkime ndër shikuesit e përshtatshëm të hartës, modeli prodhon aktivitet të konfirmuar në shumë qeliza përgjatë ditës.",
        "Shpërndarja uniforme është hipotetike: përfshin gjithë zonën e shërbimit, jo dendësinë reale të banorëve.",
    ),
    (
        "30000-5pct",
        "Kur bashkohen më shumë.",
        "E njëjta audiencë dhe gatishmëri bazë; më shumë shikues vendosin të shkojnë kur shohin aktivitet pranë tyre.",
        "Shpërndarje uniforme",
        "Në këtë provë, mundësia e bashkimit nga harta rritet nga 10% në 25%. Aktiviteti i dukshëm tërheq më shumë sesione në model dhe ruan praninë e konfirmuar në shumë zona.",
        "Rezultati përfshin shumë bashkime shtesë. Nuk mbahet vetëm nga 5% e audiencës: mesatarja totale aktive këtu arrin rreth 15%.",
    ),
]


def number(value, digits=0):
    return f"{value:,.{digits}f}".replace(",", " ").replace(".", ",")


def extent(values, digits=1):
    return f"{number(min(values), digits)}–{number(max(values), digits)}"


def shade(minutes):
    return ["#c5ddcc", "#83b89e", "#43886a", "#175840"][
        sum(minutes >= n for n in (60, 180, 360))
    ]


def map_svg(run, grid, roads):
    totals = {}
    for gathering in run["gatherings"]:
        if gathering["confirmed_minutes"] > 0:
            cell = gathering["cell"]
            totals[cell] = totals.get(cell, 0) + gathering["confirmed_minutes"]
    if len(totals) != run["metrics"]["confirmed_cells"]:
        raise ValueError("map cells disagree with reported coverage")
    rectangles = []
    for cell, minutes in sorted(totals.items()):
        _, size, x, y = cell.split(":")
        ratio = int(size) / grid["size_meters"]
        width = ratio * grid["lon_step"] / (grid["east"] - grid["west"]) * 900
        height = ratio * grid["lat_step"] / (grid["north"] - grid["south"]) * 800
        rectangles.append(
            f'<rect x="{int(x) * width:.2f}" y="{800 - (int(y) + 1) * height:.2f}" width="{width:.2f}" height="{height:.2f}" fill="{shade(minutes)}" fill-opacity=".82" stroke="white" stroke-width="1.3"/>'
        )
    return f"""<svg class="map" viewBox="0 0 900 800" role="img" aria-label="Takimet sintetike të konfirmuara në qeliza të Tiranës">
    <rect width="900" height="800" fill="#eef0e7"/>
    <g fill="none" stroke="#b5c1b1" stroke-width=".75">{roads}</g>
    {"".join(rectangles)}
    <rect x="25" y="20" width="133" height="45" rx="8" fill="#fff" opacity=".95"/>
    <text x="44" y="50" font-size="21" fill="#284539">TIRANË</text>
    <path d="M852 85V35m-9 13 9-13 9 13" fill="none" stroke="#284539" stroke-width="2.5"/>
    <text x="830" y="111" font-size="18" fill="#284539">Veri</text></svg>"""


def chart_svg(run):
    frames = run["frames"]
    maximum = max(1, max(frame["confirmed"] for frame in frames))
    points = [
        (45 + f["minute"] / 720 * 830, 105 - f["confirmed"] / maximum * 78)
        for f in frames
    ]
    path = f"M{points[0][0]:.1f},{points[0][1]:.1f}" + "".join(
        f"H{x:.1f}V{y:.1f}" for x, y in points[1:]
    )
    ticks = "".join(
        f'<text x="{45 + h / 12 * 830:.1f}" y="129" text-anchor="middle">{8 + h:02d}:00</text>'
        for h in range(0, 13, 2)
    )
    return f'''<svg class="chart" viewBox="0 0 900 138" role="img" aria-label="Numri i takimeve me JEMI KËTU përgjatë ditës">
    <path d="M45 27H875M45 105H875" stroke="#dce4db" fill="none"/>
    <path d="{path}" stroke="#287855" stroke-width="2.5" fill="none"/>
    <g fill="#607367" font-size="14">{ticks}<text x="12" y="33">{maximum}</text><text x="20" y="110">0</text></g></svg>'''


def footer(page, text="GATI 🦩 · Bëje vullnetin tënd të dukshëm."):
    return f"<footer><span>{text}</span><span>{page:02d} / 05</span></footer>"


def scenario_page(case, results, grid, roads, page):
    name, title, intro, distribution, interpretation, qualification = case
    runs = [r for r in results if r["scenario"]["name"] == name]
    if sorted(r["seed"] for r in runs) != [42, 43, 44]:
        raise ValueError("three predefined seeds required for each featured scenario")
    run = next(r for r in runs if r["seed"] == 42)
    s, m = run["scenario"], run["metrics"]
    hours = [r["metrics"]["minutes_any_confirmed"] / 60 for r in runs]
    cells = [r["metrics"]["confirmed_cells"] for r in runs]
    mean_active = [r["metrics"]["mean_total_active"] for r in runs]
    return f"""<section class="page scenario">
    <div class="eyebrow">SIMULIM · SKENARI {page - 1} · 08:00–20:00</div>
    <h2>{title}</h2><p class="deck">{intro}</p>
    <div class="parameters"><b>{number(s["population"])} persona potencialë</b><span>{number(s["willing_fraction"] * 100)}% GATI bazë</span><span>{number(s["reactive_fraction"] * 100)}% bashkime nga harta*</span></div>
    <div class="metrics">
    <div><strong>{number(m["minutes_any_confirmed"] / 60, 1)} <small>orë</small></strong><span>me ≥1 takim të konfirmuar</span><em>Tri prova: {extent(hours)} orë</em></div>
    <div><strong>{number(m["confirmed_cells"])} <small>qeliza</small></strong><span>me konfirmime gjatë ditës</span><em>Tri prova: {extent(cells, 0)} qeliza</em></div>
    <div><strong>{number(m["mean_total_active"])}</strong><span>sesione aktive mesatarisht</span><em>Tri prova: {extent(mean_active, 0)}</em></div>
    </div>
    <p class="micro">Shifrat kryesore, harta dhe grafiku: fara 42. Intervalet: farat 42, 43, 44. Sesionet aktive përfshijnë bashkimet nga harta, jo vetëm gatishmërinë bazë.</p>
    <div class="map-row"><div class="map-column">{map_svg(run, grid, roads)}<div class="attribution"><a href="https://www.openstreetmap.org/copyright">© Kontribuesit e OpenStreetMap · ODbL</a> · Qeliza 1 km × 1 km</div></div>
    <aside><h3>{distribution}</h3><p>{interpretation}</p><div class="legend"><b>Minuta-takime të konfirmuara</b><div><i style="background:#c5ddcc"></i> Më pak se 60</div><div><i style="background:#83b89e"></i> 60–180</div><div><i style="background:#43886a"></i> 180–360</div><div><i style="background:#175840"></i> 360 e më shumë</div></div><p class="micro">Shuma për qelizë gjatë ditës. Ngjyra nuk është numër njerëzish apo një takim i vetëm që zgjat aq kohë.</p></aside></div>
    <div class="chart-title">Takime me JEMI KËTU gjatë ditës <span>Vëzhgim i modelit çdo 5 minuta</span></div>{chart_svg(run)}
    <p class="takeaway">{qualification}</p>
    <p class="micro">* Një kontroll harte/orë. Përqindja zbatohet ndër shikuesit joaktivë pranë një takimi të publikuar e të arritshëm, që vendosin se kanë kohë të bashkohen. Rreze 1 km; gatishmëri 60 min. Simulim, jo parashikim. Hollësitë në faqen 5.</p>
    {footer(page)}</section>"""


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("output", type=Path)
    parser.add_argument(
        "--main", type=Path, default=ROOT / "reports/local/tirana-day-final.json"
    )
    parser.add_argument(
        "--reaction",
        type=Path,
        default=ROOT / "reports/local/tirana-reaction-sensitivity.json",
    )
    args = parser.parse_args()
    reports = [json.loads(path.read_text()) for path in (args.main, args.reaction)]
    for report in reports:
        day_report.summarize(report)
    if any(
        reports[0][k] != reports[1][k]
        for k in ("config_sha256", "map_sha256", "source_revision")
    ):
        raise ValueError("evidence sources differ")
    # The copy describes the reviewed study, not arbitrary future assumptions.
    results = reports[0]["results"] + reports[1]["results"]
    common = dict(
        hours=12,
        step_seconds=30,
        availability_minutes=60,
        radius_km=1,
        map_check_minutes=60,
        accept_fraction=0.8,
        arrive_fraction=0.85,
        renewal_fraction=0.8,
        dwell_minutes=30,
        walk_meters_second=1.2,
        detour=1.3,
    )
    for result in results:
        if result["scenario"]["name"] in {c[0] for c in CASES} and any(
            result["scenario"][k] != v for k, v in common.items()
        ):
            raise ValueError("review pitch copy before changing study assumptions")
    expected = json.loads((ROOT / "docs/outreach/tirana-day-summary.json").read_text())
    if reports[0]["config_sha256"] != expected["config_sha256"]:
        raise ValueError("reviewed config mismatch")
    reviewed = {row["name"]: row["scenario"] for row in expected["scenarios"]}
    for result in results:
        name = result["scenario"]["name"]
        if name in {c[0] for c in CASES} and result["scenario"] != reviewed[name]:
            raise ValueError("featured assumptions differ from reviewed evidence")
    roads = day_report.road_paths(reports[0]["grid"])
    pages = "".join(
        scenario_page(case, results, reports[0]["grid"], roads, i + 2)
        for i, case in enumerate(CASES)
    )
    template = (ROOT / "docs/outreach/pitch-layout.html").read_text()
    provenance = f"Konfigurimi {reports[0]['config_sha256'][:16]}… · Modeli {reports[0]['source_revision'].strip()[:7]} · 14.09.2026"
    rendered = (
        template.replace("__SCENARIOS__", pages)
        .replace("__PROVENANCE__", html.escape(provenance))
        .replace("__FOOTER1__", footer(1))
        .replace("__FOOTER5__", footer(5))
    )
    args.output.mkdir(parents=True, exist_ok=False)
    (args.output / "pitch.html").write_text(rendered)
    manifest = {
        "kind": "gati-albanian-pitch-v1",
        "featured": [c[0] for c in CASES],
        "illustration_seed": 42,
        "source_reports": [
            {
                "sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
                "study_sha256": report["study_sha256"],
                "config_sha256": report["config_sha256"],
            }
            for path, report in zip((args.main, args.reaction), reports)
        ],
        "html_sha256": hashlib.sha256(rendered.encode()).hexdigest(),
    }
    (args.output / "manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(args.output / "pitch.html")


if __name__ == "__main__":
    main()
