(function () {
  const model = {"JobID":"local-preview","AppName":"Policy Defaults","Domain":"saas","Plan":"starter","Region":"us-east-1","Deployment":"managed","PrimaryUsers":["admin","operator"],"Workflows":["create_record","approve_record","notify_user"],"Entities":["account","workspace","activity"],"Integrations":["stripe"],"WorkloadCount":9,"FileCount":12,"TerminalCount":6,"ConsoleCount":6};
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
})();