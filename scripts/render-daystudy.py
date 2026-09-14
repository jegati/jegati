"""Render offline synthetic evidence without external assets or plotting packages."""

import argparse
import json
import pathlib
import statistics

ROOT = pathlib.Path(__file__).resolve().parent.parent
LABELS = {
    "30000-5pct-reaction10": "30 000 · 5% · bashkime 10%",
    "10000-3pct-clustered-reaction10": "10 000 · 3% · grumbullime · bashkime 10%",
    "3000-1pct": "3 000 · 1% · 1 km",
    "3000-5pct": "3 000 · 5% · 1 km",
    "10000-3pct-no-reaction": "10 000 · 3% · pa bashkime nga harta",
    "10000-3pct": "10 000 · 3% · 1 km",
    "30000-3pct-no-reaction": "30 000 · 3% · pa bashkime nga harta",
    "30000-3pct": "30 000 · 3% · 1 km",
    "30000-5pct": "30 000 · 5% · 1 km",
    "10000-3pct-3km": "10 000 · 3% · 3 km",
    "10000-3pct-clustered": "10 000 · 3% · në tri grumbullime",
}
METRICS = [
    "mean_baseline_active",
    "mean_total_active",
    "gatherings_activated",
    "gatherings_ever_confirmed",
    "minutes_any_confirmed",
    "confirmed_gathering_minutes",
    "confirmed_cells",
    "peak_simultaneous_confirmed",
    "reactive_sessions",
]


def summarize(report):
    if report.get("status") != "completed" or not report.get("results"):
        raise ValueError("only completed studies can be rendered")
    grouped = {}
    for result in report["results"]:
        grouped.setdefault(result["scenario"]["name"], []).append(result)
    rows = []
    for name, runs in grouped.items():
        if len({r["seed"] for r in runs}) != len(runs):
            raise ValueError("duplicate seeds")
        if any(r["scenario"] != runs[0]["scenario"] for r in runs):
            raise ValueError("same label with different assumptions")
        metrics = {}
        for key in METRICS:
            values = [r["metrics"].get(key, 0) for r in runs]
            metrics[key] = dict(
                median=statistics.median(values),
                minimum=min(values),
                maximum=max(values),
            )
        rows.append(
            dict(
                name=name,
                scenario=runs[0]["scenario"],
                seeds=[r["seed"] for r in runs],
                metrics=metrics,
            )
        )
    return rows


def road_paths(grid):
    roads = json.loads((ROOT / "data/tirana/roads.geojson").read_text())
    paths = []
    for feature in roads["features"]:
        coords = feature["geometry"]["coordinates"]
        points = []
        for lon, lat in coords:
            xy = (
                round((lon - grid["west"]) / (grid["east"] - grid["west"]) * 900, 1),
                round(
                    800 - (lat - grid["south"]) / (grid["north"] - grid["south"]) * 800,
                    1,
                ),
            )
            if not points or points[-1] != xy:
                points.append(xy)
        if len(points) > 1:
            paths.append("M" + "L".join(f"{x},{y}" for x, y in points))
    return '<path d="' + " ".join(paths) + '"/>'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("report", type=pathlib.Path)
    parser.add_argument("output", type=pathlib.Path)
    parser.add_argument("--supplement", type=pathlib.Path, action="append", default=[])
    args = parser.parse_args()
    report = json.loads(args.report.read_text())
    source_keys = ("study_sha256", "source_revision", "source_dirty", "wall_seconds")
    report["source_studies"] = [{k: report[k] for k in source_keys}]
    for path in args.supplement:
        extra = json.loads(path.read_text())
        summarize(extra)
        if any(
            extra[k] != report[k]
            for k in ("kind", "config_sha256", "map_sha256", "source_revision")
        ):
            raise ValueError("supplement has incompatible implementation inputs")
        report["results"].extend(extra["results"])
        report["source_studies"].append({k: extra[k] for k in source_keys})
    rows = summarize(report)
    args.output.mkdir(parents=True, exist_ok=False)
    summary = {
        key: report[key]
        for key in (
            "kind",
            "config_sha256",
            "study_sha256",
            "map_sha256",
            "source_revision",
            "source_dirty",
            "wall_seconds",
        )
    }
    summary["scenarios"] = rows
    summary["source_studies"] = report["source_studies"]
    (args.output / "summary.json").write_text(json.dumps(summary, indent=2) + "\n")
    markdown = [
        "| Skenari | Mesatarisht GATI bazë | Takime të formuara | Arritën JEMI KËTU | Orë me ≥1 JEMI KËTU | Qeliza me JEMI KËTU |",
        "|---|---:|---:|---:|---:|---:|",
    ]
    for row in rows:

        def cell(key, divisor=1):
            m = row["metrics"][key]
            precision = 2 if divisor != 1 else 1
            return f"{m['median'] / divisor:.{precision}f} ({m['minimum'] / divisor:.{precision}f}–{m['maximum'] / divisor:.{precision}f})"

        markdown.append(
            "| "
            + " | ".join(
                [
                    LABELS.get(row["name"], row["name"]),
                    cell("mean_baseline_active"),
                    cell("gatherings_activated"),
                    cell("gatherings_ever_confirmed"),
                    cell("minutes_any_confirmed", 60),
                    cell("confirmed_cells"),
                ]
            )
            + " |"
        )
    (args.output / "table.md").write_text("\n".join(markdown) + "\n")
    template = (ROOT / "simulation/studies/day-report.html").read_text()
    data = (
        json.dumps(
            dict(report=report, summary=rows, labels=LABELS),
            separators=(",", ":"),
            ensure_ascii=False,
        )
        .replace("<", "\\u003c")
        .replace(">", "\\u003e")
        .replace("&", "\\u0026")
    )
    page = template.replace("__ROADS__", road_paths(report["grid"])).replace(
        "__DATA__", data
    )
    (args.output / "index.html").write_text(page)
    print(f"Standalone report: {args.output / 'index.html'}")


if __name__ == "__main__":
    main()
