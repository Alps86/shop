# Suche und kuratierte Merkmale

OpenSearch ist ein abgeleitetes Product-Lesemodell. Ein Dokument pro Product, `variants` als nested Objekte mit SKU, typisierten Attributen, Inventorypreis/-verfügbarkeit, Bild und Pflichtdatenzusammenfassung. Gemeinsame Productattribute/Brand/Kategorien auf Dokumentebene; keine flache Kreuzkombination mehrerer Varianten. ProductCard zeigt eine passende Variant, weitere Varianten bleiben erkennbar. Suchausfall: sichtbarer Hinweis und Retry, Detail/Cart/Checkout lesen PostgreSQL und funktionieren weiter.

## Getrennte Query-Verantwortlichkeiten

Filterplanung validiert Attribute und erzeugt Selektionsprädikate. Facettenplanung bildet pro Facette eigenen Zählkontext. Ranking liefert begrenzte zusätzliche Scores. Sortierung bestimmt eindeutige Ordnung. Vorschläge sind eigener Querypfad. Kleine konkrete Komponenten genügen; Strategy nur bei tatsächlich mehreren Implementierungen, kein monolithischer Universalbuilder.

Textsuche auf deutsch analysiertem Name/Brand/Beschreibung, Name stärker gewichtet als Beschreibung. Innerhalb einer Facette ausgewählte Werte OR, zwischen verschiedenen Facetten AND. Alle Variantenprädikate einschließlich Preis müssen in **einem** nested-Kontext dieselbe SKU erfüllen. Nichtverfügbare Varianten dürfen sichtbar sein, sind mit unavailable markiert; Defaultkarte bevorzugt verfügbare passende Variant, sonst passende nichtverfügbare. Pflichtdatengesperrte Produkte liefern kein kaufbares Angebot.

Trefferselektion als `post_filter` erhält die Aggregationsbasis. Für jede Facette zählen Aggregationen mit allen aktiven Filtern **außer der eigenen Facette**, einschließlich Text/Kategorie. Bei nested Aggregationen Productanzahl über reverse_nested zählen, nicht SKUanzahl; selbst bei mehreren passenden Varianten zählt ein Product einmal. Marken bleiben mit positiver kontextueller Anzahl sichtbar, ausgewählte Nullwerte bleiben als abwählbare Chips. Bereiche bilden dieselbe disjunktive Logik. Kategorie liefert maximal vier Attributfacetten plus Marke/Preis, höchstens drei Topmerkmale je Karte.

## Exaktes Zählbeispiel

Alle Produkte in derselben Smartphonekategorie:

| Product | Brand | Varianten (SKU: Farbe, Speicher, Preis EUR) |
| --- | --- | --- |
| p-a | Alpha | a-red: red, 128 GB, 100; a-blue: blue, 256 GB, 150 |
| p-b | Beta | b-red: red, 256 GB, 120 |
| p-c | Gamma | c-blue: blue, 128 GB, 90 |

Ohne Filter: 3 Produkte, Marke Alpha=1/Beta=1/Gamma=1. Brand Alpha: Treffer p-a, Markenfacette bleibt Alpha=1/Beta=1/Gamma=1. Zusätzlich Beta: Treffer p-a,p-b (OR erweitert). Farbe red + Speicher256: nur p-b, niemals p-a durch Vermischen von a-red/a-blue. Dazu Brand Alpha: 0 Treffer, Markenfacette Alpha=0/Beta=1 (ausgewähltes Alpha bleibt abwählbar); Farb-/Speicherfacetten ignorieren jeweils nur ihren eigenen Filter. Preis≤130 und Speicher256: p-b, nicht p-a. Ein Modell mit zwei passenden SKUs liefert weiterhin eine Karte/Zählung.

## Sortierung, Boosts und Vorschläge

Passende Kartenvariant: bei `priceAsc` günstigste passende, bei `priceDesc` teuerste passende, sonst verfügbare günstigste passende; Gleichstand SKU lexikografisch. Relevanz = Textscore + 0,10×InterestWeight + 0,05×MarginSignal, beide Signale auf [0,1] begrenzt. MarginSignal intern aus vorbereiteter Einkaufskostengrundlage normalisiert, fehlt → 0. Nicht im öffentlichen DTO/HTML/URL. Profilwirkung nur auf Relevanz, keine unpassenden Treffer. Relevanzgleichstand ProductID aufsteigend.

`priceAsc`/`priceDesc` ordnen strikt nach gewähltem passendem Bruttopreis, dann ProductID; Boost darf explizite Ordnung nicht überstimmen. `nameAsc` normalisierter Name dann ProductID. Alle Preise in angefragter unterstützter Währung. Kein stiller Currencymix. Pagination 12 Modelle, stabile Ordnung innerhalb gleicher Indexversion; Liveänderungen können Seiten verschieben, keine Snapshotgarantie.

Autocomplete ab 2 Zeichen, höchstens 5 Vorschläge aus publizierten Produktnamen/Kategorien, Prefix vor Tippfehler, deterministischer Namens-/ID-Gleichstand. Tippfehler höchstens Levenshtein1 bei 4–7 Zeichen, 2 ab8, erste zwei Zeichen unverändert. Vorschläge kennzeichnen „Meinten Sie …“, niemals automatisch Query ersetzen. Leere Suche zeigt Kategorie/Angebote ohne erfundene Treffer; keine Treffer zeigt aktive Chips und Filterzurücksetzen. Indexverzug ist sichtbar über Datenstand, Preise beim Kauf immer maßgeblich neu prüfen.

## Transport für Hauptsuche und Completion

Sobald die bestehende Vorschlagsregel Completion erzeugt, sendet der Adapter genau einen aliasgebundenen `POST /{alias}/_msearch` mit `application/x-ndjson`. Vier JSON-Zeilen mit abschließendem Newline enthalten zweimal Metadaten mit `request_cache: true` und je einen Suchkörper. Ohne Completion bleibt der bisherige `_search`-Pfad bestehen.

Der erste Kontext enthält unverändert Haupttreffer, vollständige gewählte Variantkarten, Sortierung, Mengen und Facetten. Sein Root-Sourcevertrag bleibt ausschließlich `product_id`; die vollständige Karte kommt aus dem gewählten Inner-Hit. Der zweite Kontext enthält nur die bisherige Completion, `size: 0`, `_source: false` und `match_all`. Dessen Trefferzahl und Treffer werden nicht in das fachliche Ergebnis übernommen. Die Publikationsgrenze bleibt für Hauptdaten über `deleted/published` und für Completion über denselben öffentlichen Visibility-Kontext erhalten.

Beide Teile können bei Liveänderungen oder einem Aliaswechsel unterschiedliche Generationen sehen. Ältere Hauptdaten mit neueren Vorschlägen und die umgekehrte Kombination sind erlaubt; Hauptkarten und Facetten bleiben untereinander ein Kontext. Es gibt keine Snapshotgarantie oder atomare Aliasauflösung über beide Teile. Rebuild mit Catchup und atomarem Aliaswechsel sowie CDC bleiben unverändert.

Äußerer Fehler, unlesbare Antwort, andere Antwortanzahl als zwei, fehlender erfolgreicher Teilstatus, ein Teilfehler, Timeout oder fehlgeschlagene Shards führen insgesamt zu `SearchUnavailable` ohne Teilergebnis. Leere Treffer und fehlende Vorschläge bleiben gültig. Der bestehende 250-ms-Kontext gilt einmal für den gesamten HTTPaufruf einschließlich Bodylesen, ohne Retry oder neues Budget je Teil. Synchrones JSON-Dekodieren wird durch diesen Kontext nicht hart unterbrochen.
