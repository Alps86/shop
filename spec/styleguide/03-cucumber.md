# Cucumber und ausführbare Spezifikation

Cucumber-Blackbox-Szenarien sind die fachliche ausführbare Spezifikation. [Go](02-go.md) beschreibt technische Anbindung und Messung; das Setup allein implementiert noch keinen Shop-Vertrag.

## Sprache, Struktur und Index

Jede Feature-Datei beginnt mit `# language: de`. Titel, Story-Satz und feste Steptexte werden auf Deutsch ohne Gendern geschrieben; technische Bezeichner und Literale bleiben Englisch. Deutsche Gherkin-Schlüsselwörter wie `Funktionalität`, `Szenario`, `Szenariogrundriss`, `Beispiele`, `Angenommen`, `Wenn` und `Dann` verwenden. Fachliche Dateinamen und Themenverzeichnisse deutsch im Singular, Umlaute als `ae`, `oe`, `ue`, `ss`; technische Go-Dateien und Bezeichner bleiben Englisch.

Pro fachlich zusammenhängender Funktionalität genau eine Datei mit kurzem Story-Satz zu Ziel und Nutzen. Szenarien derselben Funktionalität bündeln, eigenständige Verträge trennen; keine langen Technik- oder Architekturbeschreibungen im Featurekopf.

Features liegen unter `spec/features/`: `domain/<fachthema>/`, `application/<fachthema>/`, `adapter/<fachthema>/`, bei tatsächlichem Bedarf `ui/<fachthema>/`. Maßgeblich ist der beobachtete öffentliche Vertrag; vorbereitende Testadapter ändern ihn nicht. Keine leeren Bäume und keine getrennten App-/UI-Projekte. Der vollständige Index `spec/features/README.md` verlinkt jede vorhandene Feature-Datei mit genau einem kurzen Vertragssatz, nach Themen gegliedert. Bei Änderungen im selben Paket mitpflegen; keine zusätzlichen Indizes pro Szenario.

## Öffentliche Blackbox-Verträge und Specification by Example

Tests liegen getrennt von Produktivcode und beobachten öffentlich vereinbarte Domain-, Application-/Port- oder Adapterverträge. Keine privaten Felder, Methoden oder internen Datenbankschemata zur Testgrenze machen. Die Außengrenze muss kein Netzwerk-Endpunkt sein. Die Spezifikation muss eine erneute Implementierung desselben beobachtbaren Verhaltens ohne den bisherigen Quellcode ermöglichen.

| Gegenstand | Regel |
| --- | --- |
| Verhalten | Vorbedingungen, exakte Eingaben/Ergebnisse, Zustandsübergänge, Fehlersemantik, Invarianten und beobachtbare Nebenwirkungen explizit beschreiben. |
| Zeit | Einheit, Grenzwert und Wirkung vor, genau an und nach einer relevanten Zeitgrenze zeigen; kontrollierte Uhr statt realer Wartezeit. |
| Konfiguration | Verhaltensrelevante Werte mit Default, Einheit, Gültigkeitsbereich und ungültigen Fällen spezifizieren. Keine versteckten Defaults erfinden. |
| Konsistenz | Szenarien, öffentliche Interfaces und arc42 gemeinsam pflegen; offene fachliche Fragen nicht durch Step-Code entscheiden. |
| Unabhängigkeit | Keine unnötige Abhängigkeit von Netzwerk, realer Uhr oder Szenarioreihenfolge. |

Die breite Basis sind Komponentenfälle gegen öffentliche Verträge, ergänzt durch wenige gezielte Integrationsfälle mit echten Adaptern. Beide Ebenen verwenden Cucumber. Keine fachlichen Unit-Tests oder zusätzlichen tabellengetriebenen Unit-Tests. Die minimale Godog-/Gherkin-Standardanbindung ist Frameworkintegration; eigene Quality-Harnessprogramme werden nicht gepflegt.

UI-Verträge prüfen sichtbare Nutzung des Fachvertrags und duplizieren seine Regeln nicht. Eine breite Template-/Render-Basis verwendet explizite DTOs und Fixtures; wenige Browserfälle prüfen Navigation und Browsersemantik. Daraus folgt keine zweite Anwendung und keine zusätzliche Clientbibliotheksentscheidung.

## Testgetriebene Entwicklung mit Cucumber

Neue oder geänderte Fachfunktionalität folgt tatsächlich Red → Green → Refactor:

1. Zuerst ein fachlich konkretes Blackbox-Szenario ausführen und seinen erwarteten fachlichen Fehlschlag beobachten. Fehlende Bindung, Buildfehler oder Infrastrukturausfall zählen nicht als Red.
2. Minimal implementieren, bis der neue und die betroffenen vorhandenen Verträge erfolgreich laufen.
3. Verhaltensbewahrend refaktorieren und die passenden Szenarien erneut ausführen.

Bestehendes Refactoring mit passenden vorhandenen Szenarien beginnen; keine künstlichen neuen Fälle erzwingen. Fehlende Anbindung oder Szenarien ausdrücklich als fehlenden Verhaltensnachweis berichten. Keine gesonderten historischen TDD-Protokolle, Red-Quellstände, Snapshots oder ausschließlich dafür angelegten Nachweisdateien verlangen oder committen.

Reine noch nicht implementierte Spezifikationsfeatures erhalten `@spec-only`. Reguläre rekursive Discovery bleibt bestehen; nur der tatsächliche Verhaltenstest schließt sie über `~@spec-only` aus. Keine Dateilisten oder versteckten Ignore-Pfade. Alle Features einschließlich `@spec-only` werden mit einem echten Gherkin-Parser auf Syntax geprüft. Dieser Parse belegt weder Stepbindungen noch fachliches Green.

Sobald ein kohärenter Slice angebunden wird, dessen `@spec-only` vor dem fachlichen Red entfernen; anschließend Strict-Verhaltenstest, Green und Refactor durchführen. Bereits aktive Verträge bleiben aktiv. Implementierte Slices dürfen nicht unter `@spec-only` versteckt werden. Zusätzliche Tagselektion darf den globalen Ausschluss nur ergänzen, niemals ersetzen. Nach Abschluss einer Implementierungsstory keine dafür eingeführten Spec-only-/Ignore-Tags behalten.

## Sichtbare Daten und Fixtures

Feste, kurze Steptexte benennen Handlung oder Beobachtung. **Alle Datenwerte stehen ausschließlich in Data Tables**, auch einzelne IDs, Fehler, Wahrheitswerte, Anzahlen und leere Ergebnisse. Keine eingebetteten Parameter oder Outline-Platzhalter im Steptext und keine Docstrings als alternativer Datenträger. Outline-Platzhalter nur in Tabellenzellen. Tabellenüberschriften und Platzhalter deutsch; vertragliche API-Felder behalten ihre Schreibweise.

Überwiegend Szenariogrundrisse mit Examples für gleichartige Varianten verwenden; unterschiedliche Abläufe bekommen eigene Szenarien. Gleichartige Feldprüfungen eines Objekts in einer Tabelle bündeln und jede Zeile wirklich prüfen. Datenlose Aktionen benötigen keine künstliche Tabelle. Leere Sammlungen können durch eine reine Kopfzeilentabelle sichtbar sein. Erfolg, Fehler, Zahl null und feste Zustände nicht hinter Steptext oder Step-Code verstecken. Kurze gemeinsame Steps wiederverwenden, ohne unterschiedliche Vertragswirkungen zu vermischen.

Große strukturierte Daten können unter `spec/fixtures/` liegen und über konkrete relative Referenzen in Tabellen geladen werden. Keine absoluten Rechnerpfade. Eingangs- und Erwartungsfixtures unterscheiden; tatsächliches Ergebnis gegen explizite Erwartung prüfen, keine Selbstvergleiche. Kleine Daten bleiben in Tabellen. Szenario und Fixture müssen zusammen den Vertrag verständlich machen.

## Coverage und Mutation

100 % Coverage des gesamten eigenen Produktivcodes aus fachlichen Cucumber-Blackbox-Läufen sind verbindlich. Für Go gilt die dynamische native Statement-Messung und das Standard-100-%-Gate im [Go-Styleguide](02-go.md#bootstrap-und-vollständige-coverage). Nicht abgedeckter benötigter Code verlangt präzisere fachliche Szenarien; unbegründeter Code wird entfernt oder vereinfacht. Keine künstlichen Szenarien, Test-Hooks, Sonderpfade oder Coverage-Ausnahmen.

Regelmäßig und besonders nach relevanten Verhaltensänderungen gezielte verhaltensrelevante Mutationen prüfen: vorhandene Szenarien müssen den Fehler entdecken. Anschließend Originalzustand wiederherstellen. Überlebende Mutationen auf fehlende Assertions oder fachliche Äquivalenz prüfen, echte Fehler nicht pauschal unterdrücken. Umfang und Einschränkungen berichten; vollständige Statement-Coverage allein beweist keine Testwirksamkeit. Ein automatisiertes Mutationstool ist dadurch noch nicht eingerichtet.

## Ausführung und Aussagegrenzen

Die direkten nativen [Prüfaufrufe](../../README.md#werkzeuge-und-lokale-prüfung) und die [Coveragegrenzen](02-go.md#bootstrap-und-vollständige-coverage) sind maßgeblich. Im leeren Bootstrap sind tatsächliche technische Prüfungen und produktive Coverage **nicht anwendbar** nachweisbar. Mit Produktivcode laufen echte gebundene Strict-Cucumber-Verträge, native Coverage mit Dateis-/Package-/Gesamtschwelle 100 % und Race. Syntax ist kein fachliches Green. Review prüft fachliche Vollständigkeit, aktive Slices, Tabellen, Grenzfälle, Index, Buildumfang und Vertragskonsistenz; Standardwerkzeuge allein beweisen keinen vollständigen Repository-/AST-/Slice-Antibypass.
