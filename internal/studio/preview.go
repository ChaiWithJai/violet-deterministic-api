package studio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"strings"
)

type previewModel struct {
	JobID         string
	AppName       string
	Domain        string
	Plan          string
	Region        string
	Deployment    string
	PrimaryUsers  []string
	Workflows     []string
	Entities      []string
	Integrations  []string
	WorkloadCount int
	FileCount     int
	TerminalCount int
	ConsoleCount  int
}

type previewDocument struct {
	AppName    string
	Client     string
	ClientName string
	CSSURL     string
	JSURL      string
}

func (s *Service) RenderPreview(tenantID, jobID, client, token string) (string, bool) {
	job, ok := s.GetJob(tenantID, jobID)
	if !ok {
		return "", false
	}

	mode := runtimeClient(client)
	clientName := "Web"
	if mode == "mobile" {
		clientName = "Mobile"
	}

	assetQuery := url.Values{}
	assetQuery.Set("v", fmt.Sprintf("%d", job.UpdatedAt.UnixNano()))
	if strings.TrimSpace(token) != "" {
		assetQuery.Set("token", token)
	}
	query := assetQuery.Encode()
	base := fmt.Sprintf("/v1/studio/jobs/%s/runtime/%s", job.JobID, mode)
	doc := previewDocument{
		AppName:    fallback(job.Confirmation.AppName, "Generated App"),
		Client:     mode,
		ClientName: clientName,
		CSSURL:     base + "/app.css?" + query,
		JSURL:      base + "/app.js?" + query,
	}
	return executeTemplate(previewDocumentTemplate, doc), true
}

func (s *Service) RenderRuntimeAsset(tenantID, jobID, client, asset string) (string, []byte, bool) {
	job, ok := s.GetJob(tenantID, jobID)
	if !ok {
		return "", nil, false
	}
	mode := runtimeClient(client)
	name := strings.Trim(strings.TrimSpace(asset), "/")
	if name == "" {
		name = "index.html"
	}

	model := buildPreviewModel(job)
	if contentType, content, ok := lookupGeneratedRuntimeAsset(job.Files, mode, name); ok {
		return contentType, content, true
	}
	switch name {
	case "index.html":
		html, found := s.RenderPreview(tenantID, jobID, mode, "")
		if !found {
			return "", nil, false
		}
		return "text/html; charset=utf-8", []byte(html), true
	case "app.css":
		if mode == "mobile" {
			return "text/css; charset=utf-8", []byte(mobileRuntimeCSS), true
		}
		return "text/css; charset=utf-8", []byte(webRuntimeCSS), true
	case "app.js":
		if mode == "mobile" {
			return "application/javascript; charset=utf-8", []byte(mobileRuntimeJS(model)), true
		}
		return "application/javascript; charset=utf-8", []byte(webRuntimeJS(model)), true
	default:
		return "", nil, false
	}
}

func runtimeSourceArtifacts(slug string, conf Confirmation) []FileArtifact {
	base := fmt.Sprintf("apps/%s/clients", slug)
	model := previewModelFromConfirmation(conf)
	web := []FileArtifact{
		{Path: base + "/web/index.html", Language: "html", Content: standaloneRuntimeHTML(conf.AppName, "Web")},
		{Path: base + "/web/app.css", Language: "css", Content: webRuntimeCSS},
		{Path: base + "/web/app.js", Language: "javascript", Content: webRuntimeJS(model)},
	}
	mobile := []FileArtifact{
		{Path: base + "/mobile/index.html", Language: "html", Content: standaloneRuntimeHTML(conf.AppName, "Mobile")},
		{Path: base + "/mobile/app.css", Language: "css", Content: mobileRuntimeCSS},
		{Path: base + "/mobile/app.js", Language: "javascript", Content: mobileRuntimeJS(model)},
	}
	return append(web, mobile...)
}

func standaloneRuntimeHTML(appName, clientName string) string {
	doc := previewDocument{
		AppName:    fallback(appName, "Generated App"),
		ClientName: clientName,
		CSSURL:     "./app.css",
		JSURL:      "./app.js",
	}
	return executeTemplate(previewDocumentTemplate, doc)
}

func previewModelFromConfirmation(conf Confirmation) previewModel {
	return previewModel{
		JobID:         "local-preview",
		AppName:       fallback(conf.AppName, "Generated App"),
		Domain:        fallback(conf.Domain, "saas"),
		Plan:          fallback(conf.Plan, "starter"),
		Region:        fallback(conf.Region, "us-east-1"),
		Deployment:    fallback(conf.DeploymentTarget, "managed"),
		PrimaryUsers:  withDefault(conf.PrimaryUsers, "admin"),
		Workflows:     withDefault(conf.CoreWorkflows, "approve_request"),
		Entities:      withDefault(conf.DataEntities, "account"),
		Integrations:  withDefault(conf.Integrations, "none"),
		WorkloadCount: len(conf.CoreWorkflows) + 6,
		FileCount:     len(conf.CoreWorkflows) + len(conf.DataEntities) + 6,
		TerminalCount: 6,
		ConsoleCount:  6,
	}
}

func buildPreviewModel(job Job) previewModel {
	return previewModel{
		JobID:         job.JobID,
		AppName:       fallback(job.Confirmation.AppName, "Generated App"),
		Domain:        fallback(job.Confirmation.Domain, "saas"),
		Plan:          fallback(job.Confirmation.Plan, "starter"),
		Region:        fallback(job.Confirmation.Region, "us-east-1"),
		Deployment:    fallback(job.Confirmation.DeploymentTarget, "managed"),
		PrimaryUsers:  withDefault(job.Confirmation.PrimaryUsers, "admin"),
		Workflows:     withDefault(job.Confirmation.CoreWorkflows, "approve_request"),
		Entities:      withDefault(job.Confirmation.DataEntities, "account"),
		Integrations:  withDefault(job.Confirmation.Integrations, "none"),
		WorkloadCount: len(job.Workload),
		FileCount:     len(job.Files),
		TerminalCount: len(job.TerminalLogs),
		ConsoleCount:  len(job.ConsoleLogs),
	}
}

func runtimeClient(client string) string {
	if strings.EqualFold(strings.TrimSpace(client), "mobile") {
		return "mobile"
	}
	return "web"
}

func lookupGeneratedRuntimeAsset(files []FileArtifact, mode, asset string) (string, []byte, bool) {
	suffix := filepath.ToSlash(filepath.Join("clients", mode, asset))
	for _, f := range files {
		if strings.HasSuffix(filepath.ToSlash(f.Path), suffix) {
			return runtimeContentType(asset), []byte(f.Content), true
		}
	}
	return "", nil, false
}

func runtimeContentType(asset string) string {
	switch strings.ToLower(filepath.Ext(asset)) {
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	default:
		return "text/html; charset=utf-8"
	}
}

func webRuntimeJS(model previewModel) string {
	payload, _ := json.Marshal(model)
	return fmt.Sprintf(`(function () {
  const model = %s;
  const state = {
    view: "dashboard",
    entities: model.Entities.map((name, idx) => ({ id: idx + 1, name, status: "active" })),
    workflows: model.Workflows.map((name, idx) => ({ id: idx + 1, name, status: "ready" })),
  };

  const root = document.getElementById("app");
  if (!root) return;

  function esc(v) {
    return String(v)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function navButton(id, label) {
    return '<button class="tab ' + (state.view === id ? 'active' : '') + '" data-view="' + id + '">' + label + '</button>';
  }

  function metrics() {
    return [
      ["Workload", model.WorkloadCount],
      ["Files", model.FileCount],
      ["Terminal", model.TerminalCount],
      ["Console", model.ConsoleCount],
    ].map(([k, v]) => '<div class="metric"><div class="k">' + esc(k) + '</div><div class="v">' + esc(v) + '</div></div>').join("");
  }

  function dashboardView() {
    return '<section class="panel"><h2>Runtime Overview</h2><div class="metrics">' + metrics() + '</div><div class="hint">Click around to validate generated client structure before deployment.</div></section>';
  }

  function entitiesView() {
    const rows = state.entities.map((e) => '<tr><td>' + esc(e.id) + '</td><td>' + esc(e.name) + '</td><td>' + esc(e.status) + '</td></tr>').join("");
    return '<section class="panel"><h2>Entities</h2><table><thead><tr><th>ID</th><th>Name</th><th>Status</th></tr></thead><tbody>' + rows + '</tbody></table></section>';
  }

  function workflowsView() {
    const items = state.workflows.map((wf) => '<div class="item"><strong>' + esc(wf.name) + '</strong><span>' + esc(wf.status) + '</span></div>').join("");
    return '<section class="panel"><h2>Workflows</h2><div class="list">' + items + '</div></section>';
  }

  function integrationsView() {
    const chips = model.Integrations.map((it) => '<span class="chip">' + esc(it) + '</span>').join("");
    return '<section class="panel"><h2>Integrations</h2><div class="chips">' + chips + '</div></section>';
  }

  function opsView() {
    return '<section class="panel"><h2>Ops</h2><div class="item"><strong>Job</strong><span>' + esc(model.JobID) + '</span></div><div class="item"><strong>Plan</strong><span>' + esc(model.Plan) + '</span></div><div class="item"><strong>Deployment</strong><span>' + esc(model.Deployment) + '</span></div></section>';
  }

  function viewHTML() {
    switch (state.view) {
      case "entities":
        return entitiesView();
      case "workflows":
        return workflowsView();
      case "integrations":
        return integrationsView();
      case "ops":
        return opsView();
      default:
        return dashboardView();
    }
  }

  function render() {
    root.innerHTML = '<div class="shell">'
      + '<aside class="sidebar">'
      + '<h1>' + esc(model.AppName) + '</h1>'
      + '<p>' + esc(model.Domain) + ' · ' + esc(model.Region) + '</p>'
      + navButton('dashboard', 'Dashboard')
      + navButton('entities', 'Entities')
      + navButton('workflows', 'Workflows')
      + navButton('integrations', 'Integrations')
      + navButton('ops', 'Ops')
      + '</aside>'
      + '<main class="main">'
      + viewHTML()
      + '</main>'
      + '</div>';

    root.querySelectorAll("[data-view]").forEach((btn) => {
      btn.addEventListener("click", () => {
        state.view = btn.getAttribute("data-view") || "dashboard";
        render();
      });
    });
  }

  render();
})();`, string(payload))
}

func mobileRuntimeJS(model previewModel) string {
	payload, _ := json.Marshal(model)
	return fmt.Sprintf(`(function () {
  const model = %s;
  const state = { view: "home" };
  const root = document.getElementById("app");
  if (!root) return;

  function esc(v) {
    return String(v)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function nav(id, label) {
    return '<button class="tab ' + (state.view === id ? 'active' : '') + '" data-view="' + id + '">' + label + '</button>';
  }

  function home() {
    return '<section class="card"><div class="k">App</div><div class="v">' + esc(model.AppName) + '</div></section>'
      + '<section class="card"><div class="k">Plan</div><div class="v">' + esc(model.Plan) + '</div></section>'
      + '<section class="card"><div class="k">Deployment</div><div class="v">' + esc(model.Deployment) + '</div></section>';
  }

  function flows() {
    return '<section class="card"><div class="k">Flows</div><ul>' + model.Workflows.map((wf) => '<li>' + esc(wf) + '</li>').join("") + '</ul></section>';
  }

  function data() {
    return '<section class="card"><div class="k">Data</div><ul>' + model.Entities.map((e) => '<li>' + esc(e) + '</li>').join("") + '</ul></section>';
  }

  function ops() {
    return '<section class="card"><div class="k">Job</div><div class="v">' + esc(model.JobID) + '</div></section>'
      + '<section class="card"><div class="k">Integrations</div><div class="chips">' + model.Integrations.map((it) => '<span class="chip">' + esc(it) + '</span>').join("") + '</div></section>';
  }

  function body() {
    switch (state.view) {
      case "flows":
        return flows();
      case "data":
        return data();
      case "ops":
        return ops();
      default:
        return home();
    }
  }

  function render() {
    root.innerHTML = '<div class="phone">'
      + '<header><h1>' + esc(model.AppName) + '</h1><p>' + esc(model.Domain) + '</p></header>'
      + '<main>' + body() + '</main>'
      + '<nav>'
      + nav('home', 'Home')
      + nav('flows', 'Flows')
      + nav('data', 'Data')
      + nav('ops', 'Ops')
      + '</nav>'
      + '</div>';

    root.querySelectorAll("[data-view]").forEach((btn) => {
      btn.addEventListener("click", () => {
        state.view = btn.getAttribute("data-view") || "home";
        render();
      });
    });
  }

  render();
})();`, string(payload))
}

const webRuntimeCSS = `
:root { --bg:#071123; --panel:#12213d; --ink:#eaf1ff; --muted:#9bb2dc; --accent:#21c7aa; }
* { box-sizing:border-box; }
body { margin:0; font-family:"Plus Jakarta Sans","Avenir Next","Segoe UI",sans-serif; background:linear-gradient(170deg,#071123,#0f1f39 70%,#1f3558); color:var(--ink); min-height:100vh; }
.shell { display:grid; grid-template-columns:240px 1fr; min-height:100vh; }
.sidebar { padding:16px; border-right:1px solid rgba(164,191,241,.24); background:rgba(9,18,35,.88); }
.sidebar h1 { margin:0 0 4px; font-size:18px; }
.sidebar p { margin:0 0 12px; color:var(--muted); font-size:12px; }
.tab { width:100%%; text-align:left; padding:9px; margin:0 0 8px; border-radius:10px; border:1px solid rgba(164,191,241,.24); background:#102447; color:#e6efff; cursor:pointer; }
.tab.active { border-color:#2de0c1; background:linear-gradient(120deg,#0d4f61,#175081); }
.main { padding:16px; }
.panel { border:1px solid rgba(164,191,241,.24); border-radius:12px; background:rgba(18,33,61,.85); padding:12px; }
.panel h2 { margin:0 0 10px; }
.metrics { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:8px; margin-bottom:12px; }
.metric { border:1px solid rgba(164,191,241,.24); background:#102447; border-radius:10px; padding:10px; }
.metric .k { font-size:11px; text-transform:uppercase; color:var(--muted); }
.metric .v { font-size:22px; margin-top:6px; font-weight:700; }
.hint { color:var(--muted); font-size:13px; }
table { width:100%%; border-collapse:collapse; }
th, td { text-align:left; border-bottom:1px solid rgba(164,191,241,.2); padding:8px; font-size:13px; }
.list { display:grid; gap:8px; }
.item { display:flex; justify-content:space-between; gap:8px; padding:9px; border-radius:10px; border:1px solid rgba(164,191,241,.24); background:#102447; }
.chips { display:flex; flex-wrap:wrap; gap:8px; }
.chip { border:1px solid rgba(164,191,241,.34); border-radius:999px; background:#102447; padding:5px 9px; font-size:12px; }
@media (max-width: 920px) { .shell { grid-template-columns:1fr; } .metrics { grid-template-columns:repeat(2,minmax(0,1fr)); } }
`

const mobileRuntimeCSS = `
:root { --bg:#0b1220; --panel:#1f2937; --ink:#f8fafc; --muted:#9fb1cb; }
* { box-sizing:border-box; }
body { margin:0; min-height:100vh; background:radial-gradient(circle at 18%% 8%%,#1b3355,#0a111d); font-family:"Plus Jakarta Sans","Avenir Next","Segoe UI",sans-serif; color:var(--ink); }
#app { min-height:100vh; display:flex; align-items:center; justify-content:center; padding:16px; }
.phone { width:100%%; max-width:390px; height:760px; border-radius:28px; border:1px solid rgba(159,177,203,.35); background:linear-gradient(180deg,#0f172a,#111827); box-shadow:0 28px 60px rgba(2,6,23,.64); display:grid; grid-template-rows:auto 1fr auto; overflow:hidden; }
header { padding:14px; border-bottom:1px solid rgba(159,177,203,.2); background:rgba(15,23,42,.92); }
header h1 { margin:0; font-size:16px; }
header p { margin:3px 0 0; color:var(--muted); font-size:12px; }
main { padding:12px; overflow:auto; display:grid; gap:10px; }
.card { border:1px solid rgba(159,177,203,.27); border-radius:14px; padding:11px; background:rgba(31,41,55,.86); }
.k { color:var(--muted); font-size:11px; text-transform:uppercase; }
.v { margin-top:6px; font-weight:700; }
ul { margin:8px 0 0; padding-left:18px; }
li { margin:5px 0; }
nav { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:4px; padding:8px; border-top:1px solid rgba(159,177,203,.2); background:rgba(15,23,42,.96); }
.tab { border:0; border-radius:10px; padding:8px 5px; font-size:11px; background:#1f2937; color:#d6e2f4; cursor:pointer; }
.tab.active { background:linear-gradient(120deg,#0e7490,#0f766e); color:#ecfeff; }
.chips { display:flex; gap:8px; flex-wrap:wrap; }
.chip { border:1px solid rgba(159,177,203,.35); border-radius:999px; padding:5px 8px; font-size:11px; }
`

const previewDocumentTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>{{.AppName}} {{.ClientName}} Runtime</title>
  <link rel="stylesheet" href="{{.CSSURL}}" />
</head>
<body>
  <div id="app"></div>
  <script src="{{.JSURL}}" defer></script>
</body>
</html>`

func executeTemplate(tpl string, model any) string {
	t, err := template.New("preview").Parse(tpl)
	if err != nil {
		return fmt.Sprintf("<html><body><pre>template_parse_failed: %s</pre></body></html>", template.HTMLEscapeString(err.Error()))
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, model); err != nil {
		return fmt.Sprintf("<html><body><pre>template_exec_failed: %s</pre></body></html>", template.HTMLEscapeString(err.Error()))
	}
	return buf.String()
}

func fallback(value, defaultValue string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultValue
	}
	return value
}

func withDefault(items []string, defaultItem string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return []string{defaultItem}
	}
	return out
}
