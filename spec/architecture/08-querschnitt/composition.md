# Hugo, Includes, HTMX und Cache

Gewählte Standardengines: **Nginx SSI** für gemeinsame statische Header/Footer und **Varnish ESI** für Katalog-/Detail-/Angebotsfragmente. Hugo rendert kurze eigene Markdowntexte, Navigation und Seitenrahmen voraus. Go html/template rendert dynamische DTOs. Nginx verarbeitet zuerst SSI; Varnish löst danach ESI-Marker auf. Komponiertes HTML, nicht rohe Includes, erreicht den Browser. Kein eigener Includeinterpreter.

Hugo erzeugt Rahmen pro publizierter Kategorie/ProductID und Editorialroute; ein Import-/Publikationsschritt regeneriert sie atomar. Routing löst stabile ID/aktuellen Slug gegen DB auf, kann bei neuem Produkt bis zum Build einen funktionsgleichen Go-Rahmen liefern. HTTP-Metadaten/Canonical vor Ausgabe korrekt; keine zufällige Cacheübernahme eines alten Slugs. Dynamische Pflichtdaten befinden sich im vollständigen Detail-/Kartenfragment, nicht im lang gecachten Markdown.

| Klasse | Schlüssel und TTL | Invalidierung / Fehler / Stale |
| --- | --- | --- |
| Statische Assets | Contenthash; 1 Jahr immutable | Neuer Hash bei Änderung; fehlend404, kein stiller Ersatz von JS/CSS |
| Markdownrahmen/Header/Footer | Route+Locale+BuildVersion; 1h | Atomarer neuer Build/Version; maximal24h stale nur redaktioneller Content |
| Öffentliche Kategorienavigation | Locale+Market+CatalogVersion; 5min | Katalogpublikation bump Version; maximal60s stale |
| Angebots-/Listing-/Detailfragment | ID/Query+Locale+Currency+Market+SKU+Projection/RootVersion; höchstens5s | Änderung invalidiert Product-Surrogatekey, TTL ist Backup. Kein stale von Preis/Bestand/Pflichtdaten über5s; bei Timeout sichere Ersatzdarstellung |
| Personalisierte Suche/Empfehlung | Nicht shared; Cache-Control private,no-store | Gastcookie erzwingt pass oder explizit unpersonalisierter öffentlicher Querypfad; nie Profilwerte im Key/URL |
| Cart/Checkout/Order/Profilumschalter | no-store, keine ESI-Shareroute | Direkte persönliche HTTPantwort, Fehler bleibt persönlich |
| Statischer Rahmen mit Cartanzeige | Rahmen shared ohne Namen/Mengen | Persönlichen Cartbadge nach Laden HTMX separat no-store lesen; Ausgangslink „Warenkorb“ ohne fremde Anzahl |

Kontext normalisieren/validieren vor Cachezugriff. Nur explizite Allowlist öffentlicher Fragmente darf shared gecacht werden, Cookies/Authorization nicht blind weiterreichen. Pflichtdatenkontext SKU/Markt/Sprache/Zeitversion in Fragmentkeys. Veröffentlichte zeitabhängige Regeln begrenzen TTL zusätzlich bis zum nächsten Gültigkeitswechsel. Fragmentausfall darf keine alte SKU-/Pflichtdatenkombination zusammenkleben.

Fragmenttimeout150ms, maximal ein interner Abruf pro Fragmentklasse pro Seitenzusammensetzung, keine rekursiven Includes. Private Seiten umgehen ESIcache. Varnish unterstützt nur eine ESI-Teilmenge, deshalb kein `alt`-Versprechen: Der Go-Fragmentendpunkt liefert bei fachlicher/technischer Nichtverfügbarkeit selbst sicheres 200 Ersatz-HTML „Daten vorübergehend nicht verfügbar“, keine Kaufaktion/kein ungeprüfter Preis, Link zum erneuten Laden und no-store. Der vollständige Detailrouter prüft vor Ausgabe der Parentheader die Verfügbarkeit des Pflichtdatenpakets; bei bereits bekanntem Ausfall503 mit Retry-After2. Fällt erst danach ein ESI-Subrequest aus, kann der Parentstatus nicht rückwirkend geändert werden: optionaler Include mit `onerror=continue` entfällt, ein statischer benachbarter Ladefehlerhinweis bleibt sichtbar; die Kaufaktion liegt ausschließlich im erfolgreichen vollständigen Fragment. Diesen degradierten200 separat beobachten, niemals als kaufbaren Erfolg zählen. SearchUnavailable als503/lesbarer Hinweis, Kategoriebaum/Detail direkt weiter bedienbar. Keine sensiblen Fehlerdetails. Vollständige Fehlseiten no-store.

HTMX verbessert gewöhnliche Formulare/Links: alle zentralen Aktionen funktionieren über normales HTML POST/Redirect/GET. `HX-Request` liefert denselben sichtbaren Vertrag als Teilfragment, Vary: HX-Request; niemals Fragment und Ganzseite unter gleichem ungegliedertem Cacheobjekt. Variantenwechsel aktualisiert Galerie/Preis/Verfügbarkeit/Pflichtdaten/Zubehör zusammen, URL über History ersetzen/übernehmen. Ein Ladeindikator deaktiviert betroffene Aktion, Fehler bleibt mit Eingaben/Fokus/Retry sichtbar. Filterform GET erhält teilbaren Zustand und Browserzurück. Keine schwere JS-App.

Server setzt escaped HTML, Formular-CSRF-Token und SameSite-Cookie. CSP/weitere produktive Browserhärtung kann später vertieft werden; sichere Gasttrennung und kein gemeinsames Cachen persönlicher Daten bleiben Demobaseline. [Qualität](../10-qualitaet.md) misst das vollständige zusammengesetzte HTML separat von Browseranzeige.

Standardenginegrenze: [Varnish ESI-Teilmenge](https://varnish-cache.org/docs/7.4/users-guide/esi.html). Originfallback und Parentstatusentscheidung werden über HTTP-/Rendererverträge geprüft.
