# Shop

Entwicklungsgrundlage für einen B2C-Elektroshop: eine Go-Anwendung in einem Root-Modul `github.com/Alps86/shop`, intern als modularer Monolith mit DDD und hexagonalen Ports und Adaptern.

- [Agentenregeln](AGENTS.md): Umsetzung, unabhängige Plan- und Abschlussreview, Nachweise und Squash-Merge.
- [Styleguide](spec/styleguide/README.md): allgemeine Regeln, Go und Cucumber.

Der aktuelle Stand enthält Architekturverträge, Qualitätskonfiguration und minimale Standardanbindung für Gherkin/Godog. Die [arc42-Architektur](spec/architecture/README.md) beschreibt den verbindlichen Fach-, Prozess-, Seiten- und SEO-Vertrag in zwölf Kapiteln; anschließend folgen Fachszenarien in Story #4. Shop-Produktivcode ist noch nicht vorhanden. Syntaxprüfung ist kein fachliches Green.

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
go test ./... -count=1 -coverpkg=./... -covermode=atomic -coverprofile=cover.out
go run github.com/vladopajic/go-test-coverage/v2@v2.19.0 --config=.testcoverage.yml
go test -race ./... -count=1
```

Das [Standardwerkzeug](https://github.com/vladopajic/go-test-coverage) verlangt durch `.testcoverage.yml` je Datei, Package und insgesamt 100 %. Keine Excludes, Overrides oder `coverage-ignore`-Annotationen verwenden. Fachliche Tests bleiben Cucumber-Blackbox-Verträge. Nicht implementierte Spezifikation bleibt als `@spec-only` sichtbar; aktive Implementierungsslices müssen angebunden und aktiviert werden.

**Aussagegrenze:** Native Packageermittlung und Standard-Coverageprüfung erfassen den aktuellen buildfähigen Go-Umfang. Sie sind kein unabhängiger AST-Blockvergleich aller Repositoryquellen oder aller Plattform-/Build-Tag-Kombinationen und kein automatischer Beweis vollständiger Slicezuordnung. Solche zusätzliche Härtung ist offen. Review prüft deshalb insbesondere neue Quellen, Build-Tags, zusätzliche Module und unzulässige `spec-only`-Markierungen. Es gibt keine frei konfigurierbare Bootstrap-Abschaltung.
