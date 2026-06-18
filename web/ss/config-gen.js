// shadowsocks-rust config functions — SIP002 URI parsing and config generation
// Requires: URI.parse() from uri-js (loaded by HTML page)

// Supported encryption methods
var SS_METHODS = [
  "2022-blake3-aes-256-gcm",
  "2022-blake3-aes-128-gcm",
  "2022-blake3-chacha20-poly1305",
  "aes-256-gcm",
  "aes-128-gcm",
  "chacha20-ietf-poly1305"
];

// Parse a raw ss:// URI (SIP002 format) → ss outbound object
// Supports:
//   ss://base64(method:password)@server:port#name           (standard SIP002)
//   ss://base64(method:password)@server:port/?plugin=xxx#name
//   ss://base64(method:password@server:port)#name           (non-standard, entire userinfo encoded)
function parseSsUri(raw) {
  var uri = raw.trim();
  if (!uri.startsWith("ss://")) throw new Error("不是 ss:// 链接");

  var body = uri.substring(5); // remove "ss://"
  var hashIdx = body.indexOf("#");
  var tag = hashIdx >= 0 ? decodeURIComponent(body.substring(hashIdx + 1)) : "";
  if (hashIdx >= 0) body = body.substring(0, hashIdx);

  var atIdx = body.lastIndexOf("@");
  var method, password, host, port, query = "";

  if (atIdx >= 0) {
    // Has "@" — could be standard SIP002 or plaintext method:password@host:port
    var userinfoPart = body.substring(0, atIdx);
    var serverPart = body.substring(atIdx + 1);

    // Try to split query from serverPart
    var qIdx = serverPart.indexOf("?");
    if (qIdx >= 0) {
      query = serverPart.substring(qIdx + 1);
      serverPart = serverPart.substring(0, qIdx);
    }
    if (serverPart.startsWith("/")) serverPart = serverPart.substring(1);

    var lastColon = serverPart.lastIndexOf(":");
    var candidateHost = lastColon > 0 ? serverPart.substring(0, lastColon) : serverPart;
    var candidatePort = lastColon > 0 ? parseInt(serverPart.substring(lastColon + 1)) : 8388;

    // Try decoding userinfoPart as base64
    var decoded;
    try { decoded = atob(userinfoPart); } catch(e) { decoded = null; }

    if (decoded && decoded.indexOf(":") >= 0) {
      // base64-decoded successfully
      var colonIdx = decoded.indexOf(":");
      var decodedMethod = decoded.substring(0, colonIdx);
      var decodedRest = decoded.substring(colonIdx + 1);

      // Check if decoded contains server info (non-standard: method:password@server:port)
      var innerAt = decodedRest.lastIndexOf("@");
      if (innerAt >= 0) {
        // Non-standard: server info is inside the decoded base64
        method = decodedMethod;
        password = decodedRest.substring(0, innerAt);
        var innerServer = decodedRest.substring(innerAt + 1);
        var innerColon = innerServer.lastIndexOf(":");
        host = innerColon > 0 ? innerServer.substring(0, innerColon) : innerServer;
        port = innerColon > 0 ? parseInt(innerServer.substring(innerColon + 1)) : 8388;
      } else {
        // Standard SIP002: method:password in base64, server:port after @
        method = decodedMethod;
        password = decodedRest;
        host = candidateHost;
        port = candidatePort;
      }
    } else {
      // Not base64 — treat as plaintext method:password@server:port
      var colonIdx = userinfoPart.indexOf(":");
      if (colonIdx >= 0) {
        method = userinfoPart.substring(0, colonIdx);
        password = userinfoPart.substring(colonIdx + 1);
      } else {
        method = userinfoPart;
        password = "";
      }
      host = candidateHost;
      port = candidatePort;
    }
  } else {
    // No "@" — entire string is base64(method:password@server:port)
    var decoded;
    try { decoded = atob(body); } catch(e) { throw new Error("无法解析 ss:// 链接"); }

    var atIdx2 = decoded.indexOf("@");
    if (atIdx2 >= 0) {
      var methodPass = decoded.substring(0, atIdx2);
      var serverPart = decoded.substring(atIdx2 + 1);
      var colonIdx = methodPass.indexOf(":");
      method = methodPass.substring(0, colonIdx);
      password = methodPass.substring(colonIdx + 1);
      var lastColon = serverPart.lastIndexOf(":");
      host = serverPart.substring(0, lastColon);
      port = parseInt(serverPart.substring(lastColon + 1));
    } else {
      var colonIdx = decoded.indexOf(":");
      method = decoded.substring(0, colonIdx);
      password = decoded.substring(colonIdx + 1);
      host = "";
      port = 8388;
    }
  }

  if (!tag) tag = host + ":" + port;

  var result = {
    server: host,
    server_port: port || 8388,
    method: method,
    password: password,
    plugin: "",
    plugin_opts: "",
    tag: tag
  };

  // Parse query parameters (plugin, etc.)
  if (query) {
    query.split("&").forEach(function(pair) {
      var eqIdx = pair.indexOf("=");
      if (eqIdx < 0) return;
      var key = decodeURIComponent(pair.substring(0, eqIdx));
      var val = decodeURIComponent(pair.substring(eqIdx + 1));
      if (key === "plugin") {
        // plugin format: "name;opts" or "name%3Bopts"
        var parts = val.split(";");
        result.plugin = parts[0] || "";
        result.plugin_opts = parts.slice(1).join(";") || "";
      }
    });
  }

  return result;
}

// Generate complete shadowsocks-rust config JSON from a single outbound object.
// socksPort/httpPort override the defaults.
// Output format: shadowsocks-rust "locals" style (multi-protocol)
function generateSsConfig(outbound, socksPort, httpPort) {
  var cfg = {
    server: outbound.server,
    server_port: outbound.server_port,
    password: outbound.password,
    method: outbound.method,
    mode: "tcp_and_udp",
    locals: [
      {
        protocol: "socks",
        local_address: "127.0.0.1",
        local_port: parseInt(socksPort) || 10808,
        mode: "tcp_and_udp"
      },
      {
        protocol: "http",
        local_address: "127.0.0.1",
        local_port: parseInt(httpPort) || 10809
      }
    ]
  };

  // Optional plugin
  if (outbound.plugin) {
    cfg.plugin = outbound.plugin;
    if (outbound.plugin_opts) cfg.plugin_opts = outbound.plugin_opts;
  }

  return JSON.stringify(cfg, null, 2);
}

// Generate ss inbound config from localStorage ports (for MDPC inbound page compatibility).
// Returns a minimal object with just the local settings.
function generateSsInbound() {
  var socksPort = parseInt(localStorage.getItem("socks-port")) || 10808;
  var httpPort = parseInt(localStorage.getItem("http-port")) || 10809;
  return {
    locals: [
      { protocol: "socks", local_address: "127.0.0.1", local_port: socksPort, mode: "tcp_and_udp" },
      { protocol: "http",  local_address: "127.0.0.1", local_port: httpPort }
    ]
  };
}