package simulation

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func WritePopulationReport(directory, roadsPath string, r PopulationReport) error {
	if err := os.MkdirAll(directory, 0700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(directory, "report.json"), append(raw, '\n'), 0600); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(directory, "timeline.csv"))
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	w.Write([]string{"simulation_minutes", "active_credentials", "going_credentials", "fresh_arrivals", "open_gatherings_observed"})
	for _, frame := range r.Frames {
		active := 0
		for _, n := range frame.Cells {
			active += n
		}
		w.Write([]string{fmt.Sprint(float64(frame.At-r.StartedAt) / 60000), strconv.Itoa(active), strconv.Itoa(frame.Going), strconv.Itoa(frame.Here), strconv.Itoa(len(frame.Gatherings))})
	}
	w.Flush()
	err = w.Error()
	f.Close()
	if err != nil {
		return err
	}
	if r.Grid.SizeMeters == 0 {
		return fmt.Errorf("no valid simulation geography to render")
	}
	var roads struct {
		Features []struct {
			Geometry struct {
				Coordinates [][2]float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"features"`
	}
	data, err := os.ReadFile(roadsPath)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &roads); err != nil {
		return err
	}
	var drawing strings.Builder
	for _, feature := range roads.Features {
		drawing.WriteString(`<path d="`)
		for i, p := range feature.Geometry.Coordinates {
			command := "L"
			if i == 0 {
				command = "M"
			}
			fmt.Fprintf(&drawing, "%s%.1f %.1f", command, (p[0]-r.Grid.West)/(r.Grid.East-r.Grid.West)*900, 700-(p[1]-r.Grid.South)/(r.Grid.North-r.Grid.South)*700)
		}
		drawing.WriteString(`"/>`)
	}
	page := `<!doctype html><html lang="sq"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><meta name="referrer" content="no-referrer"><title>GATI — Simulim i Tiranës</title>
<style>body{font:16px system-ui;color:#352c31;background:#fff8f3;max-width:1100px;margin:auto;padding:24px}h1{margin-bottom:8px}strong{color:#963754}.controls{display:flex;gap:16px;align-items:center;flex-wrap:wrap}button,select{font:inherit;padding:10px;border:1px solid #963754;border-radius:8px;background:white}input{flex:1;min-width:200px}svg{width:100%;background:#eef0e8;border-radius:16px;margin-top:16px}#roads path{fill:none;stroke:#bcc5bb;stroke-width:.7}.stats{display:flex;gap:16px;flex-wrap:wrap;margin:16px 0}.stats p{background:#fff;border:1px solid #ded5d5;padding:12px;border-radius:8px;margin:0}table{width:100%;border-collapse:collapse;margin:12px 0}th,td{text-align:left;padding:8px;border-bottom:1px solid #ddd}code{overflow-wrap:anywhere}footer{font-size:13px}#clock{min-width:100px}</style>
<h1>🦩 SIMULIM — Tiranë</h1><strong>Vetëm persona sintetikë. Kjo nuk është hartë e pjesëmarrjes reale.</strong>
<p id="description"></p><p>Rrathët: gatishmëri aktive. Katrorët: kryqëzime takimi. Ngjyra jeshile: JEMI KËTU sipas përgjigjes së fundit të vërejtur. Numrat përfaqësojnë kredenciale, jo njerëz të verifikuar.</p>
<div class="controls"><button id="play">Luaj</button><label for="timeline">Koha</label><input id="timeline" type="range" min="0" value="0"><output id="clock"></output></div>
<div id="stats" class="stats"></div><svg viewBox="0 0 900 700" role="img" aria-label="Rishikim i simulimit në hartën e Tiranës"><g id="roads">{{.Roads}}</g><g id="activity"></g></svg>
<h2>Rezultati i gjithë provës</h2><div id="funnel" class="stats"></div><table><thead><tr><th>Rrezja</th><th>Pranuar</th><th>Me ftesë</th></tr></thead><tbody id="radii"></tbody></table>
<p>Udhëtimet janë vlerësime nga distanca gjeografike. Zbulimi i takimeve për bashkim të drejtpërdrejtë injektohet nga simulimi; njoftimet publike ende nuk janë zbatuar. Afatet përdorin orë virtuale; kufijtë kundër abuzimit përdorin kohë reale.</p>
<footer>© OpenStreetMap · ODbL. Pa shërbime të jashtme. <p id="provenance"></p></footer>
<script id="data" type="application/json">{{.Data}}</script><script>
const r=JSON.parse(document.getElementById('data').textContent),el=id=>document.getElementById(id),ns='http://www.w3.org/2000/svg';
const point=p=>[(p[0]-r.grid.west)/(r.grid.east-r.grid.west)*900,700-(p[1]-r.grid.south)/(r.grid.north-r.grid.south)*700];
function node(type,attrs,parent,text){const n=document.createElementNS(ns,type);Object.entries(attrs).forEach(([k,v])=>n.setAttribute(k,v));if(text!==undefined)n.textContent=text;parent.appendChild(n);return n}
function cards(id,pairs){el(id).replaceChildren();for(const [label,n] of pairs){const p=document.createElement('p');p.textContent=label+': '+n;el(id).appendChild(p)}}
el('description').textContent=r.scenario.population+' persona · Fara '+r.scenario.seed+' · '+r.config_sha256.slice(0,12)+' · Pragjet '+r.effective_config.matching.activation_count+' / '+r.effective_config.arrivals.confirmation_count+' · Qeliza '+r.grid.size_meters+' m';
el('provenance').textContent='Konfigurimi: '+r.config_sha256+' · Harta: '+r.dataset+' · Hyrjet: '+r.input_sha256+' · Kohë reale: '+r.wall_seconds.toFixed(1)+' s';
const c=r.counts;
if(r.status!=='completed'){const warning=document.createElement('p');warning.setAttribute('role','alert');warning.textContent='SIMULIM I PAPËRFUNDUAR — Rezultatet janë të pjesshme dhe nuk kalojnë kontrollin.';el('description').before(warning)}
cards('funnel',[['Pranuar',c.credentials_accepted||0],['Refuzuar',c.credentials_rejected||0],['Me ftesë',c.credentials_invited||0],['Po shkoj (aktorë)',c.credentials_went||0],['Jo tani',c.decline_responses_accepted||0],['Pa përgjigje',c.ignored_invitations||0],['Mbërritur (aktorë)',c.credentials_arrived||0],['Konfirmime',c.arrivals_accepted||0],['Takime',c.gatherings_observed||0],['JEMI KËTU',c.gatherings_confirmed_observed||0],['Kërkesa të kufizuara',c.rate_limited_requests||0]]);
Object.entries(r.by_radius).sort((a,b)=>Number(a[0])-Number(b[0])).forEach(([radius,v])=>{const tr=document.createElement('tr');for(const t of [radius+' km',v.accepted||0,v.invited||0]){const td=document.createElement('td');td.textContent=t;tr.appendChild(td)}el('radii').appendChild(tr)});
el('timeline').max=Math.max(0,r.frames.length-1);
function show(){const f=r.frames[Number(el('timeline').value)];if(!f)return;el('clock').textContent=Math.round((f.at-r.started_at)/60000)+' minuta';el('activity').replaceChildren();let active=0;
for(const [cell,count] of Object.entries(f.cells)){active+=count;const parts=cell.split(':');const xy=point([r.grid.west+(Number(parts[2])+.5)*r.grid.lon_step,r.grid.south+(Number(parts[3])+.5)*r.grid.lat_step]);const g=node('g',{},el('activity'));node('circle',{cx:xy[0],cy:xy[1],r:r.grid.size_meters<500?5:14,fill:'#963754',opacity:.65},g);node('title',{},g,cell+': '+count);if(r.grid.size_meters>=500)node('text',{x:xy[0],y:xy[1]+4,'text-anchor':'middle',fill:'white','font-size':10},g,count)}
for(const g of f.gatherings){const p=point(g.intersection.point);const group=node('g',{},el('activity'));node('rect',{x:p[0]-6,y:p[1]-6,width:12,height:12,fill:g.state==='jemi_ketu'?'#167246':'#dc9d17',stroke:'white'},group);node('title',{},group,g.alias+' · '+g.intersection.label+' · '+g.state)}
cards('stats',[['Aktivë',active],['Po shkojnë',f.going],['Këtu përkohësisht',f.here],['Takime të hapura',f.gatherings.length]])}
let timer;el('play').onclick=()=>{if(timer){clearInterval(timer);timer=null;el('play').textContent='Luaj';return}el('play').textContent='Ndalo';timer=setInterval(()=>{if(Number(el('timeline').value)>=r.frames.length-1){clearInterval(timer);timer=null;el('play').textContent='Luaj';return}el('timeline').value=Number(el('timeline').value)+1;show()},180)};el('timeline').oninput=show;show();
</script></html>`
	t, err := template.New("replay").Parse(page)
	if err != nil {
		return err
	}
	f, err = os.Create(filepath.Join(directory, "index.html"))
	if err != nil {
		return err
	}
	defer f.Close()
	// json.Marshal escapes '<', '>' and '&', so public map names cannot close script.
	return t.Execute(f, struct {
		Roads template.HTML
		Data  template.JS
	}{template.HTML(drawing.String()), template.JS(raw)})
}
