// hy2 config functions — URI parsing and config generation
// Requires: URI.parse() from uri-js (loaded by HTML page)

// Parse a raw hysteria2:// or hy2:// URI → hy2 outbound object
function parseHy2Uri(raw) {
  var uri = raw.trim();
  // Normalize prefix
  if (uri.startsWith("hy2://")) uri = "hysteria2://" + uri.substring(6);

  var parts = URI.parse(uri);
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

  return {
    server: parts.host + ":" + (parts.port || 443),
    auth: parts.userinfo ? decodeURIComponent(parts.userinfo) : "",
    tls: {
      sni: params.sni || undefined,
      insecure: (params.insecure === "1") ? true : undefined,
      pinSHA256: params.pinSHA256 || params["pin-sha256"] || undefined,
      alpn: params.alpn ? [params.alpn] : undefined
    },
    socks5: { listen: "127.0.0.1:10808" },
    http: { listen: "127.0.0.1:10809" }
  };
}

// Generate complete hy2 config YAML from a single outbound object.
// socksPort/httpPort can override the defaults in the outbound.
function generateHy2Config(outbound, socksPort, httpPort) {
  var cfg = {
    server: outbound.server,
    auth: outbound.auth,
    tls: outbound.tls || {}
  };
  // Clean undefined tls fields
  for (var k in cfg.tls) {
    if (cfg.tls[k] === undefined) delete cfg.tls[k];
  }
  if (Object.keys(cfg.tls).length === 0) delete cfg.tls;

  cfg.socks5 = { listen: "127.0.0.1:" + (socksPort || "10808") };
  cfg.http = { listen: "127.0.0.1:" + (httpPort || "10809") };

  return jsyaml.dump(cfg, { lineWidth: -1 });
}
