// sing-box config functions — URI parsing, direct outbound, config generation

// Parse a raw anytls:// URI → sing-box outbound object
function parseSingboxUri(raw) {
  var uri = raw.trim();
  var regex = /^anytls:\/\/([^@]+)@([^:?#]+):(\d+)(?:\?([^#]*))?(?:#(.*))?$/;
  var m = uri.match(regex);
  if (!m) throw new Error("URI 格式不正确: " + uri.substring(0, 40));

  var password = m[1];
  var server = m[2];
  var port = parseInt(m[3]);
  var params = m[4] || "";
  var name = m[5] ? decodeURIComponent(m[5]) : "proxy";

  var insecure = false;
  var sni = server;
  if (params) {
    params.split("&").forEach(function(pair) {
      var kv = pair.split("=");
      if (kv[0] === "allowInsecure" && kv[1] === "1") insecure = true;
      if (kv[0] === "sni" && kv[1]) sni = decodeURIComponent(kv[1]);
    });
  }

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
