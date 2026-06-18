// sing-box config functions — URI parsing, direct outbound, config generation
// Requires: URI.parse() from uri-js (loaded by HTML page)

// Parse a raw anytls:// URI → sing-box outbound object
function parseSingboxUri(raw) {
  var uri = raw.trim();
  var parts = URI.parse(uri);
  if (!parts.host || !parts.port) throw new Error("URI 格式不正确: " + uri.substring(0, 40));

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

  var password = parts.userinfo ? decodeURIComponent(parts.userinfo) : "";
  var server = parts.host;
  var port = parseInt(parts.port);
  var name = parts.fragment ? decodeURIComponent(parts.fragment) : "proxy";
  var sni = params.sni || server;
  var insecure = (params.allowInsecure === "1" || params.insecure === "1");

  var ob = {
    type: "anytls",
    tag: name,
    server: server,
    server_port: port,
    password: password,
    tls: { enabled: true, server_name: sni }
  };
  if (insecure) ob.tls.insecure = true;

  return ob;
}

// Create a direct outbound
function makeSingboxDirect() {
  return { type: "direct", tag: "direct" };
}

// Generate complete sing-box config JSON from an array of outbounds.
// First outbound tag is forced to "proxy" (sing-box default).
// Direct outbound is appended automatically.
function generateSingboxConfig(outbounds) {
  if (outbounds.length > 0) outbounds[0].tag = "proxy";
  var all = outbounds.concat([makeSingboxDirect()]);
  return JSON.stringify({ outbounds: all }, null, 2);
}

// Generate sing-box inbound config from localStorage ports.
// Keys: "socks-port", "http-port" (defaults: 10808, 10809)
function generateSingboxInbound() {
  var socksPort = parseInt(localStorage.getItem("socks-port")) || 10808;
  var httpPort = parseInt(localStorage.getItem("http-port")) || 10809;
  return {
    inbounds: [
      { type: "socks", tag: "socks-in", listen: "127.0.0.1", listen_port: socksPort },
      { type: "http", tag: "http-in", listen: "127.0.0.1", listen_port: httpPort }
    ]
  };
}
