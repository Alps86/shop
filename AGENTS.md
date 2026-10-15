# Verbindliche Regeln für Agenten

## Sprache und Projektgrenzen

Kommunikation, Dokumentation, Issues, PRs und Reviewkommentare werden auf Deutsch ohne Gendern geschrieben. Codebezeichner, Typen, API-Felder und vertragliche Literale bleiben Englisch. Vor Änderungen diese Datei und die relevanten [Styleguide-Kapitel](spec/styleguide/README.md) lesen.

Der Shop ist eine Anwendung mit genau einem Go-Modul `github.com/Alps86/shop` im Repository-Root. Keine getrennten App-/UI-Teilprojekte oder weiteren Go-Module anlegen. Die Architektur folgt einem intern modularen Monolithen mit DDD und hexagonalen Grenzen: Domain, Application/Ports, Adapter. Fachliche Typen enthalten keine Pointer; optionale Werte erhalten explizite Präsenz. Technisches Prüftooling darf notwendige AST-/Framework-Pointer verwenden. Infrastruktur, Domainlogik und UI nicht als Setup-Platzhalter vorwegnehmen.

## Umsetzung und Nachweise

- Zugehöriges Issue einschließlich aktueller Kommentare und die bindende Projektgrundlage vollständig lesen. Alle Akzeptanzkriterien erfassen und vorhandenen Nachweisen zuordnen.
- Nach Scopeklärung vor eigentlicher Umsetzung einen minimalen prüfbaren Start-Commit pushen und einen Draft-PR erstellen und verlinken. Nur die minimale Änderung zur Draft-Eröffnung darf davor erfolgen.
- Lokale Prüfungen einschließlich Qlty vor jedem Push ausführen, soweit verfügbar. Fehlende lokale Werkzeuge ausdrücklich melden; sie ersetzen keine serverseitigen Nachweise.
- Nach jedem nachgewiesen abgeschlossenen Paket die Issue-Checkboxen synchronisieren und den geschriebenen GitHub-Stand zurücklesen. Vor Fortschrittsmeldungen müssen Bericht und Checkliste übereinstimmen.
- Ein Kriterium nur bei tatsächlicher Umsetzung und passendem aktuellen Nachweis abhaken. Konfiguration allein ist kein Ausführungsnachweis. Offenes als `Umsetzung fehlt`, `Nachweis fehlt` oder `Checkbox veraltet trotz belegter Erfüllung` einordnen.
- Anforderungen nicht abschwächen, um offene Kriterien loszuwerden. Einschränkungen und Blocker ausdrücklich berichten.
- Vor Review und Abschluss sämtliche Kriterien erneut prüfen. Issue erst schließen, wenn alle abschlussrelevanten Kriterien erfüllt sind.
- Fachliches TDD folgt [Red → Green → Refactor](spec/styleguide/03-cucumber.md#testgetriebene-entwicklung-mit-cucumber). Keine historischen Red-Protokolle, Red-Quellstände, Snapshots oder ausschließlich dafür angelegten Nachweisdateien verlangen oder committen.

## Unabhängige Review

Für jede Story einen frischen unabhängigen Review-Subagenten einsetzen; der Umsetzer prüft seine eigene Arbeit nicht als unabhängiger Reviewer. Reviewer lesen Issue und Kommentare vollständig und prüfen jedes Kriterium, auch abgehakte, gegen echte Artefakte und aktuelle passende Nachweise. Pro Kriterium `erfüllt`, `nicht erfüllt` oder `unbelegt` mit Fundstelle festhalten. Nicht auf Diff und grüne CI beschränken. Fehlende oder unbelegte Erfüllung ist ein Finding. Findings beheben und anhand des konkreten Befunds gezielt nachprüfen. Freigabe erst nach korrekter Checkliste und vollständigem Nachweis.

## Quality-Gate und Merge

Am aktuellen PR-Head müssen alle fünf Jobs `static`, `build`, `cucumber-and-coverage`, `race` und `quality-gate` tatsächlich erfolgreich abgeschlossen sein, auch bei Dokumentationsänderungen. Fehlende, laufende, übersprungene und fehlgeschlagene Jobs sowie frühere grüne Commits genügen nicht. Draft erst nach vollständiger Abnahme und unabhängiger Freigabe aufheben. Keine Fertigmeldung, Mergefreigabe oder Auto-Merge vorher. Diese Regel gilt auch ohne technische Branchsperre und für jeden Zielbranch.

`make check` führt die lokalen Prüfungen aus. Die einzelnen Aufrufe und der Bootstrapvertrag stehen im [Go-Styleguide](spec/styleguide/02-go.md). Bootstrap bedeutet bei dynamisch nachgewiesen leerem Produktivinventar Coverage **nicht anwendbar**, niemals 100 %. Alle Jobs führen weiterhin tatsächliche technische Prüfungen aus. Ab erstem eigenen Produktivcode sind echte fachliche Cucumber-Ausführung, vollständiger Coverage-Inventarvergleich und Race-Erkennung verbindlich. Fehlende Bindungen, Inventarlücken, fehlende Blöcke oder unzulässiger Bootstrap-Rückfall müssen fehlschlagen. Keine Dummylogik oder künstlichen Tests einführen.

Jeden PR ausschließlich per Squash-Merge zusammenführen (`gh pr merge --squash`); genau ein Commit je PR im Zielbranch. Branches beginnen mit `codex/`.

## Spezifikation und Dokumentation

Aktuelle abgenommene arc42, Fachverträge und Styleguide sind maßgeblich. Vor betroffenen Änderungen die relevanten Kapitel lesen. Historische Issues, PRs oder bestehender Code rechtfertigen keinen Widerspruch zur aktuellen Spezifikation. Offene Entscheidungen ausdrücklich kennzeichnen. Arbeitsnachweise gehören in Issues/PRs; die Spezifikation beschreibt den aktuellen Vertrag eigenständig verständlich.

Die Root-README erschließt die Architektur mit zwölf arc42-Kapiteln. Große Unterkapitel nach eigenständig lesbaren Themen aufteilen; Kapitel 5, 6 und 8 über die README ihres nummerierten Verzeichnisses erschließen. Keine leeren Architektur- oder Featurebäume vorsorglich erzeugen. Die Setup-Story erzeugt noch keine arc42 oder fachlichen Features.

Features liegen unter `spec/features/`, geordnet nach öffentlichem Vertrag: `domain/<fachthema>/`, `application/<fachthema>/`, `adapter/<fachthema>/` und bei tatsächlichem Bedarf `ui/<fachthema>/`. Ein Testadapter verschiebt keine Vertragsgrenze. Der Featureindex `spec/features/README.md` entsteht mit den Features und verlinkt jede Datei mit einem kurzen Vertragssatz; im selben Änderungspaket mitpflegen. Keine fachliche Regel in UI-Szenarien duplizieren.

## Neue Arbeitspakete

Vor jedem neuen Issue und seiner Umsetzung prüft ein unabhängiger Agent den Plan. Größere eigenständig abarbeitbare Pakete bevorzugen; keine kleinen Vorratsissues erstellen. Unabhängige Abschlussreview bleibt zusätzlich erforderlich.
