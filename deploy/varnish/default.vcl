vcl 4.1;

backend default {
    .host = "nginx";
    .port = "8082";
    .connect_timeout = 150ms;
    .first_byte_timeout = 3s;
    .between_bytes_timeout = 150ms;
}

acl invalidators { "127.0.0.1"; "shop"; }

sub vcl_recv {
    if (req.method == "PURGE") {
        if (client.ip !~ invalidators) { return (synth(403)); }
        if (req.http.X-Cache-Key != "SHOP_CACHE_KEY") { return (synth(403)); }
        if (req.http.X-Cache-Public-Search == "1") {
            ban("obj.http.Surrogate-Key ~ (^|[[:space:]])shop:public-search([[:space:]]|$)");
            return (synth(200));
        }
        if (req.http.X-Cache-SKU !~ "^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$") { return (synth(422)); }
        ban("obj.http.Surrogate-Key ~ " + req.http.X-Cache-SKU);
        return (synth(200));
    }
    if (req.url ~ "^/(health|ready)$") { return (pass); }
    unset req.http.X-Shop-Fragment;
    if (req.url ~ "^/backend") { return (synth(404)); }
    if (req.url ~ "^/fragments/" && req.esi_level == 0) { return (synth(404)); }
    if (req.esi_level > 0) {
        if (req.url !~ "^/fragments/de-DE/(produkt/|kategorie/|suche|$)") { return (synth(404)); }
        unset req.http.Cookie;
        unset req.http.Authorization;
        set req.http.X-Shop-Fragment = "1";
    }
    if (req.method != "GET" && req.method != "HEAD") { return (pass); }
    if (req.url ~ "^/de-DE/suche($|[?])" && req.method != "GET") { return (pass); }
    if (req.http.Cookie || req.http.Authorization) { return (pass); }
    if (req.url ~ "^/de-DE/(warenkorb|kasse|bestellung|profil|gast)") { return (pass); }
    if (req.url ~ "^/de-DE/(produkt/|kategorie/)") { return (pass); }
}

sub vcl_backend_response {
    if (bereq.url ~ "^/de-DE/suche($|[?])") {
        set beresp.grace = 0s;
        set beresp.keep = 0s;
        if (beresp.status != 200 || beresp.http.Cache-Control !~ "public" || beresp.ttl <= 0s) {
            set beresp.uncacheable = true;
            set beresp.ttl = 0s;
            return (deliver);
        }
        if (beresp.ttl > 5s) { set beresp.ttl = 5s; }
    }
    if (bereq.url ~ "^/fragments/") {
        if (beresp.ttl > 5s) { set beresp.ttl = 5s; }
        set beresp.grace = 0s;
        set beresp.keep = 0s;
    }
    if (beresp.http.Content-Type ~ "text/html") { set beresp.do_esi = true; }
    if (beresp.http.Cache-Control ~ "no-store|private" || beresp.http.Set-Cookie) {
        set beresp.uncacheable = true;
        set beresp.ttl = 0s;
        return (deliver);
    }
    if (bereq.url ~ "^/fragments/") {
        if (beresp.http.Cache-Control !~ "public") {
            set beresp.uncacheable = true;
            set beresp.ttl = 0s;
        }
    }
}

sub vcl_hash {
    hash_data(req.url);
    hash_data(req.http.host);
    hash_data(req.http.HX-Request);
}

sub vcl_deliver {
    if (req.esi_level == 0 && resp.http.Content-Type ~ "text/html" && resp.http.Cache-Control !~ "no-store|private") {
        set resp.http.Cache-Control = "public,max-age=0";
    }
    unset resp.http.Server;
    unset resp.http.Via;
}

sub vcl_backend_fetch {
    if (bereq.url ~ "^/fragments/") { set bereq.first_byte_timeout = 150ms; }
}
