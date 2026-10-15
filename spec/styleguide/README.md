# Entwicklungs-Styleguide

Diese Regeln gelten gemeinsam für den Shop, eine Anwendung mit einem Go-Modul `github.com/Alps86/shop` im Repository-Root. Sie beschreiben Entwicklung und Nachweise, ohne einen bereits implementierten Shop-Kern oder eine vorhandene fachliche Suite zu behaupten.

| Kapitel | Inhalt |
| --- | --- |
| [Allgemeine Regeln](01-allgemeine-regeln.md) | Verantwortung, Objektmodell, Kontrollfluss und technische Qualitätskategorien |
| [Go](02-go.md) | Typdateien, Fehler, Formatierung, Qlty, Ausführung und dynamische Coverage |
| [Cucumber](03-cucumber.md) | Öffentliche Verträge, Specification by Example, TDD, Tabellen und Testpyramide |

Architektur und Fachmodell werden in der arc42 festgelegt; technische Sprachregeln ersetzen keine Architekturentscheidung. Features entstehen separat nach Abnahme dieser Grundlage. Es gibt keine getrennten App-/UI-Module und keine durch diesen Styleguide festgelegte zusätzliche Client-Komponentenbibliothek.

[Projekteinstieg](../../README.md) · [Agentenregeln](../../AGENTS.md)
