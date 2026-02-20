(function () {
  const model = {"JobID":"local-preview","AppName":"Artifacts","Domain":"saas","Plan":"starter","Region":"us-east-1","Deployment":"managed","PrimaryUsers":["admin","operator"],"Workflows":["create_record","approve_record","notify_user"],"Entities":["account","workspace","activity"],"Integrations":["none"],"WorkloadCount":9,"FileCount":0,"TerminalCount":0,"ConsoleCount":0};
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
})();