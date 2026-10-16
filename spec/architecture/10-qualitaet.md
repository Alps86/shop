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

## Feste Referenz und Messprofil

Die HTTP-Leistungsabnahme verwendet eine vor dem Lauf festgelegte bekannte Referenzmaschine mit Intel Core i7-13650HX (14 physische Kerne, 20 logische CPUs). Deterministisch gewählt werden die ersten vier verschiedenen physischen Kerne mit den logischen CPUkennungen `0,2,4,6`; die tatsächliche Socket-/Core-/SMT-Zuordnung muss vor Ausführung bestätigt sein. Die gemeinsame Vier-CPU-/16-GiB-Hülle einschließlich Swapverbot und sämtlicher Teilnehmer ist in [Kapitel 7](07-verteilung.md#feste-lokale-vollstackreferenz) festgelegt. Diese Hardware ist keine Behauptung gleicher Rechenleistung wie ein Cloudrunner.

Ein vorab konservierter Runmanifest hält Produkt-/Testcommit, Appbinaryhash und Toolchainbeleg getrennt fest. `images.json` enthält ausschließlich die sieben Imageobjekte mit reference und tatsächlichem amd64Digest. Go 1.26.9, PostgreSQL 16.15, TemporalCLI 1.9.1 mit echtem persistentem Historyserver, OpenSearch 3.9.0 samt gebündelter JVM/Lucene und 1 GiB Heap, Hugo 0.168.0, Nginx 1.30.6 und Varnish 9.1.0 bleiben identisch zum vereinbarten Produktlauf. Die tatsächlich aus der Affinität ermittelten Defaults müssen Go-/Pool-NumCPU 4, PoolMaxConns 4 und OpenSearch allocated_processors 4 ergeben; es gibt keine zusätzliche CPU-, Pool-, Heap- oder Ressourcenerhöhung zur Zielerreichung. Container-/Hostbetriebssystem, CPUtopologie, effektive Grenzen und konkurrierende Last werden zum Ergebnis angegeben. Fehlende Default- oder Hüllennachweise bleiben offen.

Nach funktionalem Vorlauf und unabhängiger Ausführungsfreigabe wird genau ein vorher festgelegtes vollständiges Paar ausgeführt: zuerst native Coverage einschließlich CLI, danach Race, strikt sequenziell und jeweils mit vorab getrennten frischen Referenzzuständen. Keine Wiederholung bis Erfolg, keine Auswahl günstiger Teilberichte und kein verdeckter Zustandsübertrag vom Coverage- in den Racelauf. Ein abgebrochener Lauf bleibt als solcher sichtbar; eine weitere Messung benötigt einen neuen ausdrücklich begründeten Auftrag.

Unverändert bleiben Loopbacknetz ohne künstliche WANlatenz, beide vollständigen Projekte, 30 Produkte/90 Varianten/fünf Kategorien/vier echte Gastprofile/100 echte Demoorders, Parallelität 10 und 1000 warme Requests je Route nach 100 Warmup sowie 100 kalte Requests je Route ohne Warmup. TCPwiederverwendung, dokumentierte Kompression, ursprüngliche Cachepartitionen und persönliche Isolation bleiben erhalten. Alle 16 vollständigen HTMLmessungen sowie acht Origin- und vier Fragmentberichte werden einschließlich p50/p95/p99 und Fehlerquote ausgewiesen. Kaltes Profil bedeutet weiter ausschließlich frische Proxycachepartitionen bei bereiten Datenbanken/Index/Workern, nicht kalter Dienststart. Fristen und Fehlerzählung bleiben unverändert; fehlerhafte Antworten werden nicht zu guten Samples.

Die fünf Cloud-CI-Jobs `static`, `build`, `cucumber-and-coverage`, `race` und `quality-gate` bleiben am aktuellen Head zwingend. Dort laufen weiterhin die vollständigen unveränderten funktionalen, Coverage- und Raceverträge samt ehrlichen 28 Latenzberichten. Sämtliche bisherigen positiven und negativen Ergebnisse bleiben bestehen. Eine bestandene lokale Referenz erlaubt ausschließlich die Aussage „HTTPziel auf der festgelegten Referenz erfüllt“, sofern alle dafür geltenden Werte tatsächlich bestehen; sie erklärt Cloudüberschreitungen nicht zu Erfolgen. Eine externe CPUursache ist nicht bewiesen. Keine Cloud-, Exklusivitäts- oder allgemeine Hardwaregarantie entsteht. Das Referenzziel ist bis zum vollständigen aktuellen Nachweis offen; Browser-LCP bleibt eine gesonderte Messgrenze.

Warm öffentlich: Rahmen/Fragmente cachehit; warm persönlich: DB/Prozess laufend, keine shared Cacheantwort. Kalt: frischer Rahmen-/Fragmentcache, DB/Index/Worker bereit; separat100 Requests je Route ohne Warmup, Misszeiten ausweisen. Gesamter kalter Dienststart zusätzlich Startzeit, nicht mit kaltem Cache verwechseln. Personalisierte und neutrale Suche separat. Unbekannte/Fehlerseiten keine guten Samples zurechnen. Erwartet keine HTTP5xx im gesunden Profil.

Clientassets minimal: kein externes Fontframework/JSframework; native HTML, kleines eigenes CSS, HTMX lokal gepinnt. Bilder passend skaliert/lazy außerhalb Hauptsicht. Diese Strategie ist kein Leistungsbeleg. Echte Messung folgt im Integrationspaket; HTTP-/Rendererintegration kann ohne umfassende Browserautomation erfolgen. Falls unter100ms nicht erreichbar, genaue Route/Cache-/Netz-/Browsergrenze und Messwerte melden, keine stille Änderung des SLA.
