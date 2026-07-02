// v2ray config functions — VMess URI parsing, direct outbound, config generation
// VMess share link format: vmess://base64(json)
// The JSON payload uses the standard v2rayN field set:
//   v, ps, add, port, id, aid, scy, net, type, host, path, tls, sni, alpn, fp

// Decode base64 (handles URL-safe variant and missing padding)
function v2rayB64Decode(str) {
  var s = str.replace(/-/g, "+").replace(/_/g, "/");
  while (s.length % 4 !== 0) s += "=";
  return decodeURIComponent(escape(atob(s)));
}

// Detect whether a host string is a literal IP address (IPv4 or IPv6).
// Used to auto-skip TLS cert verification: IP endpoints almost never have
// a valid cert matching the connect address.
function v2rayIsIP(host) {
  if (!host) return false;
  host = host.trim();
  // Strip brackets from bracketed IPv6, e.g. [2001:db8::1]
  if (host.charAt(0) === "[" && host.charAt(host.length - 1) === "]") {
    host = host.substring(1, host.length - 1);
  }
  // IPv4: four 0-255 octets
  var ipv4 = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/;
  var m = host.match(ipv4);
  if (m) {
    for (var i = 1; i <= 4; i++) {
      if (parseInt(m[i], 10) > 255) return false;
    }
    return true;
  }
  // IPv6: presence of ":" plus only hex/colon chars is a reliable signal
  // (domains never contain ":"). Covers ::, compressed, and full forms.
  if (host.indexOf(":") >= 0 && /^[0-9a-fA-F:]+$/.test(host)) {
    return true;
  }
  return false;
}

// Parse a raw vmess:// URI → v2ray outbound object
function parseVmessUri(raw) {
  raw = raw.trim();
  if (raw.toLowerCase().indexOf("vmess://") !== 0) {
    throw new Error("不是 vmess:// 链接");
  }
  var payload = raw.substring(8); // strip "vmess://"
  var obj;
  try {
    obj = JSON.parse(v2rayB64Decode(payload));
  } catch (e) {
    throw new Error("VMess base64/JSON 解析失败: " + e.message);
  }

  var net = (obj.net || "tcp").toLowerCase();
  var security = (obj.tls || "").toLowerCase(); // "tls" or ""
  var headerType = obj.type || "none";

  // User: alterId 0 → AEAD (modern). scy = client security (auto by default)
  var user = {
    id: obj.id || "",
    alterId: parseInt(obj.aid) || 0,
    security: obj.scy || obj.security || "auto"
  };

  var ss = { network: net, security: security === "tls" ? "tls" : "none" };

  // TLS settings
  if (security === "tls") {
    // effectiveSNI follows the fallback chain: sni → host → add.
    // Auto-skip cert verify when add is an IP, or when the effective SNI
    // differs from add (forged SNI like example.com on JMS/机场 nodes).
    var effectiveSNI = obj.sni || obj.host || obj.add;
    var skipVerify = v2rayIsIP(obj.add) || (effectiveSNI !== obj.add);
    ss.tlsSettings = {
      serverName: effectiveSNI,
      allowInsecure: skipVerify
    };
    if (obj.alpn) ss.tlsSettings.alpn = obj.alpn.split(",").map(function(a){ return a.trim(); }).filter(Boolean);
    if (obj.fp) ss.tlsSettings.fingerprint = obj.fp;
  }

  // Transport-specific settings
  if (net === "ws") {
    ss.wsSettings = {
      path: obj.path || "/",
      headers: { Host: obj.host || obj.add }
    };
  } else if (net === "kcp") {
    ss.kcpSettings = { header: { type: headerType || "none" } };
    if (obj.path) ss.kcpSettings.seed = obj.path;
  } else if (net === "h2" || net === "http") {
    ss.network = "http";
    ss.httpSettings = {
      path: obj.path || "/",
      host: (obj.host ? obj.host.split(",").map(function(h){ return h.trim(); }).filter(Boolean) : [])
    };
  } else if (net === "quic") {
    ss.quicSettings = {
      security: obj.host || "none",
      key: obj.path || "",
      header: { type: headerType || "none" }
    };
  } else if (net === "grpc") {
    ss.grpcSettings = { serviceName: obj.path || "" };
  } else {
    // tcp
    ss.tcpSettings = { header: { type: headerType || "none" } };
    if (headerType === "http") {
      ss.tcpSettings.header = {
        type: "http",
        request: {
          path: obj.path ? obj.path.split(",") : ["/"],
          headers: obj.host ? { Host: obj.host.split(",").map(function(h){ return h.trim(); }) } : {}
        }
      };
    }
  }

  return {
    tag: "proxy",
    protocol: "vmess",
    settings: { vnext: [{ address: obj.add, port: parseInt(obj.port) || 443, users: [user] }] },
    streamSettings: ss
  };
}

// Create a freedom/direct outbound
function makeV2rayDirect() {
  return { tag: "direct", protocol: "freedom" };
}

// Generate complete v2ray config JSON from an array of outbounds.
// Direct outbound is appended automatically.
function generateV2rayConfig(outbounds) {
  var all = outbounds.concat([makeV2rayDirect()]);
  return JSON.stringify({ outbounds: all }, null, 2);
}

// Generate v2ray inbound config from localStorage ports.
// Keys: "socks-port", "http-port" (defaults: 10808, 10809)
function generateV2rayInbound() {
  var socksPort = parseInt(localStorage.getItem("socks-port")) || 10808;
  var httpPort = parseInt(localStorage.getItem("http-port")) || 10809;
  return {
    inbounds: [
      { listen: "127.0.0.1", port: socksPort, protocol: "socks", settings: { udp: true }, tag: "socks-in" },
      { listen: "127.0.0.1", port: httpPort, protocol: "http", tag: "http-in" }
    ]
  };
}
