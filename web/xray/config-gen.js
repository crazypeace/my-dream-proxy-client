// xray config functions — URI parsing, direct outbound, config generation
// Requires: URI.parse() from uri-js (loaded by HTML page)

// Parse a raw vless://vmess://trojan:// URI → xray outbound object
function parseXrayUri(raw) {
  var parts = URI.parse(raw.trim());
  var params = {};
  if (parts.query) {
    parts.query.split("&").forEach(function(pair) {
      var eqIdx = pair.indexOf("=");
      if (eqIdx >= 0) {
        params[decodeURIComponent(pair.substring(0, eqIdx))] = decodeURIComponent(pair.substring(eqIdx + 1));
      } else if (pair) {
        params[decodeURIComponent(pair)] = "";
      }
    });
  }

  var user = {
    id: parts.userinfo ? decodeURIComponent(parts.userinfo) : "",
    encryption: params.encryption || "none"
  };
  if (params.flow) user.flow = params.flow;

  var ss = {
    network: params.type || "tcp",
    security: params.security || "none"
  };

  if (params.security === "reality") {
    ss.realitySettings = {
      serverName: params.sni || parts.host,
      fingerprint: params.fp || "chrome",
      publicKey: params.pbk,
      shortId: params.sid || "",
      spiderX: params.spx || ""
    };
  } else if (params.security === "tls") {
    ss.tlsSettings = {
      serverName: params.sni || parts.host,
      fingerprint: params.fp || "chrome",
      alpn: params.alpn ? params.alpn.split(",") : undefined
    };
    if (params.allowInsecure === "1") ss.tlsSettings.allowInsecure = true;
  }

  if (params.type === "ws") {
    ss.wsSettings = {
      path: params.path || "/",
      headers: { Host: params.host || parts.host }
    };
  } else if (params.type === "grpc") {
    ss.grpcSettings = { serviceName: params.serviceName || "" };
  }

  return {
    tag: "proxy",
    protocol: parts.scheme,
    settings: { vnext: [{ address: parts.host, port: parts.port || 443, users: [user] }] },
    streamSettings: ss
  };
}

// Create a freedom/direct outbound
function makeXrayDirect() {
  return { tag: "direct", protocol: "freedom" };
}

// Generate complete xray config JSON from an array of outbounds.
// Direct outbound is appended automatically.
function generateXrayConfig(outbounds) {
  var all = outbounds.concat([makeXrayDirect()]);
  return JSON.stringify({ outbounds: all }, null, 2);
}

// Generate xray inbound config from localStorage ports.
// Keys: "socks-port", "http-port" (defaults: 10808, 10809)
function generateXrayInbound() {
  var socksPort = parseInt(localStorage.getItem("socks-port")) || 10808;
  var httpPort = parseInt(localStorage.getItem("http-port")) || 10809;
  return {
    inbounds: [
      { listen: "127.0.0.1", port: socksPort, protocol: "socks", settings: { udp: true }, tag: "socks-in" },
      { listen: "127.0.0.1", port: httpPort, protocol: "http", tag: "http-in" }
    ]
  };
}
