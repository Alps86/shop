# 1 Ziele und Anforderungen

Ein Gast findet ein Modell, wählt Farbe/Speicher, kauft mit Zubehör über echten Temporal-Prozess und sieht danach den simulierten Lieferfortschritt. Der Betreiber demonstriert Preis-/Bestandsänderung und Teillieferung. Das EP-Grundkonzept schafft vertraute Navigation, ohne Händlerplattform oder Marken-/Designkopie.

Priorität: zuerst ein zeigbarer vertikaler Ablauf mit echten Diensten, dann Breite. Leichte HTML-Ausgabe, minimales CSS, klare Varianten und wenige hilfreiche Filter haben Vorrang vor Grafikperfektion. Eine vollständige Produktionsfreigabe ist kein Demo-Versprechen.

## Vollständiges Vertragsinventar

`Bestätigt` stammt aus dem Projektauftrag; `Entschieden` konkretisiert bisherige Vorschläge. Alle verlinkten Verträge gehören zum Demo-Kern, soweit dort nicht ausdrücklich Folgearbeit steht.

| Anforderung | Status und verbindliche Fundstelle |
| --- | --- |
| Product/Variant, Inventory, Cart, User, Order; spezifische IDs; keine Pointer | Bestätigt: [Typen](05-bausteine/type-model.md), [Katalog](05-bausteine/catalog.md), [Handel](05-bausteine/commerce.md) |
| Family/Category/DiscountRule, Revision/Reservierung und genaue Validierung | Entschieden: eigenständige versionierte Konfigurationsroots; [Katalog](05-bausteine/catalog.md), [Handel](05-bausteine/commerce.md) |
| Locale/Land/Währung, Optional, Decimal, typisierte Attribute | Bestätigt und präzisiert: [Katalog](05-bausteine/catalog.md) |
| Hersteller, Sicherheit, Energie, Grundpreis, Elektro/Batterie/Data Act | Bestätigt: [Pflichtdaten](05-bausteine/compliance.md), [Import](08-querschnitt/import.md) |
| Paar-Rabatt, Positionen, Steuern, Versand und Kauf-Snapshots | Bestätigt; konkrete Rundung/Preise entschieden: [Handel](05-bausteine/commerce.md) |
| Konkurrenz, Preisbestätigung, aktive Ersatzwahl und Idempotenz | Bestätigt: [Abschluss](06-laufzeit/checkout.md) |
| Gastprofile, Zugriff, Wiederaufnahme und 7 Tage | Bestätigt; getrennte Sessionfähigkeit entschieden: [Gast](08-querschnitt/guest.md), [Temporal](06-laufzeit/temporal.md) |
| Cart UND Order Temporal; Versionen und externe Abwicklung | Bestätigt; Workflow-/Activityvertrag entschieden: [Temporal](06-laufzeit/temporal.md) |
| PostgreSQL/Outbox/CDC/OpenSearch; Wiederanlauf und Suchausfall | Bestätigt; logisches WAL-CDC ohne Broker entschieden: [Projektion](06-laufzeit/projection.md) |
| Varianten, disjunktive Facetten, Topmerkmale, Vorschläge, Ranking/Marge | Bestätigt; Query-/Zähl-/Gewichtungsvertrag entschieden: [Suche](05-bausteine/search.md) |
| Häufig zusammen gekauft, Profile, Fallback | Bestätigt; synthetische Statistik und Berechnung entschieden: [Empfehlung](05-bausteine/recommendation.md) |
| EP-Seiten/Komponenten, eigener Cart/Checkout, moderner einfacher Auftritt | Bestätigt: [Seitenvertrag](08-querschnitt/pages.md) |
| Hugo-Markdown, Go/HTMX, SSI/ESI, persönliche Cachetrennung | Bestätigt; Nginx-SSI/Varnish-ESI entschieden: [Komposition](08-querschnitt/composition.md) |
| Eigene SEO-URLs, Varianten, Redirects, Facetten | Bestätigt; genaue Pfade entschieden: [Routing](08-querschnitt/seo.md) |
| Ein Betreiber, /backend lokal, echte Compose-Dienste | Bestätigt: [Verteilung](07-verteilung.md), [Seiten](08-querschnitt/pages.md) |
| Unter 100 ms; schnelle öffentliche Tests statt Browsermatrix | Bestätigt; Messbedingungen entschieden: [Qualität](10-qualitaet.md) |
| 100 % Go, Qlty-Standards, unabhängige Reviews und fünf Jobs | Bestätigt: [Randbedingungen](02-randbedingungen.md) |

Es bleiben keine zentralen Demo-Produktentscheidungen offen. Offene produktspezifische Rechts-/Lieferantendaten werden nicht erfunden und sperren betroffene Angebote. Produktionshärtung steht ausdrücklich in [Risiken](11-risiken.md).
