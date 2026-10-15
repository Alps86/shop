# Ausführbare Shop-Verträge

Diese Verträge konkretisieren die abgenommene [arc42](../architecture/README.md) und das Epic für die lokale Demo. Sämtliche Domainverträge und die reinen Applicationfälle sind aktiv. Rekursive Strict-Godog-Suites `@value`, `@catalog`, `@sale` und `@app` binden ihre öffentlichen Verträge; der zentrale Gherkinparser prüft zusätzlich jede noch als `@spec-only` markierte Datei beziehungsweise jedes offene Szenario. Syntaxerfolg ist kein fachliches Green. Native Coverage misst den gesamten buildfähigen Kern, ohne Excludes.


## Inventar und Aktivierung

Jede Zeile benennt einen öffentlichen Blackbox-Vertrag, seine maßgebliche Architektur und den kohärenten Aktivierungsslice. Die Tags sind Verantwortungszuordnung, keine Discovery-Allowlist. Domain ist der öffentliche Fachtypvertrag, Application der öffentliche Service-/Portvertrag, Adapter der reale IO-Vertrag, UI die sichtbare Verwendung expliziter DTOs. Testadapter verschieben keine Grenze.

| Feature / Vertrag | arc42 | Slice |
| --- | --- | --- |
| [Produktmodell und Varianten](domain/katalog/produkt.feature) | [05-bausteine/catalog.md](../architecture/05-bausteine/catalog.md) | `core` |
| [Validierte Kennungen und Kontexte](domain/wert/kennung.feature) | [05-bausteine/catalog.md](../architecture/05-bausteine/catalog.md) | `core` |
| [Attributpräsenz und Familientypen](domain/attribut/wert.feature) | [05-bausteine/catalog.md](../architecture/05-bausteine/catalog.md) | `core` |
| [Preis und Bestand einer SKU](domain/handel/inventar.feature) | [05-bausteine/commerce.md](../architecture/05-bausteine/commerce.md) | `core` |
| [Positionen Rabatt und Steuern](domain/handel/bewertung.feature) | [05-bausteine/commerce.md](../architecture/05-bausteine/commerce.md) | `core` |
| [Warenkorbpositionen und Kaufkopien](domain/handel/warenkorb.feature) | [05-bausteine/commerce.md](../architecture/05-bausteine/commerce.md) | `core` |
| [Kontextgebundene verifizierte Pflichtangaben](domain/pflichtangabe/aufloesung.feature) | [05-bausteine/compliance.md](../architecture/05-bausteine/compliance.md) | `core` |
| [Fachliche Gastinteressen](domain/gast/profil.feature) | [05-bausteine/commerce.md](../architecture/05-bausteine/commerce.md) | `core` |
| [Gastkoordination und Fehlergrenzen](application/gast/koordination.feature) | [08-querschnitt/guest.md](../architecture/08-querschnitt/guest.md) | `application`, aktiv |
| [Activityzustände und Portfehler](application/checkout/koordination.feature) | [05-bausteine/core-api.md](../architecture/05-bausteine/core-api.md) | `application`, aktiv |
| [Gastfähigkeit und Interessenprofile](application/gast/zugriff.feature) | [08-querschnitt/guest.md](../architecture/08-querschnitt/guest.md) | `application` |
| [Bestätigter Gastabschluss](application/checkout/abschluss.feature) | [06-laufzeit/checkout.md](../architecture/06-laufzeit/checkout.md) | `application` |
| [Atomarer und idempotenter Kauf](application/checkout/atomaritaet.feature) | [06-laufzeit/checkout.md](../architecture/06-laufzeit/checkout.md) | `application` |
| [Echte langlebige Cart- und Orderprozesse](adapter/prozess/temporal.feature) | [06-laufzeit/temporal.md](../architecture/06-laufzeit/temporal.md) | `runtime` |
| [Begrenzte Liefer- und Teilliefersimulation](application/lieferung/fortschritt.feature) | [06-laufzeit/temporal.md](../architecture/06-laufzeit/temporal.md) | `application` |
| [Atomare Outbox und logisches CDC](adapter/projektion/outbox.feature) | [06-laufzeit/projection.md](../architecture/06-laufzeit/projection.md) | `runtime` |
| [Variantenkonsistente disjunktive Suche](application/suche/facette.feature) | [05-bausteine/search.md](../architecture/05-bausteine/search.md) | `application` |
| [Sortierung Vorschläge und Ausfallgrenzen](application/suche/reihenfolge.feature) | [05-bausteine/search.md](../architecture/05-bausteine/search.md) | `application` |
| [Häufig zusammen gekauft aus Demostatistik](application/empfehlung/zubehoer.feature) | [05-bausteine/recommendation.md](../architecture/05-bausteine/recommendation.md) | `application` |
| [Validierter versionierter Demoimport](adapter/import/publikation.feature) | [08-querschnitt/import.md](../architecture/08-querschnitt/import.md) | `runtime` |
| [Schlichte Shopseiten nach EP-Grundkonzept](ui/seite/struktur.feature) | [08-querschnitt/pages.md](../architecture/08-querschnitt/pages.md) | `frontend` |
| [Zusammenhängende Variantenansicht](ui/produkt/variante.feature) | [08-querschnitt/pages.md](../architecture/08-querschnitt/pages.md) | `frontend` |
| [Durchgängiger eigener Gastpfad](ui/navigation/klickpfad.feature) | [08-querschnitt/pages.md](../architecture/08-querschnitt/pages.md) | `integration` |
| [Hugo SSI ESI und sichere Cachegrenzen](adapter/http/komposition.feature) | [08-querschnitt/composition.md](../architecture/08-querschnitt/composition.md) | `frontend` |
| [Stabile lesbare Routen und Canonicals](adapter/http/route.feature) | [08-querschnitt/seo.md](../architecture/08-querschnitt/seo.md) | `frontend` |
| [Lokale Verwaltung desselben Shops](adapter/http/verwaltung.feature) | [07-verteilung.md](../architecture/07-verteilung.md) | `integration` |
| [Reale Demo und ehrliche Performanceabnahme](adapter/betrieb/demo.feature) | [10-qualitaet.md](../architecture/10-qualitaet.md) | `integration` |

## Datenkonventionen und Zuordnung

Alle Eingaben und Erwartungen stehen in Data Tables; Outlinewerte nur in Tabellenzellen. `absent` bezeichnet explizite Abwesenheit, `empty` eine leere Zeichenfolge, `[]` eine leere Sammlung, `none` keinen vorhandenen Gegenstand beziehungsweise keine Aktion. Diese Testnotation ist keine neue öffentliche JSON-/Go-Repräsentation. Zusammengesetzte Tabellenzellen benennen vollständige kleine Testdaten; kommaseparierte Listen sind geordnete erwartete Listen, sofern der Fachvertrag keine Menge definiert. Geldwerte sind Minor Units, falls nicht ausdrücklich anders benannt. Fehler werden als stabile Codes geprüft, nicht als Meldungstextvergleich. Jeder erwartete Tabellenwert ist zu prüfen; keine ignorierten Spalten.

Sämtliche Alpha-/Beta-/Gamma-Produkte, Kennungen, Adressen, Preise, Regelwerte und statistischen Daten sind **synthetische Testdaten**, keine realen Hersteller-/Sicherheits-/Rechtsangaben. Pflichtdatenmarker stehen ausschließlich für ausdrücklich fiktive Prüfpakete. Ein realer Entkalker mit unvollständigen Lieferantendaten bleibt gesperrt. Keine erfundene Gefahrenklassifizierung oder Rechtskonformitätsgarantie. Kleine Daten stehen vollständig in Tabellen; externe Fixturedateien sind deshalb hier nicht erforderlich, es gibt keine unaufgelösten Fixtureverweise.

Die fünf Verkaufsroots werden durch Produkt-, Inventory-, Warenkorb-/Snapshot- und Gastverträge abgedeckt. Family/Category/DiscountRule bleiben eigene Konfigurationsgrenzen. Pointerfreiheit, spezifische Root-ID-Feldtypen und Abhängigkeitsrichtung sind strukturelle Architektur-/Stilreviewverträge; fachliche Cucumber prüft ihr beobachtbares Verhalten und erfindet keine Runtime-Reflection als Fachtest.

Die Seitenmatrix E1–E6/M1 wird in `ui/seite/struktur.feature` auf gemeinsame Rahmen, Start/Angebote/Markenwelt, Kategorie, Produkt und eigenen Cart/Checkout/Order/Backend abgebildet. E-Kürzel sind die Quellen aus [Seiten](../architecture/08-querschnitt/pages.md), keine externen Testziele. MEDIMAX bleibt ergänzender Vergleich. Keine fremden Warenkörbe, Logins oder Zahlungsflüsse ausführen. Nicht übernommene Standort-/Händler-/Account-/Bewertungsfunktionen bleiben ausdrücklich außerhalb. EP ist Strukturreferenz, das Design eigenständig schlicht und modern.

UI prüft Komponenten, Navigation und DTO-Verwendung. Rabatt-/Steuer-/Mengen-/Suchregeln liegen ausschließlich an Domain-/Applicationgrenzen. UI erhält fertige öffentliche Resultate, keine eigenen Preisalgorithmen. HTTP-/Renderer-/Templatefälle bilden die breite Basis, echte Dienstintegration folgt im Runtime-/Integrationsslice; höchstens ein kurzer HTMX-/History-Browser-Smoke, keine Visual-/Screenshotmatrix.

## Schrittweise echte Ausführung

Core A implementiert alle acht Domainfeatures und reine Applicationverträge. Gastprofil-/Session-/Besitz-/Activityfälle sowie Checkout/Atomicity/Lieferung/Empfehlung sind aktiv; HTTPstatus, CSRF und Cookieattribute bleiben im Gastfeature ausdrücklich szenarioweise `@spec-only` für C. Im Suchfeature sind Request-, Kontext-, Auswahl- und Intervallvalidierung aktiv, tatsächliches Ranking/Autocomplete und Ausfallintegration bleiben `@spec-only` für B. Das Facettenfeature wartet vollständig auf echtes OpenSearch. Atomicity verwendet ausdrücklich Transaktionsdoubles und benötigt zusätzlich reale PostgreSQL-/Temporalabnahme in B.

Die Folgepakete bündeln Runtime B (echte Adapter einschließlich tatsächlicher OpenSearch-Queryausführung), Frontend C und Integration D. Such-/Empfehlungsportverträge dürfen Applicationfälle bleiben; ihre adapterabhängigen Examples werden erst mit Runtime B aktiviert, niemals durch Selbstvergleiche scheinbar erfüllt.

1. `core`: kohärente öffentliche Katalog-/Wert-/Pflichtdaten-/Handelsslices einzeln vollständig anbinden. Vor fachlichem Red das `@spec-only` des tatsächlich implementierten Vertrags entfernen; gemischte Implementierung innerhalb einer Datei verlangt entsprechend sichtbare szenarioweise Markierung der noch fehlenden Teile, keine Verstecke aktiver Regeln.
2. `application`: Gastzugriff, Bewertung/Abschluss, Suche/Empfehlung und Lieferung an die öffentlichen Ports anbinden. Nach stabilen Ports können unabhängige Adapter parallel entstehen.
3. `runtime`: echte PostgreSQL-/pgoutput-/OpenSearch-/Temporaladapter, Prozesswiederanlauf und Import prüfen.
4. `frontend`: vorbereitete öffentliche DTOs rendern; Hugo-/SSI-/ESI-/HTMX-/SEO-Verträge prüfen.
5. `integration`: reale vertikale Gastdemo, Verwaltung, HTTPnavigation und ehrliche Messberichte gegen echte Dienste ausführen.

Jeder kohärente Slice durchläuft fachliches Red → Green → Refactor. Fehlende Stepbindung, Buildfehler oder Dienstausfall ist kein fachliches Red. `Strict: true` und global `~@spec-only` bleiben; aktive Verträge dürfen nie zurück hinter das Tag. Keine Dateiallowlist oder Coverage-Ausnahmen. Alle fünf CI-Jobs laufen am jeweiligen Head; ab produktivem Go gelten native dynamische 100-%-Datei-/Package-/Gesamtcoverage und Race. Der Standardtoolvertrag beweist keine vollständige AST-/Buildtag-/Slice-Antibypasshärtung.

## Demo und spätere Vertiefung

Alle hier konkretisierten Fach-/Seitenverträge sind Demo-Kern. Produktionsbackup/HA, öffentliches TLS-/Secret-/Rate-Limitkonzept, juristische Gesamtfreigabe, exhaustive Ausfallkombinationen, dauerhafte Kontowiederherstellung, echte Zahlung/Fulfillment, Akeneo und Workfloweditor/Migration bleiben Folgearbeit gemäß [Risiken](../architecture/11-risiken.md). Keine breite Browsermatrix oder zusätzliche Perfektionshärtung blockiert die vorführbare Demo. Die Messszenarien definieren das Verfahren, ihre Syntaxprüfung belegt keine Laufzeit und keine SLA-Erfüllung.
