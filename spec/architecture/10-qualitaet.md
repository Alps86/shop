# 10 Qualität und Abnahme

Die Spezifikation beschreibt überprüfbare Ziele, noch keine gemessenen Ergebnisse. Breite öffentliche Cucumber-Verträge für Domain/Application; kleine reale Integrationsbasis für PostgreSQL/CDC/OpenSearch/Temporal und HTTP-/Templatekomposition. Höchstens ein kurzer Browser-Smoke, keine Visual-/Screenshotmatrix. Alle fünf Pflichtjobs/100-%-Go-Coverage/Review gemäß [Randbedingungen](02-randbedingungen.md).

| Szenario | Reiz / erwartete Beobachtung |
| --- | --- |
| Gastabschluss | Standard+Rechnung und Express+simulierte Karte erzeugen je genau eine Order, gebundene Snapshotwerte und reservierten Bestand |
| Konkurrenz | Zwei Gäste bei einer verfügbaren Einheit: genau eine Order, andere Unavailable, kein Überbestand |
| Neustart | Worker nach Commit/vor Antwort stoppen, neu starten, gleiche CommandID liefert gleiche Order, Orderworkflow wird nachgeholt |
| Suchprojektion | Preis-/Bestandsänderung committed, echte WAL-CDC-Projektion innerhalb normalem Ziel≤2s sichtbar; alte/duplizierte Version ohne Rückfall |
| Suche aus | Detail/Checkout bleiben funktionsfähig, Suche zeigt verständliche Nichtverfügbarkeit |
| Cacheisolation | Zwei Sessions/Varianten/Sprachkontexte: keine fremden persönlichen Daten, richtige SKU-Pflichtdaten nach Include/HTMX |
| Lieferdemo | Teilmenge versandt/geliefert, Status fortschreitend, Snapshot unverändert, unzulässige Menge abgewiesen |
| SEO | Alter Slug301, Variante SSR/Canonical korrekt, Facetten noindex und teilbarer Zustand |

## Seitenaufruf unter 100 ms

Nutzerziel bleibt **Seitenaufruf <100 ms**. Es wird in zwei getrennten Grenzen berichtet:

1. Vollständig zusammengesetzte HTMLantwort: vom Start des HTTPrequests bis zum letzten HTMLbyte am lokalen öffentlichen Varnish-Einstieg, einschließlich SSI/ESI, kein einzelnes Backendfragment. Primärer Demozielwert p95<100ms für Start, Kategorie, Suche, Produktdetail, Cart, Checkoutansicht, Orderstatus. Mutierende Abschluss-/Zahlungs-/Liefercommands separat ausweisen, keine Anzeigelatenz daraus ableiten.
2. Sichtbarer Browseraufbau: Navigationstart bis LCP der vollständigen Seite einschließlich CSS/Bilder/HTMX-Personalisierung. Ebenfalls Ziel<100ms, aber nur bei tatsächlicher Browsermessung belegt. Maximal kurzer Smoke; bleibt LCP ungemessen oder überschritten, ausdrücklich unbelegt/Zielabweichung berichten, niemals HTTPwert als Ersatz ausgeben.

Messprofil: lokale Compose-Laufzeit auf dokumentiertem Host mit mindestens4 CPU/8GiB RAM, Loopbacknetz ohne künstliche WANlatenz, 30 Products/90 Varianten/5 Kategorien/4 Profile/100 Demoorders. Tatsächliche CPU/RAM/OS/Imageversionen und Datenzahlen mit Ergebnis nennen; abweichender Host nicht als identischer Test ausgeben. Parallelität10, 1000 Requests je Route nach100 Warmup, Wiederverwendung bestehender TCPverbindungen, Kompression dokumentieren. p50/p95/p99 und Fehlerquote ausweisen, nicht Durchschnitt allein.

Warm öffentlich: Rahmen/Fragmente cachehit; warm persönlich: DB/Prozess laufend, keine shared Cacheantwort. Kalt: frischer Rahmen-/Fragmentcache, DB/Index/Worker bereit; separat100 Requests je Route ohne Warmup, Misszeiten ausweisen. Gesamter kalter Dienststart zusätzlich Startzeit, nicht mit kaltem Cache verwechseln. Personalisierte und neutrale Suche separat. Unbekannte/Fehlerseiten keine guten Samples zurechnen. Erwartet keine HTTP5xx im gesunden Profil.

Clientassets minimal: kein externes Fontframework/JSframework; native HTML, kleines eigenes CSS, HTMX lokal gepinnt. Bilder passend skaliert/lazy außerhalb Hauptsicht. Diese Strategie ist kein Leistungsbeleg. Echte Messung folgt im Integrationspaket; HTTP-/Rendererintegration kann ohne umfassende Browserautomation erfolgen. Falls unter100ms nicht erreichbar, genaue Route/Cache-/Netz-/Browsergrenze und Messwerte melden, keine stille Änderung des SLA.
