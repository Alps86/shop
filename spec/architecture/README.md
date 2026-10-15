# Architektur des Shop-Kerns

Diese arc42 beschreibt den verbindlichen Zielvertrag des eigenständigen B2C-Elektroshops. Die Anwendung ist noch nicht implementiert. Ein Betreiber verkauft Smartphones, Haushaltsgeräte und Zubehör; EP dient als Referenz für Struktur und Abläufe, das Design bleibt eigenständig, schlicht und modern. Gastkäufe und Lieferungen sind ausdrücklich simuliert, ohne fremde Bestellung oder echte Zahlung.

| Kapitel | Vertrag |
| --- | --- |
| [1 Ziele](01-ziele.md) | Demo, Stakeholder, Prioritäten und Anforderungsinventar |
| [2 Randbedingungen](02-randbedingungen.md) | Ein Modul, Sprache, Qualitätsbaseline und Grenzen |
| [3 Kontext](03-kontext.md) | Benutzer, Systeme und externe Abwicklungsgrenze |
| [4 Strategie](04-strategie.md) | DDD, Ports, verbindliche Technologien und Umsetzung |
| [5 Bausteine](05-bausteine/README.md) | Typen, Aggregate, Ports, Suche und Pflichtdaten |
| [6 Laufzeit](06-laufzeit/README.md) | Cart, Checkout, Order, Reservierung und CDC |
| [7 Verteilung](07-verteilung.md) | Echte Compose-Dienste und Betriebsgrenzen |
| [8 Querschnitt](08-querschnitt/README.md) | Seiten, Includes, Cache, SEO, Gastzugriff und Import |
| [9 Entscheidungen](09-entscheidungen.md) | Begründete Ergänzungen zum bestätigten Modell |
| [10 Qualität](10-qualitaet.md) | Überprüfbare Qualitätsszenarien und Performance |
| [11 Risiken](11-risiken.md) | Offene Nachweise und spätere Härtung |
| [12 Glossar](12-glossar.md) | Einheitliche Fachbegriffe |

Die Fachtypen sind Datenverträge; Go-Packagenamen und konkrete Adapterdateien folgen der Implementierung. Fachszenarien entstehen anschließend in eigener Story. Die Spezifikation behauptet weder fachliches Green noch SLA-Erfüllung. [Styleguide](../styleguide/README.md) und [Agentenregeln](../../AGENTS.md) gelten ergänzend.
