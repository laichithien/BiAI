(function () {
  var token = "";
  var pendingApproval = null;

  function init() {
    token = getQuery("token");
    byId("send").onclick = sendPrompt;
    byId("saveSettings").onclick = saveSettings;
    byId("fetchModels").onclick = fetchModels;
    byId("allow").onclick = function () { decide("allow"); };
    byId("deny").onclick = function () { decide("deny"); };
    byId("prompt").onkeydown = function (e) {
      e = e || window.event;
      if (e.keyCode === 13 && !e.shiftKey) {
        if (e.preventDefault) e.preventDefault();
        sendPrompt();
        return false;
      }
      return true;
    };
    loadSettings();
    loadHealth();
  }

  function byId(id) {
    return document.getElementById(id);
  }

  function getQuery(name) {
    var q = window.location.search.substring(1).split("&");
    for (var i = 0; i < q.length; i++) {
      var p = q[i].split("=");
      if (decodeURIComponent(p[0]) === name) return decodeURIComponent(p[1] || "");
    }
    return "";
  }

  function setStatus(text) {
    byId("status").innerText = text;
  }

  function addMessage(kind, text) {
    var div = document.createElement("div");
    div.className = "msg " + kind;
    var pre = document.createElement("pre");
    pre.appendChild(document.createTextNode(text));
    div.appendChild(pre);
    byId("chat").appendChild(div);
    byId("chat").scrollTop = byId("chat").scrollHeight;
  }

  function addEvent(ev) {
    if (!ev) return;
    var div = document.createElement("div");
    div.className = "event";
    div.appendChild(document.createTextNode((ev.ok ? "OK " : "ERR ") + ev.name + ": " + ev.message));
    byId("events").insertBefore(div, byId("events").firstChild);
  }

  function sendPrompt() {
    var prompt = byId("prompt").value;
    var workspace = byId("workspace").value;
    if (!prompt) return;
    byId("prompt").value = "";
    addMessage("user", prompt);
    setStatus("Running");
    postJSON("/api/chat", { prompt: prompt, workspace: workspace }, function (err, resp) {
      setStatus("Idle");
      if (err) {
        addMessage("agent", "Error: " + err);
        return;
      }
      if (resp.events) {
        for (var i = 0; i < resp.events.length; i++) addEvent(resp.events[i]);
      }
      addMessage("agent", resp.message || "");
      if (resp.approval) showApproval(resp.approval);
    });
  }

  function showApproval(ap) {
    pendingApproval = ap;
    var lines = [];
    lines.push("Tool: " + ap.tool_name);
    lines.push("Risk: " + ap.risk);
    lines.push("Command: " + (ap.command || ""));
    lines.push("Cwd: " + (ap.cwd || ""));
    lines.push("Reason: " + (ap.reason || ""));
    lines.push("");
    lines.push(ap.safety_summary || "");
    byId("approvalBody").innerHTML = "";
    var pre = document.createElement("pre");
    pre.appendChild(document.createTextNode(lines.join("\n")));
    byId("approvalBody").appendChild(pre);
    byId("approval").className = "";
  }

  function decide(decision) {
    if (!pendingApproval) return;
    var id = pendingApproval.id;
    byId("approval").className = "hidden";
    setStatus("Approving");
    postJSON("/api/approval", { approval_id: id, decision: decision }, function (err, resp) {
      pendingApproval = null;
      setStatus("Idle");
      if (err) {
        addMessage("agent", "Approval error: " + err);
        return;
      }
      if (resp.event) addEvent(resp.event);
      addMessage("agent", resp.message || "");
    });
  }

  function loadSettings() {
    getJSON("/api/settings", function (err, resp) {
      if (err) {
      byId("settingsStatus").innerText = "Cannot load settings";
        return;
      }
      byId("llmBaseURL").value = resp.llm_base_url || "";
      setModelOptions(resp.model || "", []);
      byId("settingsStatus").innerText = resp.has_token ? "Token saved" : "No token saved";
    });
  }

  function loadHealth() {
    getJSON("/api/health", function (err, resp) {
      if (err) {
        byId("diag").innerText = "Health check failed: " + err;
        return;
      }
      byId("diag").innerText = "Data: " + (resp.data_dir || "") + "\nLog: " + (resp.log_path || "");
    });
  }

  function saveSettings() {
    var payload = {
      llm_base_url: byId("llmBaseURL").value,
      model: byId("model").value,
      api_token: byId("apiToken").value
    };
    byId("settingsStatus").innerText = "Saving...";
    postJSON("/api/settings", payload, function (err) {
      if (err) {
        byId("settingsStatus").innerText = "Save failed: " + err;
        return;
      }
      byId("apiToken").value = "";
      byId("settingsStatus").innerText = "Saved";
    });
  }

  function fetchModels() {
    byId("settingsStatus").innerText = "Fetching models...";
    postJSON("/api/models", {
      llm_base_url: byId("llmBaseURL").value,
      api_token: byId("apiToken").value
    }, function (err, resp) {
      if (err) {
        byId("settingsStatus").innerText = "Fetch failed: " + err;
        return;
      }
      var current = byId("model").value;
      setModelOptions(current, resp.models || []);
      byId("settingsStatus").innerText = "Fetched " + (resp.models ? resp.models.length : 0) + " models";
    });
  }

  function setModelOptions(selected, models) {
    var select = byId("model");
    select.innerHTML = "";
    if (selected) addOption(select, selected, selected);
    for (var i = 0; i < models.length; i++) {
      if (models[i].id === selected) continue;
      addOption(select, models[i].id, models[i].id);
    }
    if (!selected && models.length === 0) addOption(select, "", "Fetch models first");
    select.value = selected || (models.length ? models[0].id : "");
  }

  function addOption(select, value, label) {
    var opt = document.createElement("option");
    opt.value = value;
    opt.appendChild(document.createTextNode(label));
    select.appendChild(opt);
  }

  function postJSON(url, payload, cb) {
    var xhr = new XMLHttpRequest();
    xhr.open("POST", url, true);
    xhr.setRequestHeader("Content-Type", "application/json");
    xhr.setRequestHeader("X-AgentDesk-Token", token);
    xhr.onreadystatechange = function () {
      if (xhr.readyState !== 4) return;
      if (xhr.status < 200 || xhr.status >= 300) {
        cb(xhr.responseText || ("HTTP " + xhr.status));
        return;
      }
      try {
        cb(null, JSON.parse(xhr.responseText));
      } catch (e) {
        cb(e.message);
      }
    };
    xhr.send(JSON.stringify(payload));
  }

  function getJSON(url, cb) {
    var xhr = new XMLHttpRequest();
    xhr.open("GET", url, true);
    xhr.setRequestHeader("X-AgentDesk-Token", token);
    xhr.onreadystatechange = function () {
      if (xhr.readyState !== 4) return;
      if (xhr.status < 200 || xhr.status >= 300) {
        cb(xhr.responseText || ("HTTP " + xhr.status));
        return;
      }
      try {
        cb(null, JSON.parse(xhr.responseText));
      } catch (e) {
        cb(e.message);
      }
    };
    xhr.send(null);
  }

  if (window.attachEvent) window.attachEvent("onload", init);
  else window.addEventListener("load", init, false);
})();
