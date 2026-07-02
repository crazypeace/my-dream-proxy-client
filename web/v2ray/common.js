const CORE_NAME = "v2ray";

// Load API URL from localStorage
(function() {
  const saved = localStorage.getItem("mdpc-api-url");
  if (saved) document.getElementById("apiUrl").value = saved;
})();

function saveApiUrl() {
  localStorage.setItem("mdpc-api-url", document.getElementById("apiUrl").value);
  showMsg("API 地址已保存", true);
}

function api(path) {
  return document.getElementById("apiUrl").value.replace(/\/+$/, "") + "/api/" + CORE_NAME + path;
}

function showMsg(text, ok) {
  const bar = document.getElementById("statusBar");
  const ts = new Date().toLocaleTimeString();
  bar.innerHTML = "<span class='ts'>[" + ts + "]</span> <span class='" + (ok ? "ok" : "err") + "'>" + text + "</span>";
}

async function readConfig() {
  try {
    const res = await fetch(api("/files/" + FILENAME));
    const data = await res.json();
    if (data.error) { showMsg(data.error, false); return; }
    document.getElementById("editor").value = data.data.content;
    showMsg("配置已加载", true);
  } catch (e) {
    showMsg("连接后端失败: " + e.message, false);
  }
}

async function saveConfig() {
  const content = document.getElementById("editor").value;
  if (!content.trim()) {
    // Empty content → delete the file
    try {
      const res = await fetch(api("/files/" + FILENAME), { method: "DELETE" });
      const data = await res.json();
      if (data.error) { showMsg(data.error, false); return; }
      showMsg("文件已删除", true);
    } catch (e) {
      showMsg("连接后端失败: " + e.message, false);
    }
    return;
  }
  try {
    const res = await fetch(api("/files/" + FILENAME), {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ content })
    });
    const data = await res.json();
    if (data.error) { showMsg(data.error, false); return; }
    showMsg("保存成功", true);
  } catch (e) {
    showMsg("连接后端失败: " + e.message, false);
  }
}

function fillSuggestion() {
  if (typeof SUGGESTION_DATA === "undefined") return;
  document.getElementById("editor").value = JSON.stringify(SUGGESTION_DATA, null, 2);
  showMsg("已填充建议内容", true);
}

let coreBusy = false;

function setStatusFromData(data) {
  if (data && data.running !== undefined) {
    document.getElementById("statusText").textContent =
      data.running ? "运行中 (pid " + data.pid + ")" : "已停止";
  }
}

async function startCore() {
  if (coreBusy) return;
  coreBusy = true;
  try {
    const res = await fetch(api("/core/start"), { method: "POST" });
    const data = await res.json();
    if (data.error) { showMsg(data.error, false); return; }
    setStatusFromData(data.data);
    showMsg("已启动", true);
    // Delayed confirm after restart settles
    setTimeout(updateStatus, 1000);
  } catch (e) {
    showMsg("连接后端失败: " + e.message, false);
  } finally {
    coreBusy = false;
  }
}

async function stopCore() {
  if (coreBusy) return;
  coreBusy = true;
  try {
    const res = await fetch(api("/core/stop"), { method: "POST" });
    const data = await res.json();
    if (data.error) { showMsg(data.error, false); return; }
    setStatusFromData(data.data);
    showMsg("已停止", true);
  } catch (e) {
    showMsg("连接后端失败: " + e.message, false);
  } finally {
    coreBusy = false;
  }
}

async function updateStatus() {
  try {
    const res = await fetch(api("/core/status"));
    const data = await res.json();
    if (data.data) {
      document.getElementById("statusText").textContent =
        data.data.running ? "运行中 (pid " + data.data.pid + ")" : "已停止";
    }
  } catch (e) {
    document.getElementById("statusText").textContent = "无法连接";
  }
}

updateStatus();
