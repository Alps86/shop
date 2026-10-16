# Shop

Ein eigener lokaler Demonstrationsshop mit Gastbestellungen, echten Temporal-Prozessen, PostgreSQL und OpenSearch. Die [Gesamtintegration #14](https://github.com/Alps86/shop/issues/14) dokumentiert die unabhängige Abnahme, aktuellen Ausführungsnachweise und Leistungsgrenzen.

Entwicklungsgrundlage für einen B2C-Elektroshop: eine Go-Anwendung in einem Root-Modul `github.com/Alps86/shop`, intern als modularer Monolith mit DDD und hexagonalen Ports und Adaptern.

- [Agentenregeln](AGENTS.md): Umsetzung, unabhängige Plan- und Abschlussreview, Nachweise und Squash-Merge.
- [Styleguide](spec/styleguide/README.md): allgemeine Regeln, Go und Cucumber.

Der Fachkern für Katalog, Varianten, Pflichtdaten, Bewertung, Warenkorb und Bestellkopien ist implementiert. Applicationports koordinieren Gastzugriff, tokenfreie Prozesseingänge, Preisbestätigung, Empfehlungen und getrennten Lieferfortschritt. Die [arc42](spec/architecture/README.md) beschreibt den verbindlichen Vertrag in zwölf Kapiteln; die [konkrete Kern-API](spec/architecture/05-bausteine/core-api.md) erschließt die Go-Signaturen. [Cucumber-Verträge](spec/features/README.md) prüfen die aktiven Kernfälle tatsächlich. Die reale Laufzeit [#10](https://github.com/Alps86/shop/issues/10) und die Storefront [#11](https://github.com/Alps86/shop/issues/11) sind integriert. Die Gesamtintegration [#14](https://github.com/Alps86/shop/issues/14) dokumentiert die aktuelle unabhängige Abnahme und die zugehörigen Ausführungsnachweise.


## Werkzeuge und lokale Prüfung

Go gemäß `go.mod`, Qlty 0.651.0, Git und jq bereitstellen. Node 24/npm wird ausschließlich für den Standard-AST-grep-Regeltest verwendet; es gibt kein separates UI-Projekt. Keine eigenen Shell-, Go- oder Python-Prüfprogramme werden gepflegt. GitHub Actions führt die Standardwerkzeuge direkt aus.

```sh
qlty config validate
qlty check --all --no-cache --no-fix --summary
qlty smells --all --json
qlty metrics --all --functions
qlty metrics --all --dirs
npm exec --yes --package=@ast-grep/cli@0.40.0 -- ast-grep test --skip-snapshot-tests
git diff --check
git diff --cached --check
git diff --check origin/main...HEAD
go mod verify
go build ./...
go test ./... -run '^$' -count=1
go test ./... -run '^TestFeatureSyntax$' -count=1
go test -race ./... -run '^TestFeatureSyntax$' -count=1
```

Vor Push Qlty-Checks, Smells und Metrics tatsächlich prüfen. Smells müssen ohne offene Befunde sein. [Schwellen und Abnahmewirkung](spec/styleguide/02-go.md#qlty-checks-smells-und-metrics) sind verbindlich. PRs führen unabhängig von Draftstatus und Zielbranch alle fünf Jobs `static`, `build`, `cucumber-and-coverage`, `race` und `quality-gate` aus. Der letzte Job akzeptiert nur vier erfolgreiche Teiljobs am aktuellen Head.

## Bootstrap und Produktivcode

`go list -f '{{if or .GoFiles .CgoFiles}}{{.ImportPath}}{{end}}' ./...` bestimmt die aktuell buildfähigen Packages mit Nichttestquellen. Im Bootstrap ist diese Liste leer. Trotzdem laufen Build, Gherkin-Parse, Race-Prüfung der vorhandenen Standardtestanbindung sowie Qlty-Checks/Smells/Metrics wirklich. Produktive Coverage wird ausdrücklich als **nicht anwendbar** ausgewiesen; keine 100-%-Behauptung und keine Dummylogik.

Ab dem ersten Produktivpackage führt CI die echten Tests einschließlich der strikten Godog-Anbindung aus und misst alle vom aktuellen Go-Build erfassten Packages dynamisch:

```sh
export SHOP_PROCESS_COVERDIR="$(mktemp -d)"
merged_coverage="$(mktemp -d)"
go test -p 1 ./... -count=1 -coverpkg=./... -covermode=atomic -coverprofile=cover.out -args -test.gocoverdir="$SHOP_PROCESS_COVERDIR"
go tool covdata merge -i="$SHOP_PROCESS_COVERDIR" -o="$merged_coverage" -pcombine
go tool covdata textfmt -i="$merged_coverage" -o=cover.out
go run github.com/vladopajic/go-test-coverage/v2@v2.19.0 --config=.testcoverage.yml
go test -p 1 -race ./... -count=1
```

Das [Standardwerkzeug](https://github.com/vladopajic/go-test-coverage) verlangt durch `.testcoverage.yml` je Datei, Package und insgesamt 100 %. Keine Excludes, Overrides oder `coverage-ignore`-Annotationen verwenden. Fachliche Tests bleiben Cucumber-Blackbox-Verträge. Nicht implementierte Spezifikation bleibt als `@spec-only` sichtbar; aktive Implementierungsslices müssen angebunden und aktiviert werden.

Die vollständigen nativen Dienstläufe führen Packages mit dem Standardparameter `-p 1` nacheinander aus. Damit konkurrieren mehrere unabhängig startende Integrationssuites auf dem begrenzten Runner nicht gleichzeitig um dieselbe Suchinstanz; die echten konkurrierenden Transaktions- und Workflowfälle innerhalb der Suites bleiben aktiv. Das öffentliche Suchbudget beträgt 500 ms.

Die aktiven Adapterfälle benötigen echte PostgreSQL-, Temporal- und OpenSearchdienste. `SHOP_TEST_POSTGRES_DSN`, `SHOP_TEST_TEMPORAL_ADDRESS`, `TEMPORAL_ADDRESS` und `SHOP_TEST_OPENSEARCH_URL` benennen deren Testverbindungen. `SHOP_PROCESS_COVERDIR` sammelt zusätzlich die native Coverage der tatsächlich gebauten und gestarteten Shopdatei; das Testframework überschreibt seine eigene Variable `GOCOVERDIR`, deshalb wird die Prozessgrenze ausdrücklich verdrahtet. Die Standardbefehle `covdata merge -pcombine` und `textfmt` führen beide nativen Messungen zusammen. Das Coveragegate enthält dadurch auch `cmd/shop`; keine Ausschlüsse und kein eigener Coveragechecker.

## Lokale Demo starten und bedienen

Voraussetzungen: Docker Engine mit bereits erlaubtem Benutzerzugriff und aktuellem Compose (Override-Tags benötigen Compose ≥2.24.4), freie Loopbackports 8080/8081/15432/17233/18233/19200 sowie mindestens vier CPU und acht GiB RAM. Keine Gruppen-, Socket- oder Daemonrechte ändern. Die eigene Entwicklungsmaschine wird zusätzlich mit eigenen nativen Diensten geprüft; den vollständigen Containerbeleg liefert der berechtigte CI-Runner aus dem jeweiligen PR-Head.

```sh
docker compose up --detach --build --wait --wait-timeout 180
```

Öffentlicher Shop: [http://localhost:8080/de-DE/](http://localhost:8080/de-DE/). Eigene Verwaltung: [http://localhost:8081/backend](http://localhost:8081/backend). Der öffentliche Proxy sperrt `/backend`; die separate Verwaltung ist ausschließlich für diese lokale Demo und besitzt keine produktive Betreiberanmeldung. Die [Datenquellen](spec/demo/datenquellen.md) erklären die vier realen Modelle, acht eigenen Demo-SKUs, Pflichtinformationen und die Nutzungsgrundlage. Preise, Bestände, Kaufkosten und Co-Purchase-Werte sind fiktiv. Die Grafik ist eine eigene neutrale Illustration.

Zum Vorführen über „Technik entdecken“ die Smartphoneskategorie öffnen, Farbe und Speicher wählen und dem Link zur passenden SKU folgen. Beim Fairphone 5 lässt sich die schwarze passende Hülle mit einem festen Paarrabatt von 5 EUR hinzufügen; Vorschau, Ablehnen und Entfernen verändern den Warenkorb nachvollziehbar. Über „Interessenprofil“ lassen sich neutral, mobile, home und coffee auswählen. Ein Browserprofil hält seinen Gastwarenkorb über sein eigenes Cookie; ein privates Fenster erhält einen getrennten Gastzugriff.

Im Warenkorb Standard oder Express und Rechnung oder simulierte Karte auswählen, aktuelle Preise bestätigen und die Bestellung abschließen. Es gibt keine reale Zahlung oder Fremdbestellung. Die Kundenseite zeigt die unveränderlichen Kaufdaten und getrennten Lieferfortschritt. In der Verwaltung Bestand/Preis ändern oder eine Bestellung öffnen und Teilmengen als versandt beziehungsweise geliefert markieren. Eine Preisänderung vor dem Abschluss verlangt eine neue Bestätigung. Nachträgliche Änderungen überschreiben keine Kaufkopie.

`cmd/shop` verbindet dieselben Applicationports mit PostgreSQL, echten Temporalworkern, Prozessoutbox und WAL-Projektor. Der einmalige `content`-Befehl importiert denselben validierten Seed und erzeugt aus dem aktuellen DB-Snapshot die versionsgebundenen Hugo-Markdown-Rahmen. Hugo baut sie einmalig; Nginx verarbeitet SSI und Varnish ESI. Eigene CSRF- und PURGE-Schlüssel sind für die lokale Demo ausdrücklich konfiguriert. Sichere Cookies werden aus einer HTTPS-PublicOrigin abgeleitet; die Loopbackdemo verwendet HTTP. Der Cache quittiert keine CDC-Projektion vor erfolgreicher Suchprojektion und begrenzter Invalidierung der bekannten Demo-SKUs.

`GET /health` prüft die Fach-DB, `GET /ready` zusätzlich die tatsächliche Prozessannahme. Suchstörungen bleiben separat sichtbar und sperren Detail/Kauf nicht. Diagnoseanfragen werden nicht geteilt gecacht. Nach einer Verwaltungsänderung zeigt der echte Index regulär innerhalb des 2-s-Beobachtungsbudgets neue Werte. Öffentliche Fragmente haben höchstens 5 s TTL. Eine bereits vor dem Indexrefresh laufende Suchantwort kann nach einer BAN-Invalidierung erst neu im Cache erscheinen und bis zu dieser TTL alt bleiben; eine allgemeine synchrone Cachekonsistenz wird nicht zugesichert. Der finale Kauf prüft PostgreSQL-Preise und Bestand erneut.

```sh
# Geordneter Stop; Daten und Temporal-History bleiben erhalten.
docker compose down
# Derselbe persistente Shop startet wieder.
docker compose up --detach --wait --wait-timeout 180
# Ausschließlich die eigenen lokalen Demodaten endgültig zurücksetzen.
docker compose down --volumes
docker compose up --detach --build --wait --wait-timeout 180
```

Gepinnt sind PostgreSQL16.15, TemporalCLI1.9.1, OpenSearch3.9.0, Go1.26.9 sowie Hugo0.168.0, Nginx1.30.6 und Varnish9.1.0 (die drei Webengines mit überprüftem Image-Digest). Temporal speichert History in seinem eigenen SQLitevolume; Fachzustand liegt in PostgreSQL. Alle veröffentlichten Ports binden Loopback. Das lokale Demopasswort und die ausgeschriebenen Demokeywerte sind keine produktive Zugangskonfiguration. OpenSearch verwendet einen festen Java-Heap von 1 GiB. Das öffentliche Suchbudget beträgt 500 ms.

## Reproduzierbare Messlaufzeit

Der kleine quellengeprüfte Vorführkatalog und das ausdrücklich synthetische Referenzprofil werden getrennt gemessen. Letzteres enthält 30 Produkte, 90 Varianten, fünf Kategorien, vier echte Gastprofile und 100 über echte Kaufports/Temporal erzeugte Orders. Keine Behauptung 30 recherchierter Herstellermodelle. Der folgende zweite Stack verwendet denselben App-/Proxybuild und eigene Volumes/Ports:

```sh
docker compose -p shop-benchmark -f compose.yml -f compose.benchmark.yml up --detach --build --wait --wait-timeout 180
```

Seine öffentlichen Ports sind 28080/28081; der direkte Go-Origin für getrennte Diagnosemessungen liegt auf Loopback-Port28090. Weitere Ports sind PostgreSQL25432, Temporal27233/UI28233 und OpenSearch29200. Für den unabhängigen Messlauf:

```sh
GOBIN=/tmp/shop-standard-tools go install github.com/rakyll/hey@v0.1.5
export SHOP_HEY=/tmp/shop-standard-tools/hey
export SHOP_PERFORMANCE_PUBLIC_URL=http://127.0.0.1:28080
export SHOP_PERFORMANCE_BACKEND_URL=http://127.0.0.1:28081
export SHOP_PERFORMANCE_INTERNAL_URL=http://127.0.0.1:28090
export SHOP_PERFORMANCE_POSTGRES_DSN='postgres://shop:local-demo-only@127.0.0.1:25432/shop?sslmode=disable'
export SHOP_PERFORMANCE_SCHEMA=demo_benchmark
export SHOP_PERFORMANCE_TEMPORAL_ADDRESS=127.0.0.1:27233
export SHOP_PERFORMANCE_TASK_QUEUE=demo_benchmark
export SHOP_PERFORMANCE_OPENSEARCH_URL=http://127.0.0.1:29200
export SHOP_PERFORMANCE_SEARCH_INDEX=shop-benchmark
export SHOP_PERFORMANCE_COMPOSE_PROJECT=shop-benchmark
export SHOP_PERFORMANCE_STATE_FILE=/tmp/shop-benchmark-private-state.json
export SHOP_PERFORMANCE_REPORT_DIRECTORY=/tmp/shop-benchmark-results
export SHOP_PERFORMANCE_IMAGES='PG16.15 TemporalCLI1.9.1 OS3.9.0 Go1.26.9 Hugo0.168.0 Nginx1.30.6 Varnish9.1.0; Digests in compose.yml'
go test ./cucumber/performance -count=1
```

Die private Zustandsdatei gehört genau zu diesem Stack; nach einem Volume-Reset auch diese Datei entfernen. Die beiden Compose-Parameterdateien und derselbe Projektname gelten ebenfalls für Stop/Reset des Benchmarkstacks.

Die Performance-Cucumber-Suite benötigt zusätzlich `hey` v0.1.5 und `curl` mit Parallelunterstützung. Sie misst das vollständige zusammengesetzte HTML mit Parallelität10, 100 Warmup/1000 warme und 100 kalte Anfragen pro Route. Frische Host-Cachepartitionen pro kalter Anfrage sowie echte Hit-/Miss-/Coalescing-Zähler unterscheiden kalte Fragmente von bloßer Batchleerung; private no-store-Seiten haben keinen gemeinsamen Kaltcache. Das kalte Profil leert ausschließlich den Proxycache pro Anfrage; Dienste, Index und OpenSearch-Requestcache bleiben bereit. Der Suchadapter nutzt den nach Indexrefresh invalidierten Standardrequestcache ausdrücklich auch für vollständige Trefferseiten. Ein privater Fixturezustand mit Modus0600 hält die Testgast-Capabilities außerhalb der Messberichte. Vollständige Testumgebung und tatsächliche Berichte werden in der Integrationsabnahme ausgewiesen.

Die anonyme öffentliche Suchseite rendert ihr Suchergebnis auf einem Cachemiss einmal vollständig mit SSI-Kopf/-Fuß. Varnish teilt erfolgreiche öffentliche Antworten höchstens fünf Sekunden, bei früherer Pflichtdaten-/Regelgrenze entsprechend kürzer. Ein noch gültiger Treffer bleibt bei inzwischen ausgefallenem Origin HTTP200; nach Ablauf, Ban oder Miss wird die echte Suche gefragt und ein Fehler als HTTP503/no-store ausgegeben. Es gibt keinen Stale-Fallback und keinen Fehlercache. Bestätigte Projektionen invalidieren über einen festen technischen Suchtag auch zuvor leere Ergebnisse; die genannte In-flight-BAN-Grenze bleibt bestehen. Checkout prüft Preise und Bestand erneut aus PostgreSQL.

Die warme anonyme Suchmessung belegt tatsächliche gemeinsame Hits, Misses und Coalescings. Hundert kalte Suchanfragen verwenden unverändert hundert zuvor unbenutzte Hostpartitionen und müssen genau hundert SSR-Misses ohne Hits/Coalescing sowie vollständige Antworten und TCP-Wiederverwendung belegen. Persönliche Such-, Cart-, Checkout- und Orderrouten umgehen den gemeinsamen Cache weiterhin; dort müssen alle gemeinsamen Cachezähler null bleiben. Cookies und Authorization erzwingen PASS, Host, vollständige Query und HX-Request trennen die öffentlichen Objekte. Datenumfang, Parallelität und sämtliche warmen/kalten Anfragen bleiben unverändert; der OpenSearch-Requestcache wird nicht zurückgesetzt.

HTTPzeiten sind kein Browser-LCP-Beleg. Abweichungen vom p95-Ziel unter100ms und ungemessener Browseraufbau bleiben ausdrücklich sichtbar. Der kurze Browser-Smoke ersetzt keine Lastmessung und keine Browsermatrix. Mutierende Kauf-/Zahlungs-/Liefercommands zählen nicht zur Anzeigelatenz.

Auftrag und Nachweise zum HTTP-p95-Ziel stehen im [Performanceissue #16](https://github.com/Alps86/shop/issues/16).

**Aussagegrenze:** Native Packageermittlung und Standard-Coverageprüfung erfassen den aktuellen buildfähigen Go-Umfang. Sie sind kein unabhängiger AST-Blockvergleich aller Repositoryquellen oder aller Plattform-/Build-Tag-Kombinationen und kein automatischer Beweis vollständiger Slicezuordnung. Solche zusätzliche Härtung ist offen. Review prüft deshalb insbesondere neue Quellen, Build-Tags, zusätzliche Module und unzulässige `spec-only`-Markierungen. Es gibt keine frei konfigurierbare Bootstrap-Abschaltung.

Die acht direkten Go-Antwortzeiten und vier echten Fragmentzeiten werden zusätzlich getrennt erfasst. Neutrale Kategorie- und Produktantworten am Go-Origin delegieren mit leerem Body und `X-Accel-Redirect`; deren Antwortzeit ist keine HTML-Seitenzeit. Diese Diagnosewerte ersetzen keine der 16 vollständigen HTML-Messungen. Die CI protokolliert außerdem den frischen Start beider Compose-Projekte bis zu gesunden Diensten in eigenen JSON-Berichten, einschließlich Imageabruf, Build, Inhaltsgenerierung und Seedimport. Der zweite Stack nutzt dabei den Buildcache des ersten. Bereits laufende native Infrastruktur belegt keinen vollständigen kalten Dienststart. Browser-LCP und eine durchgängig erreichte Vollseitenzeit unter100ms bleiben damit gesondert zu bewerten.
