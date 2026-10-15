# 12 Glossar

| Begriff | Bedeutung |
| --- | --- |
| Product | Modell mit gemeinsamen Daten, Root mit mindestens einer Variant, ohne eigene SKU/Livepreise |
| Variant / SKU | Enthaltene verkaufbare Ausprägung; SKU katalogweit eindeutig |
| ProductFamily | Versionierte Attributstruktur mit Typ, Ebene, Optionen, Einheiten/Grenzen |
| Category | Navigationsbaum und kuratierte Filter-/Topmerkmalsauswahl |
| Inventory | Preis und Bestand pro SKU; einzige maßgebliche aktuelle Verkaufswahrheit |
| Cart | Gastbezogene SKU-/Mengenwahl mit Aktivität/Bewertung und Temporalprozess |
| User / Profil | Eigener Gast als Root; vorbereitetes Interessenprofil ist keine geteilte Zugangsidentität |
| Evaluation | Zeit-/revisionsgebundene Positionsbewertung, keine Reservierung oder Kaufgarantie |
| Order | Verbindlicher unveränderlicher Kauf-Snapshot, eigener Temporalprozess |
| Reservation | Beim Abschluss gebundene Menge, beim Versand verbraucht |
| DiscountRule | Versionierte gleiche Paar-/Bestellregel für alle Profile |
| Checkout | Im Cartprozess koordinierte Adresse/Versand/Zahlung/Bestätigung, kein weiterer Root |
| Outbox / CDC | Atomar geschriebene Ereignisse / committed Änderungen aus WAL zur Projektion |
| ProjectionVersion | Monotone Version eines vollständigen Product-Suchdokuments, getrennt von Rootrevision |
| Disjunktive Facette | Zählung mit allen Filtern außer eigener Facette, Werte innerhalb Facette OR |
| Nested | Variantenprädikate treffen dieselbe SKU statt Kreuzkombinationen |
| SSI / ESI | Nginx-statische / Varnish-dynamische serverseitige HTML-Includes |
| HTMX | Leichte progressive Formular-/Fragmentinteraktion auf serverseitigem HTML |
| Locale / Market / Currency | Sprache/Regiontag, Absatzland und Währung als getrennte validierte Kontexte |
| Snapshot / Fulfillment | Unveränderliche Kaufwerte / separat rückgemeldeter Lieferfortschritt |
| Demo | Lokale vorführbare Simulation ohne echte externe Zahlung/Bestellung; keine Produktionsfreigabe |
