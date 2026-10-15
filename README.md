# Shop

Entwicklungsgrundlage für einen B2C-Elektroshop: eine Go-Anwendung in einem Root-Modul `github.com/Alps86/shop`, intern als modularer Monolith mit DDD und hexagonalen Ports und Adaptern.

- [Agentenregeln](AGENTS.md): Umsetzung, unabhängige Plan- und Abschlussreview, Nachweise und Squash-Merge.
- [Styleguide](spec/styleguide/README.md): allgemeine Regeln, Go und Cucumber.

Der Fachkern für Katalog, Varianten, Pflichtdaten, Bewertung, Warenkorb und Bestellkopien ist implementiert. Applicationports koordinieren Gastzugriff, tokenfreie Prozesseingänge, Preisbestätigung, Empfehlungen und getrennten Lieferfortschritt. Die [arc42](spec/architecture/README.md) beschreibt den verbindlichen Vertrag in zwölf Kapiteln; die [konkrete Kern-API](spec/architecture/05-bausteine/core-api.md) erschließt die Go-Signaturen. [Cucumber-Verträge](spec/features/README.md) prüfen die aktiven Kernfälle tatsächlich. Die reale Laufzeit wird in [Issue #10](https://github.com/Alps86/shop/issues/10) und die Storefront in [Issue #11](https://github.com/Alps86/shop/issues/11) umgesetzt; der aktuelle Stand ist noch kein vorführbarer Shop. Ein Syntaxparse belegt keine offene Runtimefunktion.


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
go test ./... -count=1 -coverpkg=./... -covermode=atomic -coverprofile=cover.out -args -test.gocoverdir="$SHOP_PROCESS_COVERDIR"
go tool covdata merge -i="$SHOP_PROCESS_COVERDIR" -o="$merged_coverage" -pcombine
go tool covdata textfmt -i="$merged_coverage" -o=cover.out
go run github.com/vladopajic/go-test-coverage/v2@v2.19.0 --config=.testcoverage.yml
go test -race ./... -count=1
```

Das [Standardwerkzeug](https://github.com/vladopajic/go-test-coverage) verlangt durch `.testcoverage.yml` je Datei, Package und insgesamt 100 %. Keine Excludes, Overrides oder `coverage-ignore`-Annotationen verwenden. Fachliche Tests bleiben Cucumber-Blackbox-Verträge. Nicht implementierte Spezifikation bleibt als `@spec-only` sichtbar; aktive Implementierungsslices müssen angebunden und aktiviert werden.

Die aktiven Adapterfälle benötigen echte PostgreSQL-, Temporal- und OpenSearchdienste. `SHOP_TEST_POSTGRES_DSN`, `SHOP_TEST_TEMPORAL_ADDRESS`, `TEMPORAL_ADDRESS` und `SHOP_TEST_OPENSEARCH_URL` benennen deren Testverbindungen. `SHOP_PROCESS_COVERDIR` sammelt zusätzlich die native Coverage der tatsächlich gebauten und gestarteten Shopdatei; das Testframework überschreibt seine eigene Variable `GOCOVERDIR`, deshalb wird die Prozessgrenze ausdrücklich verdrahtet. Die Standardbefehle `covdata merge -pcombine` und `textfmt` führen beide nativen Messungen zusammen. Das Coveragegate enthält dadurch auch `cmd/shop`; keine Ausschlüsse und kein eigener Coveragechecker.

## Laufzeit im Aufbau

`docker compose up --detach --build --wait` startet den technischen Laufzeitstand aus [compose.yml](compose.yml). `cmd/shop` verbindet dieselben Applicationports mit PostgreSQL, echten Temporalworkern, Prozessoutbox und WAL-Projektor. `SHOP_SEED_FILE` importiert vor der Adapterbindung eine versionierte Quelle mit den bestehenden Validierungs-/Noop-/Transaktionsgrenzen. Der aktuelle Container enthält ausdrücklich einen synthetischen technischen Startkatalog; er ersetzt keine verifizierten Pflichtdaten für die abschließende Demo. `GET /health` prüft die Fach-DB, `GET /ready` zusätzlich die tatsächliche Prozessannahme; Suchstörungen bleiben als eigener Diagnosestatus sichtbar und sperren die Kaufbereitschaft nicht. Storefront und separater Backendlistener werden mit #11 und dem anschließenden Integrationspaket verbunden. Der technische Start ist noch kein vollständiger Gastshopnachweis.

Gepinnt sind [PostgreSQL 16.15](https://hub.docker.com/_/postgres), [Temporal CLI 1.9.1 mit echtem Entwicklungsserver](https://github.com/temporalio/cli), [OpenSearch 3.9.0](https://docs.opensearch.org/latest/install-and-configure/install-opensearch/docker/) und [Go 1.26.9](https://hub.docker.com/_/golang). Temporal speichert seine History in einem eigenen SQLitevolume des echten Servers, der Fachzustand liegt in PostgreSQL. Alle veröffentlichten Ports binden Loopback. Das ausdrücklich lokale Demopasswort ist keine produktive Zugangskonfiguration. Persistente Volumes bleiben bei `docker compose down` erhalten; `docker compose down --volumes` löscht ausschließlich die eigenen Demodaten und Prozesse endgültig.

Die aktuelle Entwicklungsmaschine erlaubt keinen Dockerzugriff. Native echte Dienstfälle sind zusätzliche Adapterbelege; der tatsächliche Compose-Start und Restart werden im berechtigten CI-Runner ausgeführt und erst nach erfolgreichem Lauf als erfüllt gewertet.

**Aussagegrenze:** Native Packageermittlung und Standard-Coverageprüfung erfassen den aktuellen buildfähigen Go-Umfang. Sie sind kein unabhängiger AST-Blockvergleich aller Repositoryquellen oder aller Plattform-/Build-Tag-Kombinationen und kein automatischer Beweis vollständiger Slicezuordnung. Solche zusätzliche Härtung ist offen. Review prüft deshalb insbesondere neue Quellen, Build-Tags, zusätzliche Module und unzulässige `spec-only`-Markierungen. Es gibt keine frei konfigurierbare Bootstrap-Abschaltung.
