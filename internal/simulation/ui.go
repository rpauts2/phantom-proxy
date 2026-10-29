package simulation

import (
	"net/http"
)

const dashboardHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Phantom Lab · Simulation Console</title>
  <link rel="stylesheet" href="/style.css">
  <script src="/app.js" defer></script>
</head>
<body>
  <main class="shell">
    <header class="topbar">
      <div class="brand"><span class="mark">PL</span><div><strong>Phantom Lab</strong><small>SIMULATION CONSOLE</small></div></div>
      <span id="health" class="pill">Checking status…</span>
    </header>
    <section class="hero">
      <p class="eyebrow">LOCAL TRAINING ENVIRONMENT</p>
      <h1>Practice safely.<br><span>Measure what matters.</span></h1>
      <p class="intro">A local-only console for synthetic awareness events. No passwords, cookies, or session tokens are collected.</p>
      <div class="hero-actions"><button id="refresh" class="button secondary">Refresh data</button><button id="stop" class="button danger">Emergency stop</button></div>
    </section>
    <section class="stats">
      <article class="stat"><span>LAB STATE</span><strong id="state">Loading</strong><small>Ingestion status</small></article>
      <article class="stat"><span>RECENT EVENTS</span><strong id="count">0</strong><small>Retained in memory</small></article>
      <article class="stat"><span>STORAGE</span><strong>Ephemeral</strong><small>Cleared on restart</small></article>
    </section>
    <section class="content">
      <article class="panel">
        <div class="panel-heading"><div><p class="eyebrow">TEST DATA</p><h2>Record an event</h2></div><span class="tag">Synthetic only</span></div>
        <form id="event-form">
          <label>Campaign ID<input name="campaign_id" value="demo" pattern="[A-Za-z0-9_-]{1,64}" maxlength="64" required></label>
          <label>Participant ID<input name="participant_id" value="person_01" pattern="[A-Za-z0-9_-]{1,64}" maxlength="64" required></label>
          <label>Event type<select name="type"><option value="visit">Visit</option><option value="click">Click</option><option value="training_complete">Training completed</option></select></label>
          <button class="button primary" type="submit">Add synthetic event <span aria-hidden="true">↗</span></button>
          <p id="message" class="message" role="status">Only campaign IDs in the startup allowlist are accepted.</p>
        </form>
      </article>
      <article class="panel">
        <div class="panel-heading"><div><p class="eyebrow">LIVE HISTORY</p><h2>Recent events</h2></div><span class="tag">Newest first</span></div>
        <div class="table-wrap"><table><thead><tr><th>Event</th><th>Campaign</th><th>Participant</th><th>Time (UTC)</th></tr></thead><tbody id="events"><tr><td colspan="4" class="empty">Loading events…</td></tr></tbody></table></div>
      </article>
    </section>
    <footer><span>PHANTOM LAB / GO RUNTIME</span><span>Loopback only · Synthetic data only</span></footer>
  </main>
</body>
</html>`

const dashboardCSS = `:root{color-scheme:dark;--bg:#0a0d14;--panel:#111722;--line:#263144;--muted:#9aa8bc;--text:#edf3ff;--blue:#83a9ff;--green:#82e0bd;--red:#ff8d98}*{box-sizing:border-box}body{margin:0;background:radial-gradient(ellipse at 20% 0%,#17233b 0,transparent 44%),var(--bg);color:var(--text);font:15px/1.5 Inter,ui-sans-serif,system-ui,-apple-system,"Segoe UI",sans-serif}.shell{max-width:1180px;margin:auto;padding:28px 24px 20px}.topbar,.brand,.hero-actions,.panel-heading,footer{display:flex;align-items:center;justify-content:space-between;gap:16px}.brand{justify-content:flex-start}.mark{display:grid;place-items:center;width:42px;height:42px;border-radius:13px;background:linear-gradient(135deg,#8eafff,#6f83f5);color:#101526;font-weight:900;letter-spacing:-1px}.brand strong,.brand small{display:block}.brand strong{font-size:16px;letter-spacing:.2px}.brand small,.eyebrow,.stat>span,footer{font-size:10px;letter-spacing:1.7px;color:var(--muted)}.pill,.tag{border:1px solid var(--line);border-radius:99px;padding:7px 12px;color:var(--muted);font-size:12px}.pill.ok{color:var(--green);border-color:#2a5e50}.pill.stopped{color:var(--red);border-color:#633641}.hero{padding:64px 0 34px;max-width:730px}.eyebrow{margin:0 0 10px;color:var(--blue);font-weight:700}.hero h1{font-size:clamp(38px,6vw,64px);line-height:1.04;letter-spacing:-2.5px;margin:0 0 18px}.hero h1 span{color:var(--blue)}.intro{max-width:620px;color:var(--muted);font-size:17px}.hero-actions{justify-content:flex-start;margin-top:24px}.button{border:0;border-radius:10px;padding:12px 16px;font:inherit;font-weight:700;cursor:pointer;transition:transform .15s,opacity .15s}.button:hover{transform:translateY(-1px)}.button:disabled{opacity:.45;cursor:not-allowed;transform:none}.primary{background:var(--blue);color:#0a1020}.secondary{background:#1a2435;color:var(--text);border:1px solid var(--line)}.danger{background:#3a1e29;color:#ffb1ba;border:1px solid #67313e}.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:14px;margin-bottom:18px}.stat,.panel{background:linear-gradient(145deg,#131b29,#0f141e);border:1px solid var(--line);border-radius:16px}.stat{padding:20px}.stat>span,.stat small{display:block;color:var(--muted)}.stat strong{display:block;font-size:25px;margin:7px 0 1px;letter-spacing:-.5px}.stat small{font-size:12px}.content{display:grid;grid-template-columns:minmax(280px,.8fr) minmax(0,1.2fr);gap:16px}.panel{padding:22px;min-width:0}.panel-heading{margin-bottom:22px}.panel-heading h2{font-size:20px;letter-spacing:-.5px;margin:0}.panel-heading .eyebrow{margin-bottom:4px}.tag{white-space:nowrap;padding:5px 9px}form{display:grid;gap:14px}label{display:grid;gap:7px;color:#c5d0e1;font-size:13px}input,select{width:100%;background:#0b1019;border:1px solid #303c50;border-radius:9px;padding:12px;color:var(--text);font:inherit;outline:none}input:focus,select:focus{border-color:var(--blue);box-shadow:0 0 0 3px #83a9ff1c}.message{font-size:12px;color:var(--muted);margin:0}.message.error{color:var(--red)}.message.success{color:var(--green)}.table-wrap{overflow-x:auto}table{width:100%;border-collapse:collapse;text-align:left;font-size:12px}th{font-weight:600;color:var(--muted);padding:0 10px 12px;border-bottom:1px solid var(--line)}td{padding:13px 10px;border-bottom:1px solid #202a3a;white-space:nowrap}tbody tr:last-child td{border-bottom:0}.event-type{display:inline-block;border-radius:6px;padding:4px 7px;background:#1b2940;color:#b8ceff}.empty{text-align:center;color:var(--muted);padding:28px 10px}footer{padding:26px 2px 0;letter-spacing:1px;font-size:9px}@media(max-width:850px){.content{grid-template-columns:1fr}.hero{padding-top:46px}}@media(max-width:540px){.shell{padding:18px 14px}.stats{grid-template-columns:1fr}.stat{padding:14px}.hero h1{letter-spacing:-1.5px}.panel{padding:16px}.tag{display:none}footer{align-items:flex-start;flex-direction:column;gap:7px}}`

const dashboardJS = `const $=s=>document.querySelector(s);const form=$("#event-form");const message=$("#message");const stopButton=$("#stop");const safeText=v=>String(v??"");async function request(path,options){const response=await fetch(path,options);const body=await response.json().catch(()=>({error:"Unexpected server response"}));if(!response.ok)throw new Error(body.error||"Request failed");return body}function say(text,kind=""){message.textContent=text;message.className="message "+kind}function cell(text){const td=document.createElement("td");td.textContent=safeText(text);return td}async function refresh(){try{const [status,data]=await Promise.all([request("/api/v1/status"),request("/api/v1/events?limit=100")]);$("#state").textContent=status.stopped?"Stopped":status.scope_expired?"Scope expired":"Running";$("#count").textContent=String(status.events);const halted=status.stopped||status.scope_expired;$("#health").textContent=status.stopped?"Emergency stop active":status.scope_expired?"Scope expired":"Lab operational";$("#health").className="pill "+(halted?"stopped":"ok");stopButton.disabled=halted;form.querySelector("button[type=submit]").disabled=halted;const body=$("#events");body.replaceChildren();if(!data.events.length){const tr=document.createElement("tr");const td=cell("No events yet");td.colSpan=4;td.className="empty";tr.append(td);body.append(tr)}for(const event of data.events){const tr=document.createElement("tr");const type=cell(event.type);type.innerHTML="";const badge=document.createElement("span");badge.className="event-type";badge.textContent=safeText(event.type);type.append(badge);tr.append(type,cell(event.campaign_id),cell(event.participant_id),cell(new Date(event.occurred_at).toLocaleTimeString("en-GB",{timeZone:"UTC",hour12:false})));body.append(tr)}}catch(error){$("#health").textContent="API unavailable";$("#health").className="pill stopped";say(error.message,"error")}}form.addEventListener("submit",async event=>{event.preventDefault();const payload=Object.fromEntries(new FormData(form));try{await request("/api/v1/events",{method:"POST",headers:{"Content-Type":"application/json","X-Phantom-Lab-Request":"1"},body:JSON.stringify(payload)});say("Synthetic event recorded.","success");await refresh()}catch(error){say(error.message,"error")}});$("#refresh").addEventListener("click",refresh);stopButton.addEventListener("click",async()=>{if(!confirm("Stop accepting events until the server restarts?"))return;try{await request("/admin/stop",{method:"POST",headers:{"X-Phantom-Lab-Action":"stop"}});say("Emergency stop activated.","success");await refresh()}catch(error){say(error.message,"error")}});refresh();`

func serveDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(dashboardHTML))
}

func serveDashboardCSS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(dashboardCSS))
}

func serveDashboardJS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, http.MethodGet)
		return
	}
	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(dashboardJS))
}
