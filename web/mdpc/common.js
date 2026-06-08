const CORE_NAME = "mdpc";

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
    // auto-sync after loading
    syncOutboundsFromYaml();
  } catch (e) {
    showMsg("连接后端失败: " + e.message, false);
  }
}

async function saveConfig() {
  const content = document.getElementById("editor").value;
  if (!content.trim()) {
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
