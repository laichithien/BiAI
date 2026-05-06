(function () {
  var token = "";
  var pendingApproval = null;
  var sessionID = "";

  function init() {
    token = getQuery("token");
    sessionID = getStoredSession();
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
    byId("status").className = "statusBadge";
  }

  function setBusy(text) {
    byId("status").innerText = text;
    byId("status").className = "statusBadge busy";
  }

  function setError(text) {
    byId("status").innerText = text;
    byId("status").className = "statusBadge error";
  }

  function addMessage(kind, text) {
    clearEmptyState();
    var div = document.createElement("div");
    div.className = "msg " + kind;
    var head = document.createElement("div");
    head.className = "msgHead";
    head.appendChild(document.createTextNode(kind === "user" ? "You" : "BiAI"));
    div.appendChild(head);
    var body = document.createElement("div");
    body.className = "msgBody";
    var pre = document.createElement("pre");
    pre.appendChild(document.createTextNode(text));
    body.appendChild(pre);
    div.appendChild(body);
    byId("chat").appendChild(div);
    byId("chat").scrollTop = byId("chat").scrollHeight;
  }

  function addToolMessage(ev) {
    clearEmptyState();
    var div = document.createElement("div");
    div.className = "msg tool " + (ev.ok ? "ok" : "err");
    var head = document.createElement("div");
    head.className = "msgHead";
    head.appendChild(document.createTextNode((ev.ok ? "Tool completed" : "Tool blocked") + " · " + ev.name));
    var body = document.createElement("div");
    body.className = "msgBody";
    body.appendChild(document.createTextNode(ev.message || ""));
    if (ev.data && ev.data.output) {
      var pre = document.createElement("pre");
      pre.appendChild(document.createTextNode("\n" + ev.data.output));
      body.appendChild(pre);
    }
    div.appendChild(head);
    div.appendChild(body);
    byId("chat").appendChild(div);
    byId("chat").scrollTop = byId("chat").scrollHeight;
  }

  function clearEmptyState() {
    var nodes = byId("chat").getElementsByTagName("div");
    for (var i = 0; i < nodes.length; i++) {
      if (nodes[i].className === "emptyState") {
        nodes[i].parentNode.removeChild(nodes[i]);
        return;
      }
    }
  }

  function addEvent(ev) {
    if (!ev) return;
    var div = document.createElement("div");
    div.className = "event";
    div.appendChild(document.createTextNode((ev.ok ? "OK " : "ERR ") + ev.name + ": " + ev.message));
    byId("events").insertBefore(div, byId("events").firstChild);
    addToolMessage(ev);
  }

  function sendPrompt() {
    var prompt = byId("prompt").value;
    var workspace = byId("workspace").value;
    if (!prompt) return;
    byId("prompt").value = "";
    addMessage("user", prompt);
    setBusy("Dang chay");
    postJSON("/api/chat", { prompt: prompt, workspace: workspace, session_id: sessionID }, function (err, resp) {
      setStatus("San sang");
      if (err) {
        setError("Loi");
        addMessage("agent", "Error: " + err);
        return;
      }
      if (resp.events) {
        for (var i = 0; i < resp.events.length; i++) addEvent(resp.events[i]);
      }
      if (resp.session_id) {
        sessionID = resp.session_id;
        saveStoredSession(sessionID);
      }
      addMessage("agent", resp.message || "");
      if (resp.approval) showApproval(resp.approval);
    });
  }

  function showApproval(ap) {
    pendingApproval = ap;
    var lines = [];
    lines.push("Tool: " + ap.tool_name);
    lines.push("Muc rui ro: " + ap.risk);
    lines.push("Lenh: " + (ap.command || ""));
    lines.push("Thu muc chay: " + (ap.cwd || ""));
    lines.push("Ly do: " + (ap.reason || ""));
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
    setStatus("Dang xac nhan");
    postJSON("/api/approval", { approval_id: id, decision: decision }, function (err, resp) {
      pendingApproval = null;
      setStatus("San sang");
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
        byId("settingsStatus").innerText = "Khong doc duoc cau hinh";
        return;
      }
      byId("llmBaseURL").value = resp.llm_base_url || "";
      setModelOptions(resp.model || "", []);
      byId("settingsStatus").innerText = resp.has_token ? "Da luu token" : "Chua co token";
    });
  }

  function loadHealth() {
    getJSON("/api/health", function (err, resp) {
      if (err) {
        byId("diag").innerText = "Health check failed: " + err;
        return;
      }
      if (!byId("workspace").value && resp.default_workspace) {
        byId("workspace").value = resp.default_workspace;
      }
      byId("diag").innerText = "Data: " + (resp.data_dir || "") + "\nLog: " + (resp.log_path || "") + "\nHistory: " + (resp.history_path || "");
      loadContext();
    });
  }

  function loadContext() {
    var workspace = encodeURIComponent(byId("workspace").value || "");
    getJSON("/api/context?workspace=" + workspace, function (err, resp) {
      if (err || !resp || !resp.instructions) return;
      if (resp.instructions.length) {
        addEvent({ ok: true, name: "instructions.loaded", message: resp.instructions.length + " instruction file(s)" });
      }
    });
  }

  function saveSettings() {
    var payload = {
      llm_base_url: byId("llmBaseURL").value,
      model: byId("model").value,
      api_token: byId("apiToken").value
    };
    byId("settingsStatus").innerText = "Dang luu...";
    postJSON("/api/settings", payload, function (err) {
      if (err) {
        byId("settingsStatus").innerText = "Luu loi: " + err;
        return;
      }
      byId("apiToken").value = "";
      byId("settingsStatus").innerText = "Da luu. Token duoc an vi bao mat.";
      byId("modelFetchBox").className = "fetchBox ok";
    });
  }

  function fetchModels() {
    var btn = byId("fetchModels");
    btn.disabled = true;
    byId("modelFetchBox").className = "fetchBox busy";
    byId("settingsStatus").innerText = "Dang tai model...";
    byId("settingsDetail").innerText = "Dang goi /models. Neu URL/token sai, loi se hien o day.";
    addEvent({ ok: true, name: "models.fetch", message: "Dang goi API lay danh sach model" });
    postJSON("/api/models", {
      llm_base_url: byId("llmBaseURL").value,
      api_token: byId("apiToken").value
    }, function (err, resp) {
      btn.disabled = false;
      if (err) {
        byId("modelFetchBox").className = "fetchBox error";
        byId("settingsStatus").innerText = "Tai model loi: " + err;
        byId("settingsDetail").innerText = err;
        addEvent({ ok: false, name: "models.fetch", message: err });
        return;
      }
      var current = byId("model").value;
      setModelOptions(current, resp.models || []);
      byId("modelFetchBox").className = "fetchBox ok";
      byId("settingsStatus").innerText = "Da tai " + (resp.models ? resp.models.length : 0) + " model";
      byId("settingsDetail").innerText = resp.models && resp.models.length ? "Chon model trong dropdown roi bam Luu." : "API tra ve 0 model.";
      addEvent({ ok: true, name: "models.fetch", message: "Da tai xong danh sach model" });
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
    if (!selected && models.length === 0) addOption(select, "", "Bam Tai model");
    select.value = selected || (models.length ? models[0].id : "");
  }

  function addOption(select, value, label) {
    var opt = document.createElement("option");
    opt.value = value;
    opt.appendChild(document.createTextNode(label));
    select.appendChild(opt);
  }

  function getStoredSession() {
    try {
      var existing = window.localStorage ? window.localStorage.getItem("biai.session_id") : "";
      if (existing) return existing;
    } catch (e) {}
    return "session_" + (new Date().getTime());
  }

  function saveStoredSession(id) {
    try {
      if (window.localStorage) window.localStorage.setItem("biai.session_id", id);
    } catch (e) {}
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
