# Redaktion und versionsgebundene Seitenrahmen

Hugo baut eigene Markdowntexte, gemeinsame statische Teile und veröffentlichte Produkt-/Kategorierahmen. Der Import erzeugt die Rahmenquellen aus demselben publizierten Katalogsnapshot wie die öffentliche Routenallowlist. Sie werden nicht parallel als unabhängige Stammdaten gepflegt.

Beispiel für `web/content/generated/p-phone.md`; alle Werte stammen aus dem Import:

```json
{"title":"Testmodell","description":"Kurze eigene Beschreibung","url":"/_frame/product/p-phone/1/","layout":"dynamic","canonical":"/de-DE/produkt/testmodell--p-phone","fragment":"/fragments/de-DE/produkt/testmodell--p-phone","frame_kind":"product","root_version":1}
```

Nach dem JSON-Frontmatter folgt nur eigener kurzer redaktioneller Markdowntext. Kategorie entsprechend mit `frame_kind: category`, `/_frame/category/{CategoryID}/{Version}/`, Kategoriencanonical und Kategorienfragment. Produktrevision und Kategorieversion stammen aus dem maßgeblichen Snapshot. Keine SKUpreise, Bestände, Pflichttexte, Sessionfähigkeiten oder Adressen in den Rahmen schreiben.

```sh
hugo --source web --destination /tmp/shop-web --cleanDestinationDir
```

Der vollständige neue Build wird atomar bereitgestellt. Go und Nginx sehen denselben Build unter `Config.ContentDirectory` beziehungsweise dem Nginxroot. Der Go-Router liest zuerst die maßgebliche Identität, validiert Query/SKU und Slug und prüft bekannte Detailausfälle. Erst danach verweist `X-Accel-Redirect` auf den existierenden versionsgleichen internen Rahmen. Fehlende/neue Rahmen, gefilterte Kategorien, persönliche Antworten und HTMX verwenden den funktionsgleichen Go-Rahmen. NginxSSI setzt zuerst Header/Footer und die bereits validierte SKU ein; VarnishESI lädt anschließend das vollständige dynamische Fragment. Es gibt keinen eigenen Includeinterpreter.

Die eigene SVGdarstellung ist neutral und selbst erstellt. HTMX 2.0.11 liegt mit Originallizenz unter `internal/adapter/http/asset/vendor/`; [Primärdokumentation](https://htmx.org/docs/). Styles, eigene Interaktion und eingebettetes HTMX werden mit tatsächlichem Inhaltsfingerprint adressiert. Nach einer Assetänderung werden Datei und referenzierter Fingerprint gemeinsam aktualisiert.

Varnish wird mit `-p vsl_mask=-ReqHeader,-BereqHeader,-RespHeader,-BerespHeader,-ReqUnset,-BereqUnset,-RespUnset,-BerespUnset` gestartet. Sessionfähigkeiten und Set-Cookie erscheinen dadurch nicht in den gemeinsam verfügbaren Headerlogs; Nginx führt kein Accesslog und protokolliert nur kritische Laufzeitfehler.

Öffentliche zusammengesetzte HTMLantworten erhalten am Varnish-Ausgang `public,max-age=0`: Ein Browser speichert dadurch dynamische Bestandteile nicht länger als deren kürzere Gültigkeit. Varnish behält intern die getrennten Rahmen- und Fragmentcaches. Die Fragment-TTL berücksichtigt alle sichtbaren Karten einschließlich Empfehlungen und endet spätestens an deren nächster Pflichtdaten- oder Regelgrenze.

`ProductCard.URL` ist das vollständige vom Applicationport gelieferte Variantenlinkziel einschließlich SKU-Query. Karten, Empfehlungen und aktive Checkoutalternativen übernehmen diese URL unverändert.
