#!/usr/bin/env python3
"""Private Proxykonfiguration aus unverändertem Basistemplate plus festen Messfeldern."""
from pathlib import Path
import sys

root = Path.cwd()
output = Path(sys.argv[1])
nginx = (root / "deploy/nginx/nginx.conf").read_text()
nginx = nginx.replace("access_log off;", """map $upstream_http_x_shop_diagnostic_id $diagnostic_log {
        default 0;
        ~^[0-9]+$ 1;
    }
    log_format diagnostic '$upstream_http_x_shop_diagnostic_id $status $body_bytes_sent $request_time $upstream_connect_time $upstream_header_time $upstream_response_time';
    access_log /dev/stdout diagnostic if=$diagnostic_log;""", 1)
(output / "nginx.conf").write_text(nginx)
varnish = (root / "deploy/varnish/default.vcl").read_text()
varnish = varnish.replace("sub vcl_deliver {", """sub vcl_deliver {
    if (resp.http.X-Shop-Diagnostic-Id ~ "^[0-9]+$") {
        std.log("diag:" + resp.http.X-Shop-Diagnostic-Id);
    }""", 1)
if "import std;" not in varnish:
    varnish = varnish.replace("vcl 4.1;", "vcl 4.1;\nimport std;", 1)
(output / "default.vcl").write_text(varnish)
