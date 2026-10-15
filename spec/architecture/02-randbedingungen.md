# 2 Randbedingungen

Eine Anwendung, genau ein Go-Modul `github.com/Alps86/shop` im Root. Domain → Application/Ports → Adapter beschreibt die fachliche Aufbaureihenfolge; technische Imports zeigen nach innen. Domain/Application importieren weder Temporal-SDK noch SQL, OpenSearch oder Templates. Storefront und /backend sind Adapter desselben Kerns. Infrastrukturcontainer sind keine zusätzlichen Fachanwendungen.

Deutsch ohne Gendern für Dokumentation und Nutzung; Englisch für Code, Felder und Literale. Fachliche Typen ohne Pointer, echte validierende Konstruktoren und explizites `Optional[T]`. Keine generische Mandanten-/Händlerplattform. Zunächst `de-DE`, `DE`, `EUR`; die Kontexte bleiben getrennt modelliert.

Verbindlich sind [Go](../styleguide/02-go.md) und [Cucumber](../styleguide/03-cucumber.md): keine eigenen Qualityprogramme, Qlty check/smells/metrics tatsächlich ausführen, deklarative AST-Regeln nur für Standardlücken. Kein else/switch, maximal 20 Körperzeilen, höchstens zwei Kontrollflussebenen. Fachliches Red → Green → Refactor über öffentliche Cucumber-Verträge, keine fachlichen Unit-Testparallelwelten.

Alle fünf CI-Jobs `static`, `build`, `cucumber-and-coverage`, `race`, `quality-gate` am aktuellen Head tatsächlich erfolgreich. Bootstrapcoverage ist nicht anwendbar; ab Produktivcode 100 % native Go-Statement-Coverage je Datei/Package/gesamt ohne Ausnahmen. Native Buildscope-Messung beweist keine vollständige AST-/Buildtag-/Sliceinventarhärtung. Aktive Slices bleiben aktiv. Frische unabhängige Review und Squash-Merge pro Paket bleiben Pflicht.

Öffentliche Referenzen nur lesen. Keine fremden Carts ändern, keine Logins, Reservierungen, Zahlungen oder Bestellungen. Echte Zahlungsabwicklung, Kundenkonten, Marketplace, Akeneointegration, Workfloweditor und umfangreiches Fulfillment gehören nicht zum Demo-Kern.
