# Allgemeine Regeln

Diese Regeln gelten zusammen mit [Go](02-go.md) und [Cucumber](03-cucumber.md).

## Lesbarkeit und Verantwortung

- Englisch für Codebezeichner, Deutsch ohne Gendern für Kommunikation und Dokumentation. Namen aus dem Aufrufkontext lesen; unerklärte Abkürzungen und redundante Typnamen vermeiden.
- Kleine kohäsive Methoden, klare Zuständigkeiten und Composition verwenden. Fachliche Regeln auf dem verantwortlichen Typ bündeln. Keine hypothetischen Abstraktionen, Frameworks oder Features vorbauen (YAGNI).
- Konstruktoren beziehungsweise Factories initialisieren Werte und sichern Invarianten. Sonstiges Verhalten als Methoden auf verantwortlichen Typen modellieren; technisch notwendige minimale Einstiegspunkte delegieren.
- Abfragen verändern keinen beobachtbaren Zustand und lösen keine versteckten Commands aus. Commands dürfen sinnvolle Ergebnisse wie Fehler oder IDs liefern.
- Gegen kleine benötigte Interfaces/Ports implementieren. Kein Interface pro trivialem Value Object und keine globale generische Frameworkabstraktion.
- Kommentare erklären nicht offensichtliches Warum und Vertragsgrenzen; keine bloßen Paraphrasen von Namen, auch bei exportierten APIs.

## Kontrollfluss und Größe

Maximal 20 Zeilen pro Funktionskörper im formatierten Ergebnis, einschließlich Leerzeilen und Kommentaren. Signatur und äußere Klammern zählen nicht; echter Code auf kompakten Grenzzeilen zählt. Die Grenze gilt für Funktionen, Methoden und Funktionsliterale ohne Ausnahme.

Maximal zwei verschachtelte Kontrollflussblöcke pro Pfad; der Funktionsrumpf zählt nicht. `for` → `if` ist zulässig, eine dritte Ebene wird sinnvoll in eine Methode ausgelagert. Zusammengesetzte `if`-Bedingungen mit `&&` oder `||` in benannte Prädikatmethoden auf dem verantwortlichen Typ auslagern; keine künstlichen Haltertypen.

Keine `else`-, `else if`- oder `switch`-Zweige. Guard Clauses und frühe Rückgaben verwenden; Exklusivität, Reihenfolge und Seiteneffekte bewahren. Keine komplizierten Ersatzketten: passende kleine Methoden oder gezielte Dispatchstrukturen einsetzen.

Nach einem Kontrollflussblock genau eine Leerzeile, wenn weiterer Code auf derselben Ebene folgt. Vor unmittelbar schließendem umgebenden Block keine zusätzliche Leerzeile. Die [Go-Formatierung](02-go.md#formatierung) konkretisiert dies.

## Maschinenprüfbare Qualität

| Kategorie | Verbindliche Regel |
| --- | --- |
| Fehler | Fehler und prüfpflichtige Rückgaben behandeln; konkrete begründete Vertragsausnahmen nachvollziehbar dokumentieren. |
| Correctness | Toten, ungenutzten oder wirkungslosen Code entfernen; widersprüchliche und verdächtige Konstrukte klären. |
| Security und Dependencies | Keine Secrets versionieren; blockierende Scannerbefunde beheben oder konkret überprüfbar begründen. |
| Maintainability | Komplexität durch sinnvolles Refactoring reduzieren; Duplikation fachlicher Regeln vermeiden, unterschiedliche Konzepte nicht künstlich koppeln. |
| Synchronität | Neue PR-blockierende Entwicklungsregeln gleichzeitig im Styleguide dokumentieren; technische Scannerregeln nicht redundant abschreiben. |

Qlty setzt die maschinenprüfbaren Anteile durch. Technische Scanner ersetzen weder fachliche Verträge noch unabhängige Review. Patterns wie Dependency Injection, Factory, Observer, Facade, Decorator und Strategy nur bei konkretem vereinfachendem Bedarf einsetzen; keine pauschale Patternpflicht. Fachlogik bleibt auf fachlichen Typen, technische Instrumentierung an technischen Grenzen.

Die Projektregeln knüpfen an YAGNI, verständliche Absicht, Vermeidung doppelter Logik und Command Query Separation an. Ihre konkreten Größen- und Ablagegrenzen sind verbindliche Projektentscheidungen.

[Styleguide](README.md)
