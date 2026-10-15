# Shop

Entwicklungsgrundlage für einen B2C-Elektroshop. Der Shop entsteht als eine Go-Anwendung in einem Root-Modul `github.com/Alps86/shop`, intern als modularer Monolith mit DDD und hexagonalen Ports und Adaptern.

## Einstieg

- [Agentenregeln](AGENTS.md): Umsetzung, unabhängige Review, aktuelle Nachweise und Squash-Merge.
- [Entwicklungs-Styleguide](spec/styleguide/README.md): allgemeine Regeln, Go und Cucumber.

Die Entwicklungsgrundlage enthält Prüftooling und Quality-Gates. Fachmodell und arc42 sowie anschließend die fachlichen Cucumber-Verträge werden in getrennten Arbeitspaketen erarbeitet. Es gibt noch keine Shop-Produktivimplementierung; technische Bootstrap-Prüfungen belegen kein fachliches Verhalten.

## Entwicklung und Prüfung

Voraussetzungen sind die in `go.mod` festgelegte Go-Version, Python 3, Make und Qlty. Alle Befehle laufen im Repository-Root:

```sh
make check
python3 scripts/quality.py selftest
qlty check --all --no-fix --summary
qlty smells --all --json
```

Für Änderungen gegen den Hauptbranch können Qlty-Läufe mit `--upstream origin/main` statt `--all` verwendet werden. [Einzelne Prüfkommandos und Regeln](spec/styleguide/02-go.md#technische-checks) beschreiben Umfang und Grenzen.

GitHub Actions führt für jeden PR die fünf Jobs `static`, `build`, `cucumber-and-coverage`, `race` und `quality-gate` tatsächlich aus. Alle müssen am aktuellen PR-Head erfolgreich sein. Bei dynamisch nachgewiesen leerem Produktivinventar lautet Coverage **nicht anwendbar**. Ab dem ersten Produktivcode gelten echte fachliche Cucumber-Läufe, vollständige dynamisch inventarisierte 100 % Go-Statement-Coverage und Race-Prüfung. Ein Syntaxcheck belegt keine fachliche Implementierung.

## Inventar und Phasenübergang

`scripts/inventory.py` entdeckt jede eigene `.go`-Quelle im gesamten Repository, unabhängig von erwarteten Produktivverzeichnissen. `_test.go` bildet ausschließlich die Testanbindung. Die vier bestehenden AST-Checkerdateien sind als eng begrenztes Prüftooling mit Inhaltfingerprints in `.quality/tooling.json` erfasst; zusätzliche Verzeichnisse können sich dadurch nicht selbst aus der Messung ausnehmen. Änderungen daran verlangen eine Zweckprüfung und aktualisierte Fingerprints. Fachlogik gehört weder in diese Dateien noch in Testanbindung.

`.quality/phase.json` hält den geprüften Zustand fest. Beim ersten Produktivpackage auf `implementation` wechseln und jedes tatsächliche Package einem `@slice-...`-Tag seiner fachlichen Features zuordnen. Das Gate verlangt exakte Packagezuordnung, aktive Verträge ohne `@spec-only`, echte Strict-Godog-Schritte und jeden nativen Coverageblock. Quellen mit Build-Tags, zusätzliche Module und von Go übergangene Quellen scheitern, bis eine vollständige Messmatrix spezifiziert ist. Eine Rückkehr zu leerem Inventar im Implementierungsmodus ist kein gültiger Bootstrap.

Die Testproben entstehen ausschließlich in temporären Verzeichnissen und werden entfernt. Sie prüfen unter anderem echte fehlende Godog-Bindungen, ungültiges Gherkin, vollständige native Coverage, nicht ausgeführte Blöcke und nicht importierte Packages. Sie sind Prüftooling und keine Shop-Fachfeatures. Die vier Teilgates schreiben Ergebnisse mit demselben Inhaltsfingerprint; `quality-gate` verweigert fehlende oder veraltete Nachweise.
